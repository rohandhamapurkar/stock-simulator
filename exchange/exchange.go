package exchange

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type Exchange struct {
	IncomingTrades  chan Transaction
	LastTradedPrice TransactionAmtDataType
	BuyQ            TxnBST
	SellQ           TxnBST
	queueLock       sync.Mutex
	// Callbacks for price updates
	priceUpdateCallbacks []func(int)
	callbacksLock        sync.Mutex
}

// NewExchange creates and returns a new exchange with the specified initial Last Traded Price
// If the provided LTP is less than 1, it will be set to 1 (minimum valid price)
func NewExchange(ltp TransactionAmtDataType) Exchange {
	// Ensure the initial LTP is at least 1
	if ltp < 1 {
		ltp = 1
	}

	return Exchange{
		IncomingTrades:       make(chan Transaction),
		LastTradedPrice:      ltp,
		BuyQ:                 TxnBST{},
		SellQ:                TxnBST{},
		priceUpdateCallbacks: make([]func(int), 0),
	}
}

// OrderBookEntry represents an entry in the order book
type OrderBookEntry struct {
	ID     string `json:"id"`
	Price  int    `json:"price"`
	Type   string `json:"type"`
}

// OrderBook represents the current state of the order book
type OrderBook struct {
	BuyOrders  []OrderBookEntry `json:"buyOrders"`
	SellOrders []OrderBookEntry `json:"sellOrders"`
	Timestamp  time.Time        `json:"timestamp"`
}

// RegisterPriceUpdateCallback registers a callback function that will be called when the price changes
func (exch *Exchange) RegisterPriceUpdateCallback(callback func(int)) {
	exch.callbacksLock.Lock()
	defer exch.callbacksLock.Unlock()

	exch.priceUpdateCallbacks = append(exch.priceUpdateCallbacks, callback)
}

// notifyPriceUpdate notifies all registered callbacks about a price update
func (exch *Exchange) notifyPriceUpdate(price int) {
	exch.callbacksLock.Lock()
	defer exch.callbacksLock.Unlock()

	for _, callback := range exch.priceUpdateCallbacks {
		go callback(price)
	}
}

// GetOrderBook returns the current state of the order book
func (exch *Exchange) GetOrderBook() OrderBook {
	exch.queueLock.Lock()
	defer exch.queueLock.Unlock()

	// Get all buy orders
	buyOrders := exch.BuyQ.InorderTraversal()
	buyEntries := make([]OrderBookEntry, 0, len(buyOrders))
	for _, order := range buyOrders {
		buyEntries = append(buyEntries, OrderBookEntry{
			ID:    order.ID,
			Price: int(order.Amount),
			Type:  order.Type,
		})
	}

	// Sort buy orders by price (highest first)
	sort.Slice(buyEntries, func(i, j int) bool {
		return buyEntries[i].Price > buyEntries[j].Price
	})

	// Get all sell orders
	sellOrders := exch.SellQ.InorderTraversal()
	sellEntries := make([]OrderBookEntry, 0, len(sellOrders))
	for _, order := range sellOrders {
		sellEntries = append(sellEntries, OrderBookEntry{
			ID:    order.ID,
			Price: int(order.Amount),
			Type:  order.Type,
		})
	}

	// Sort sell orders by price (lowest first)
	sort.Slice(sellEntries, func(i, j int) bool {
		return sellEntries[i].Price < sellEntries[j].Price
	})

	// Limit to top 10 orders on each side for UI display
	if len(buyEntries) > 10 {
		buyEntries = buyEntries[:10]
	}
	if len(sellEntries) > 10 {
		sellEntries = sellEntries[:10]
	}

	return OrderBook{
		BuyOrders:  buyEntries,
		SellOrders: sellEntries,
		Timestamp:  time.Now(),
	}
}

// AcceptTrades processes incoming trade orders and adds them to the appropriate queue
func (exch *Exchange) AcceptTrades() {
	logger := NewLogger("AcceptTrades")
	logger.Info("Starting to accept trades")

	for txn := range exch.IncomingTrades {
		// Validate transaction price - ensure it's at least 1
		if txn.Amount < 1 {
			logger.Warn(fmt.Sprintf("Rejected order %s with invalid price: %d (minimum price is 1)",
				txn.ID, txn.Amount))
			continue
		}

		exch.queueLock.Lock()
		if txn.Type == BuyTransactionType {
			exch.BuyQ.Insert(txn)
			logger.Debug(fmt.Sprintf("Accepted buy order: %s, price: %d", txn.ID, txn.Amount))
		} else if txn.Type == SellTransactionType {
			exch.SellQ.Insert(txn)
			logger.Debug(fmt.Sprintf("Accepted sell order: %s, price: %d", txn.ID, txn.Amount))
		} else {
			logger.Warn(fmt.Sprintf("Received unknown transaction type: %s", txn.Type))
		}
		exch.queueLock.Unlock()
	}
}

// ProcessTrades periodically processes trades by matching buy and sell orders
func (exch *Exchange) ProcessTrades() {
	ticker := time.NewTicker(time.Second)
	logger := NewLogger("ProcessTrades")

	for {
		<-ticker.C
		logger.Info("Processing trades")

		// Use a timeout for acquiring the lock to prevent deadlocks
		lockAcquired := make(chan bool, 1)
		go func() {
			exch.queueLock.Lock()
			lockAcquired <- true
		}()

		// Wait for lock with timeout
		select {
		case <-lockAcquired:
			// Lock acquired, proceed with processing
		case <-time.After(500 * time.Millisecond):
			logger.Warn("Failed to acquire lock within timeout, skipping this cycle")
			continue
		}

		// Get all buy orders sorted by price (highest first)
		buyOrders := exch.BuyQ.InorderTraversal()
		// Reverse the order to get highest prices first (better for buyers)
		for i, j := 0, len(buyOrders)-1; i < j; i, j = i+1, j-1 {
			buyOrders[i], buyOrders[j] = buyOrders[j], buyOrders[i]
		}

		// Get all sell orders sorted by price (lowest first)
		sellOrders := exch.SellQ.InorderTraversal()

		// Match orders with improved algorithm
		matchedPairs := make([]struct{
			buy  Transaction
			sell Transaction
		}, 0)

		// Find matching pairs
		for _, bTxn := range buyOrders {
			for i, sTxn := range sellOrders {
				// Match if buy price >= sell price (realistic market matching)
				if bTxn.Amount >= sTxn.Amount {
					matchedPairs = append(matchedPairs, struct{
						buy  Transaction
						sell Transaction
					}{bTxn, sTxn})

					// Remove matched sell order from consideration
					sellOrders = append(sellOrders[:i], sellOrders[i+1:]...)
					break
				}
			}
		}

		// Process matched pairs
		for _, pair := range matchedPairs {
			// Use the sell price as the trade price (conservative approach)
			// Ensure the price is never less than 1 (minimum valid price)
			tradePrice := pair.sell.Amount
			if tradePrice < 1 {
				logger.Warn(fmt.Sprintf("Attempted to set LTP to %d, enforcing minimum price of 1", tradePrice))
				tradePrice = 1
			}
			exch.LastTradedPrice = tradePrice

			logger.Info(fmt.Sprintf("Matched buy order %s (price: %d) with sell order %s (price: %d)",
				pair.buy.ID, pair.buy.Amount, pair.sell.ID, pair.sell.Amount))
			logger.Info(fmt.Sprintf("LTP: %d", exch.LastTradedPrice))

			// Notify price update callbacks
			exch.notifyPriceUpdate(int(exch.LastTradedPrice))

			// Remove the matched orders from their respective queues
			exch.BuyQ.Remove(pair.buy)
			exch.SellQ.Remove(pair.sell)
		}

		// Release the lock after processing
		exch.queueLock.Unlock()
	}
}
