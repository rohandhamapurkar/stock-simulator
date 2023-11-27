package exchange

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

/**
 * NewTransaction
 * Returns a instance of a new transaction
 */
func NewTransaction(t string, amount TransactionAmtDataType) Transaction {
	return Transaction{
		Type:   t,
		Amount: amount,
	}
}
