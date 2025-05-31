package exchange

import (
	"sync"
	"testing"
	"time"
)

func TestNewExchange(t *testing.T) {
	// Test cases for exchange creation
	testCases := []struct {
		name           string
		initialLTP     TransactionAmtDataType
		expectedLTP    TransactionAmtDataType
	}{
		{
			name:           "Positive Initial LTP",
			initialLTP:     100,
			expectedLTP:    100,
		},
		{
			name:           "Zero Initial LTP",
			initialLTP:     0,
			expectedLTP:    1, // Should be set to minimum of 1
		},
		{
			name:           "Negative Initial LTP",
			initialLTP:     -10,
			expectedLTP:    1, // Should be set to minimum of 1
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exchange := NewExchange(tc.initialLTP)

			// Check initial LTP
			if exchange.LastTradedPrice != tc.expectedLTP {
				t.Errorf("Expected initial LTP %d, got %d", 
					tc.expectedLTP, exchange.LastTradedPrice)
			}

			// Check that channels are initialized
			if exchange.IncomingTrades == nil {
				t.Errorf("IncomingTrades channel not initialized")
			}

			// Check that callback slice is initialized
			if exchange.priceUpdateCallbacks == nil {
				t.Errorf("priceUpdateCallbacks slice not initialized")
			}
		})
	}
}

func TestRegisterPriceUpdateCallback(t *testing.T) {
	exchange := NewExchange(100)
	
	// Create a channel to verify callback execution
	callbackExecuted := make(chan int, 1)
	
	// Register a callback
	exchange.RegisterPriceUpdateCallback(func(price int) {
		callbackExecuted <- price
	})
	
	// Trigger the callback by calling notifyPriceUpdate
	testPrice := 150
	exchange.notifyPriceUpdate(testPrice)
	
	// Wait for callback to execute with a timeout
	select {
	case price := <-callbackExecuted:
		if price != testPrice {
			t.Errorf("Expected callback to receive price %d, got %d", testPrice, price)
		}
	case <-time.After(time.Second):
		t.Errorf("Callback was not executed within timeout")
	}
}

func TestGetOrderBook(t *testing.T) {
	exchange := NewExchange(100)
	
	// Add some buy orders
	buyOrders := []TransactionAmtDataType{90, 95, 85, 80, 75}
	for _, price := range buyOrders {
		txn := NewTransaction(BuyTransactionType, price)
		exchange.BuyQ.Insert(txn)
	}
	
	// Add some sell orders
	sellOrders := []TransactionAmtDataType{110, 105, 115, 120, 125}
	for _, price := range sellOrders {
		txn := NewTransaction(SellTransactionType, price)
		exchange.SellQ.Insert(txn)
	}
	
	// Get the order book
	orderBook := exchange.GetOrderBook()
	
	// Check buy orders (should be sorted highest first)
	if len(orderBook.BuyOrders) != 5 {
		t.Errorf("Expected 5 buy orders, got %d", len(orderBook.BuyOrders))
	} else {
		// Check sorting (highest first)
		for i := 0; i < len(orderBook.BuyOrders)-1; i++ {
			if orderBook.BuyOrders[i].Price < orderBook.BuyOrders[i+1].Price {
				t.Errorf("Buy orders not sorted correctly at index %d: %d < %d", 
					i, orderBook.BuyOrders[i].Price, orderBook.BuyOrders[i+1].Price)
			}
		}
	}
	
	// Check sell orders (should be sorted lowest first)
	if len(orderBook.SellOrders) != 5 {
		t.Errorf("Expected 5 sell orders, got %d", len(orderBook.SellOrders))
	} else {
		// Check sorting (lowest first)
		for i := 0; i < len(orderBook.SellOrders)-1; i++ {
			if orderBook.SellOrders[i].Price > orderBook.SellOrders[i+1].Price {
				t.Errorf("Sell orders not sorted correctly at index %d: %d > %d", 
					i, orderBook.SellOrders[i].Price, orderBook.SellOrders[i+1].Price)
			}
		}
	}
}

func TestAcceptTrades(t *testing.T) {
	exchange := NewExchange(100)
	
	// Start the AcceptTrades goroutine
	go exchange.AcceptTrades()
	
	// Create test transactions
	buyTxn := NewTransaction(BuyTransactionType, 90)
	sellTxn := NewTransaction(SellTransactionType, 110)
	invalidTxn := NewTransaction(BuyTransactionType, 0) // Invalid price
	
	// Send transactions to the exchange
	exchange.IncomingTrades <- buyTxn
	exchange.IncomingTrades <- sellTxn
	exchange.IncomingTrades <- invalidTxn
	
	// Give some time for processing
	time.Sleep(100 * time.Millisecond)
	
	// Check that valid transactions were added to the queues
	buyOrders := exchange.BuyQ.InorderTraversal()
	sellOrders := exchange.SellQ.InorderTraversal()
	
	if len(buyOrders) != 1 {
		t.Errorf("Expected 1 buy order, got %d", len(buyOrders))
	}
	
	if len(sellOrders) != 1 {
		t.Errorf("Expected 1 sell order, got %d", len(sellOrders))
	}
	
	if len(buyOrders) > 0 && buyOrders[0].Amount != 90 {
		t.Errorf("Expected buy order with price 90, got %d", buyOrders[0].Amount)
	}
	
	if len(sellOrders) > 0 && sellOrders[0].Amount != 110 {
		t.Errorf("Expected sell order with price 110, got %d", sellOrders[0].Amount)
	}
}

func TestProcessTrades(t *testing.T) {
	// Create a test exchange with initial LTP of 100
	exchange := NewExchange(100)
	
	// Create a channel to track price updates
	priceUpdates := make(chan int, 10)
	exchange.RegisterPriceUpdateCallback(func(price int) {
		priceUpdates <- price
	})
	
	// Start the ProcessTrades goroutine
	go exchange.ProcessTrades()
	
	// Add buy and sell orders that should match
	buyTxn := NewTransaction(BuyTransactionType, 110) // Willing to buy at 110
	sellTxn := NewTransaction(SellTransactionType, 90) // Willing to sell at 90
	
	exchange.BuyQ.Insert(buyTxn)
	exchange.SellQ.Insert(sellTxn)
	
	// Wait for processing to occur (the ticker in ProcessTrades is 1 second)
	select {
	case price := <-priceUpdates:
		// The trade should execute at the sell price (90)
		if price != 90 {
			t.Errorf("Expected trade to execute at price 90, got %d", price)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("No price update received within timeout")
	}
	
	// Verify that the orders were removed from the queues
	buyOrders := exchange.BuyQ.InorderTraversal()
	sellOrders := exchange.SellQ.InorderTraversal()
	
	if len(buyOrders) != 0 {
		t.Errorf("Expected buy queue to be empty after matching, found %d orders", len(buyOrders))
	}
	
	if len(sellOrders) != 0 {
		t.Errorf("Expected sell queue to be empty after matching, found %d orders", len(sellOrders))
	}
	
	// Verify that the LTP was updated
	if exchange.LastTradedPrice != 90 {
		t.Errorf("Expected LTP to be updated to 90, got %d", exchange.LastTradedPrice)
	}
}

func TestConcurrentOrderProcessing(t *testing.T) {
	// This test verifies that the exchange can handle concurrent order submission
	exchange := NewExchange(100)
	
	// Start the order processing goroutines
	go exchange.AcceptTrades()
	go exchange.ProcessTrades()
	
	// Track matched orders through price updates
	var priceUpdateCount int
	var priceMutex sync.Mutex
	
	exchange.RegisterPriceUpdateCallback(func(price int) {
		priceMutex.Lock()
		priceUpdateCount++
		priceMutex.Unlock()
	})
	
	// Generate a bunch of matching orders concurrently
	const numOrders = 20
	var wg sync.WaitGroup
	wg.Add(numOrders)
	
	for i := 0; i < numOrders; i++ {
		go func(i int) {
			defer wg.Done()
			
			// Create matching buy and sell orders
			buyPrice := TransactionAmtDataType(100 + i)
			sellPrice := TransactionAmtDataType(100 - i)
			
			buyTxn := NewTransaction(BuyTransactionType, buyPrice)
			sellTxn := NewTransaction(SellTransactionType, sellPrice)
			
			// Submit the orders
			exchange.IncomingTrades <- buyTxn
			exchange.IncomingTrades <- sellTxn
		}(i)
	}
	
	// Wait for all orders to be submitted
	wg.Wait()
	
	// Wait for processing to occur (give it some time)
	time.Sleep(2 * time.Second)
	
	// Check that some price updates occurred (indicating matches)
	priceMutex.Lock()
	updates := priceUpdateCount
	priceMutex.Unlock()
	
	if updates == 0 {
		t.Errorf("Expected some price updates from matching orders, got none")
	}
	
	// Check that the order books are eventually processed
	buyOrders := exchange.BuyQ.InorderTraversal()
	sellOrders := exchange.SellQ.InorderTraversal()
	
	// We can't guarantee all orders will be matched due to timing,
	// but we should see some reduction in the order books
	if len(buyOrders) == numOrders || len(sellOrders) == numOrders {
		t.Errorf("Expected some orders to be matched and removed, but found %d buy orders and %d sell orders",
			len(buyOrders), len(sellOrders))
	}
}
