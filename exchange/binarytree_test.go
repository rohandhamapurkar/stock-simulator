package exchange

import (
	"testing"
)

func TestBSTInsert(t *testing.T) {
	// Create a new BST
	bst := TxnBST{}

	// Create test transactions
	txn1 := NewTransaction(BuyTransactionType, 100)
	txn2 := NewTransaction(BuyTransactionType, 50)
	txn3 := NewTransaction(BuyTransactionType, 150)

	// Insert transactions into the BST
	bst.Insert(txn1)
	bst.Insert(txn2)
	bst.Insert(txn3)

	// Verify the BST structure using inorder traversal
	// Inorder traversal of a BST should give sorted order
	result := bst.InorderTraversal()

	// Check if we have the correct number of nodes
	if len(result) != 3 {
		t.Errorf("Expected 3 nodes in the BST, got %d", len(result))
	}

	// Check if the nodes are in the correct order
	if result[0].Amount != 50 || result[1].Amount != 100 || result[2].Amount != 150 {
		t.Errorf("BST inorder traversal incorrect. Expected [50, 100, 150], got [%d, %d, %d]",
			result[0].Amount, result[1].Amount, result[2].Amount)
	}
}

func TestBSTSearch(t *testing.T) {
	// Create a new BST
	bst := TxnBST{}

	// Create test transactions
	txn1 := NewTransaction(BuyTransactionType, 100)
	txn2 := NewTransaction(BuyTransactionType, 50)
	txn3 := NewTransaction(BuyTransactionType, 150)

	// Insert transactions into the BST
	bst.Insert(txn1)
	bst.Insert(txn2)
	bst.Insert(txn3)

	// Test cases
	testCases := []struct {
		name          string
		searchValue   TransactionAmtDataType
		expectedFound bool
	}{
		{"Find existing value (50)", 50, true},
		{"Find existing value (100)", 100, true},
		{"Find existing value (150)", 150, true},
		{"Find non-existing value (75)", 75, false},
		{"Find non-existing value (200)", 200, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := bst.Search(tc.searchValue)
			found := result != nil

			if found != tc.expectedFound {
				t.Errorf("Expected found=%v for value %d, got found=%v",
					tc.expectedFound, tc.searchValue, found)
			}

			if found && result.Amount != tc.searchValue {
				t.Errorf("Expected value %d, got %d", tc.searchValue, result.Amount)
			}
		})
	}
}

func TestBSTRemove(t *testing.T) {
	// Test cases for different removal scenarios
	testCases := []struct {
		name           string
		insertValues   []TransactionAmtDataType
		removeValue    TransactionAmtDataType
		expectedValues []TransactionAmtDataType
	}{
		{
			name:           "Remove leaf node",
			insertValues:   []TransactionAmtDataType{100, 50, 150, 25, 75},
			removeValue:    25,
			expectedValues: []TransactionAmtDataType{50, 75, 100, 150},
		},
		{
			name:           "Remove node with one child",
			insertValues:   []TransactionAmtDataType{100, 50, 150, 25},
			removeValue:    50,
			expectedValues: []TransactionAmtDataType{25, 100, 150},
		},
		{
			name:           "Remove node with two children",
			insertValues:   []TransactionAmtDataType{100, 50, 150, 25, 75, 125, 175},
			removeValue:    100,
			expectedValues: []TransactionAmtDataType{25, 50, 75, 125, 150, 175},
		},
		{
			name:           "Remove root node",
			insertValues:   []TransactionAmtDataType{100},
			removeValue:    100,
			expectedValues: []TransactionAmtDataType{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new BST
			bst := TxnBST{}

			// Insert values
			for _, val := range tc.insertValues {
				txn := NewTransaction(BuyTransactionType, val)
				bst.Insert(txn)
			}

			// Create transaction to remove
			txnToRemove := NewTransaction(BuyTransactionType, tc.removeValue)
			
			// Find the actual transaction to remove (since IDs will be different)
			foundTxn := bst.Search(tc.removeValue)
			if foundTxn != nil {
				txnToRemove = *foundTxn
			}

			// Remove the transaction
			bst.Remove(txnToRemove)

			// Check the result
			resultList := bst.InorderTraversal()
			
			// Verify the number of nodes
			if len(resultList) != len(tc.expectedValues) {
				t.Errorf("Expected %d nodes after removal, got %d", 
					len(tc.expectedValues), len(resultList))
				return
			}

			// Verify the values
			for i, val := range tc.expectedValues {
				if resultList[i].Amount != val {
					t.Errorf("Expected value at index %d to be %d, got %d", 
						i, val, resultList[i].Amount)
				}
			}
		})
	}
}

func TestBSTEmptyTree(t *testing.T) {
	// Create an empty BST
	bst := TxnBST{}

	// Test inorder traversal on empty tree
	result := bst.InorderTraversal()
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty tree, got %d elements", len(result))
	}

	// Test search on empty tree
	if bst.Search(100) != nil {
		t.Errorf("Expected nil result when searching empty tree")
	}

	// Test remove on empty tree (should not panic)
	txn := NewTransaction(BuyTransactionType, 100)
	bst.Remove(txn) // This should not cause a panic
}

func TestBSTDuplicateValues(t *testing.T) {
	// Create a new BST
	bst := TxnBST{}

	// Insert transactions with duplicate values
	txn1 := NewTransaction(BuyTransactionType, 100)
	txn2 := NewTransaction(BuyTransactionType, 100) // Same value as txn1
	txn3 := NewTransaction(BuyTransactionType, 100) // Same value as txn1 and txn2

	bst.Insert(txn1)
	bst.Insert(txn2)
	bst.Insert(txn3)

	// Check that all transactions were inserted
	result := bst.InorderTraversal()
	if len(result) != 3 {
		t.Errorf("Expected 3 nodes with duplicate values, got %d", len(result))
	}

	// All values should be 100
	for i, txn := range result {
		if txn.Amount != 100 {
			t.Errorf("Expected value 100 at index %d, got %d", i, txn.Amount)
		}
	}
}
