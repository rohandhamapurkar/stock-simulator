package exchange

import (
	"fmt"
	"time"
)

type Transaction struct {
	ID     string
	Type   string
	Amount TransactionAmtDataType
}

type TransactionAmtDataType int32

const (
	BuyTransactionType  = "BUY"
	SellTransactionType = "SELL"
)

// generateID creates a unique ID for a transaction based on timestamp and type
func generateID(txnType string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d", txnType, timestamp)
}

/**
 * NewTransaction
 * Returns an instance of a new transaction with a unique ID
 */
func NewTransaction(t string, amount TransactionAmtDataType) Transaction {
	return Transaction{
		ID:     generateID(t),
		Type:   t,
		Amount: amount,
	}
}
