package exchange

import (
	"strings"
	"testing"
	"time"
)

func TestNewTransaction(t *testing.T) {
	// Test cases for transaction creation
	testCases := []struct {
		name          string
		txnType       string
		amount        TransactionAmtDataType
		expectedType  string
		expectedAmount TransactionAmtDataType
	}{
		{
			name:          "Buy Transaction",
			txnType:       BuyTransactionType,
			amount:        100,
			expectedType:  BuyTransactionType,
			expectedAmount: 100,
		},
		{
			name:          "Sell Transaction",
			txnType:       SellTransactionType,
			amount:        150,
			expectedType:  SellTransactionType,
			expectedAmount: 150,
		},
		{
			name:          "Zero Amount Transaction",
			txnType:       BuyTransactionType,
			amount:        0,
			expectedType:  BuyTransactionType,
			expectedAmount: 0,
		},
		{
			name:          "Negative Amount Transaction",
			txnType:       SellTransactionType,
			amount:        -10,
			expectedType:  SellTransactionType,
			expectedAmount: -10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			txn := NewTransaction(tc.txnType, tc.amount)

			// Check transaction type
			if txn.Type != tc.expectedType {
				t.Errorf("Expected transaction type %s, got %s", tc.expectedType, txn.Type)
			}

			// Check transaction amount
			if txn.Amount != tc.expectedAmount {
				t.Errorf("Expected transaction amount %d, got %d", tc.expectedAmount, txn.Amount)
			}

			// Check that ID is not empty
			if txn.ID == "" {
				t.Errorf("Expected non-empty transaction ID")
			}

			// Check that ID contains the transaction type
			if !strings.Contains(txn.ID, tc.txnType) {
				t.Errorf("Expected transaction ID to contain type %s, got %s", tc.txnType, txn.ID)
			}
		})
	}
}

func TestTransactionIDUniqueness(t *testing.T) {
	// Create multiple transactions in quick succession
	const numTransactions = 100
	ids := make(map[string]bool)

	for i := 0; i < numTransactions; i++ {
		// Alternate between buy and sell transactions
		txnType := BuyTransactionType
		if i%2 == 1 {
			txnType = SellTransactionType
		}

		txn := NewTransaction(txnType, TransactionAmtDataType(i))
		
		// Check if this ID has been seen before
		if ids[txn.ID] {
			t.Errorf("Duplicate transaction ID found: %s", txn.ID)
		}
		
		ids[txn.ID] = true
		
		// Small sleep to ensure we don't generate IDs too quickly
		time.Sleep(time.Nanosecond)
	}

	// Verify we have the expected number of unique IDs
	if len(ids) != numTransactions {
		t.Errorf("Expected %d unique transaction IDs, got %d", numTransactions, len(ids))
	}
}

func TestTransactionIDFormat(t *testing.T) {
	// Test that transaction IDs follow the expected format
	
	// Create a buy transaction
	buyTxn := NewTransaction(BuyTransactionType, 100)
	
	// Check format: should be "BUY-timestamp"
	parts := strings.Split(buyTxn.ID, "-")
	if len(parts) != 2 {
		t.Errorf("Expected transaction ID format 'TYPE-TIMESTAMP', got %s", buyTxn.ID)
	}
	
	if parts[0] != BuyTransactionType {
		t.Errorf("Expected transaction ID to start with %s, got %s", BuyTransactionType, parts[0])
	}
	
	// Check that the timestamp part is a number
	_, err := time.ParseDuration(parts[1] + "ns")
	if err != nil {
		t.Errorf("Expected transaction ID timestamp to be a valid number, got %s", parts[1])
	}
	
	// Create a sell transaction
	sellTxn := NewTransaction(SellTransactionType, 100)
	
	// Check format: should be "SELL-timestamp"
	parts = strings.Split(sellTxn.ID, "-")
	if len(parts) != 2 {
		t.Errorf("Expected transaction ID format 'TYPE-TIMESTAMP', got %s", sellTxn.ID)
	}
	
	if parts[0] != SellTransactionType {
		t.Errorf("Expected transaction ID to start with %s, got %s", SellTransactionType, parts[0])
	}
}
