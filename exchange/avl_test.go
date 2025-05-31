package exchange

import (
	"testing"
)

// TestAVLBalancing tests that the AVL tree maintains balance after insertions
func TestAVLBalancing(t *testing.T) {
	// Create a new BST
	bst := TxnBST{}

	// Insert values in ascending order (which would create a right-skewed tree in a regular BST)
	values := []TransactionAmtDataType{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	for _, val := range values {
		txn := NewTransaction(BuyTransactionType, val)
		bst.Insert(txn)
	}

	// Check the height of the tree
	// For 10 nodes, a balanced tree should have a height of around log2(10) ≈ 3-4
	// An unbalanced tree would have a height of 9 (like a linked list)
	treeHeight := height(bst.Root)
	if treeHeight > 4 {
		t.Errorf("Tree is not balanced. Expected height <= 4, got %d", treeHeight)
	}

	// Verify that the tree is still a valid BST by checking inorder traversal
	result := bst.InorderTraversal()
	for i := 1; i < len(result); i++ {
		if result[i-1].Amount > result[i].Amount {
			t.Errorf("Tree is not a valid BST. Values not in ascending order at index %d", i)
		}
	}
}

// TestAVLRotations tests specific rotation cases in the AVL tree
func TestAVLRotations(t *testing.T) {
	// Test Left-Left case (requires right rotation)
	t.Run("Left-Left Case", func(t *testing.T) {
		bst := TxnBST{}
		bst.Insert(NewTransaction(BuyTransactionType, 30))
		bst.Insert(NewTransaction(BuyTransactionType, 20))
		bst.Insert(NewTransaction(BuyTransactionType, 10))

		// After balancing, the root should be 20
		if bst.Root.Value.Amount != 20 {
			t.Errorf("Left-Left rotation failed. Expected root value 20, got %d", bst.Root.Value.Amount)
		}
	})

	// Test Right-Right case (requires left rotation)
	t.Run("Right-Right Case", func(t *testing.T) {
		bst := TxnBST{}
		bst.Insert(NewTransaction(BuyTransactionType, 10))
		bst.Insert(NewTransaction(BuyTransactionType, 20))
		bst.Insert(NewTransaction(BuyTransactionType, 30))

		// After balancing, the root should be 20
		if bst.Root.Value.Amount != 20 {
			t.Errorf("Right-Right rotation failed. Expected root value 20, got %d", bst.Root.Value.Amount)
		}
	})

	// Test Left-Right case (requires left rotation then right rotation)
	t.Run("Left-Right Case", func(t *testing.T) {
		bst := TxnBST{}
		bst.Insert(NewTransaction(BuyTransactionType, 30))
		bst.Insert(NewTransaction(BuyTransactionType, 10))
		bst.Insert(NewTransaction(BuyTransactionType, 20))

		// After balancing, the root should be 20
		if bst.Root.Value.Amount != 20 {
			t.Errorf("Left-Right rotation failed. Expected root value 20, got %d", bst.Root.Value.Amount)
		}
	})

	// Test Right-Left case (requires right rotation then left rotation)
	t.Run("Right-Left Case", func(t *testing.T) {
		bst := TxnBST{}
		bst.Insert(NewTransaction(BuyTransactionType, 10))
		bst.Insert(NewTransaction(BuyTransactionType, 30))
		bst.Insert(NewTransaction(BuyTransactionType, 20))

		// After balancing, the root should be 20
		if bst.Root.Value.Amount != 20 {
			t.Errorf("Right-Left rotation failed. Expected root value 20, got %d", bst.Root.Value.Amount)
		}
	})
}

// TestAVLRemoval tests that the AVL tree maintains balance after removals
func TestAVLRemoval(t *testing.T) {
	// Create a balanced tree
	bst := TxnBST{}
	values := []TransactionAmtDataType{50, 30, 70, 20, 40, 60, 80}
	txns := make([]Transaction, len(values))

	for i, val := range values {
		txn := NewTransaction(BuyTransactionType, val)
		txns[i] = txn
		bst.Insert(txn)
	}

	// Remove nodes and check balance
	for _, txn := range txns {
		// Make a copy of the transaction to remove
		bst.Remove(txn)

		// Check that the tree is still balanced
		if bst.Root != nil {
			balance := bst.Root.balanceFactor()
			if balance < -1 || balance > 1 {
				t.Errorf("Tree became unbalanced after removing %d. Balance factor: %d", 
					txn.Amount, balance)
			}
		}
	}

	// Tree should be empty now
	if bst.Root != nil {
		t.Errorf("Tree is not empty after removing all nodes")
	}
}

// TestAVLLargeDataset tests the AVL tree with a larger dataset
func TestAVLLargeDataset(t *testing.T) {
	// Create a new BST
	bst := TxnBST{}

	// Insert 1000 values in ascending order
	const numNodes = 1000
	for i := 0; i < numNodes; i++ {
		txn := NewTransaction(BuyTransactionType, TransactionAmtDataType(i))
		bst.Insert(txn)
	}

	// Check the height of the tree
	// For 1000 nodes, a balanced tree should have a height of around log2(1000) ≈ 10
	// An unbalanced tree would have a height of 999
	treeHeight := height(bst.Root)
	maxExpectedHeight := 15 // Being generous with the height limit
	if treeHeight > maxExpectedHeight {
		t.Errorf("Tree is not balanced with large dataset. Expected height <= %d, got %d", 
			maxExpectedHeight, treeHeight)
	}

	// Verify all values can be found
	for i := 0; i < numNodes; i++ {
		result := bst.Search(TransactionAmtDataType(i))
		if result == nil {
			t.Errorf("Value %d not found in tree", i)
		}
	}

	// Remove all values and check the tree remains balanced
	for i := 0; i < numNodes; i++ {
		// Find the actual transaction to remove
		foundTxn := bst.Search(TransactionAmtDataType(i))
		if foundTxn != nil {
			bst.Remove(*foundTxn)
		}

		// Check that the tree is still balanced (if not empty)
		if bst.Root != nil {
			balance := bst.Root.balanceFactor()
			if balance < -1 || balance > 1 {
				t.Errorf("Tree became unbalanced after removing %d. Balance factor: %d", 
					i, balance)
				break // Stop after first failure to avoid too many errors
			}
		}
	}
}
