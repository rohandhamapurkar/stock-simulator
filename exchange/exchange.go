package exchange

import (
	"fmt"
	"sync"
	"time"
)

type Exchange struct {
	IncomingTrades  chan Transaction
	LastTradedPrice TransactionAmtDataType
	BuyQ            TxnBST
	SellQ           TxnBST
	queueLock       sync.Mutex
}

/**
 * NewExchange
 * Returns a new exchange
 */
func NewExchange(ltp TransactionAmtDataType) Exchange {
	return Exchange{
		IncomingTrades:  make(chan Transaction),
		LastTradedPrice: ltp,
		BuyQ:            TxnBST{},
		SellQ:           TxnBST{},
	}
}

func (exch *Exchange) AcceptTrades() {
	for txn := range exch.IncomingTrades {
		exch.queueLock.Lock()
		if txn.Type == BuyTransactionType {
			exch.BuyQ.Insert(txn)
		} else {
			exch.SellQ.Insert(txn)
		}
		exch.queueLock.Unlock()
	}
}

func (exch *Exchange) ProcessTrades() {
	ticker := time.NewTicker(time.Second)

	for {
		<-ticker.C
		fmt.Println("Processing trades")
		gotIt := exch.queueLock.TryLock()
		if !gotIt {
			continue
		}
		for _, bTxn := range exch.BuyQ.InorderTraversal() {
			sTxn := exch.SellQ.Search(bTxn.Amount)
			if sTxn == nil {
				continue
			}
			exch.LastTradedPrice = bTxn.Amount
			// fmt.Printf("Fulfilled buy: %s, amt: %d\n", bTxn.ID, bTxn.Amount)
			// fmt.Printf("Fulfilled sell: %s, amt: %d\n", sTxn.ID, sTxn.Amount)
			fmt.Printf("LTP: %d\n", exch.LastTradedPrice)
			exch.BuyQ.Remove(bTxn)
			exch.SellQ.Remove(*sTxn)
		}
		exch.queueLock.Unlock()
	}

}
