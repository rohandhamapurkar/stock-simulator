package exchange

import (
	"sync"
	"testing"
	"time"
)

func TestConcurrentTxnBST(t *testing.T) {
	// Create a new concurrent BST
	bst := NewConcurrentTxnBST()

	// Test basic operations
	t.Run("Basic Operations", func(t *testing.T) {
		// Insert some values
		bst.Insert(NewTransaction(BuyTransactionType, 100))
		bst.Insert(NewTransaction(BuyTransactionType, 50))
		bst.Insert(NewTransaction(BuyTransactionType, 150))

		// Check that the values were inserted
		result := bst.InorderTraversal()
		if len(result) != 3 {
			t.Errorf("Expected 3 nodes in the BST, got %d", len(result))
		}

		// Check that the values are in the correct order
		if result[0].Amount != 50 || result[1].Amount != 100 || result[2].Amount != 150 {
			t.Errorf("BST inorder traversal incorrect. Expected [50, 100, 150], got [%d, %d, %d]",
				result[0].Amount, result[1].Amount, result[2].Amount)
		}

		// Search for a value
		txn := bst.Search(100)
		if txn == nil {
			t.Errorf("Expected to find transaction with amount 100, got nil")
		}
		if txn != nil && txn.Amount != 100 {
			t.Errorf("Expected transaction with amount 100, got %d", txn.Amount)
		}

		// Remove a value
		bst.Remove(result[1]) // Remove the transaction with amount 100
		result = bst.InorderTraversal()
		if len(result) != 2 {
			t.Errorf("Expected 2 nodes in the BST after removal, got %d", len(result))
		}
		if result[0].Amount != 50 || result[1].Amount != 150 {
			t.Errorf("BST inorder traversal incorrect after removal. Expected [50, 150], got [%d, %d]",
				result[0].Amount, result[1].Amount)
		}
	})

	// Test concurrent operations
	t.Run("Concurrent Operations", func(t *testing.T) {
		// Create a new concurrent BST
		bst := NewConcurrentTxnBST()

		// Number of goroutines and operations
		const numGoroutines = 10
		const numOperations = 100

		// Use a wait group to wait for all goroutines to finish
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Start goroutines to perform concurrent operations
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()

				// Perform a mix of operations
				for j := 0; j < numOperations; j++ {
					// Create a unique value for this goroutine and operation
					value := TransactionAmtDataType(id*numOperations + j)
					txn := NewTransaction(BuyTransactionType, value)

					// Insert the value
					bst.Insert(txn)

					// Search for the value
					found := bst.Search(value)
					if found == nil {
						t.Errorf("Expected to find transaction with amount %d, got nil", value)
					}

					// Remove the value
					bst.Remove(txn)

					// Verify it was removed
					found = bst.Search(value)
					if found != nil {
						t.Errorf("Expected not to find transaction with amount %d after removal", value)
					}
				}
			}(i)
		}

		// Wait for all goroutines to finish
		wg.Wait()

		// Check memory stats
		allocated, recycled := bst.GetStats()
		t.Logf("Memory stats: allocated=%d, recycled=%d", allocated, recycled)

		// Verify that some nodes were recycled
		if recycled == 0 {
			t.Errorf("Expected some nodes to be recycled, got 0")
		}
	})

	// Test node recycling
	t.Run("Node Recycling", func(t *testing.T) {
		// Create a new concurrent BST
		bst := NewConcurrentTxnBST()

		// Insert and remove a large number of nodes
		const numNodes = 1000
		for i := 0; i < numNodes; i++ {
			txn := NewTransaction(BuyTransactionType, TransactionAmtDataType(i))
			bst.Insert(txn)
		}

		// Get the initial stats
		allocatedBefore, recycledBefore := bst.GetStats()

		// Remove all nodes
		result := bst.InorderTraversal()
		for _, txn := range result {
			bst.Remove(txn)
		}

		// Get the final stats
		allocatedAfter, recycledAfter := bst.GetStats()

		// Verify that nodes were recycled
		if recycledAfter <= recycledBefore {
			t.Errorf("Expected recycled count to increase, got %d -> %d", recycledBefore, recycledAfter)
		}

		// Verify that allocated count didn't increase significantly
		if allocatedAfter > allocatedBefore+10 {
			t.Errorf("Expected allocated count to remain stable, got %d -> %d", allocatedBefore, allocatedAfter)
		}

		// Log the stats
		t.Logf("Memory stats before: allocated=%d, recycled=%d", allocatedBefore, recycledBefore)
		t.Logf("Memory stats after: allocated=%d, recycled=%d", allocatedAfter, recycledAfter)
		t.Logf("Memory saved: %d nodes", recycledAfter-recycledBefore)
	})

	// Test high concurrency with read-heavy workload
	t.Run("High Concurrency Read-Heavy", func(t *testing.T) {
		// Create a new concurrent BST
		bst := NewConcurrentTxnBST()

		// Insert some initial data
		const initialNodes = 100
		for i := 0; i < initialNodes; i++ {
			txn := NewTransaction(BuyTransactionType, TransactionAmtDataType(i))
			bst.Insert(txn)
		}

		// Number of reader and writer goroutines
		const numReaders = 20
		const numWriters = 2
		const numOperations = 1000

		// Use a wait group to wait for all goroutines to finish
		var wg sync.WaitGroup
		wg.Add(numReaders + numWriters)

		// Start reader goroutines
		for i := 0; i < numReaders; i++ {
			go func() {
				defer wg.Done()

				// Perform read operations
				for j := 0; j < numOperations; j++ {
					// Get all values
					_ = bst.InorderTraversal()

					// Search for a random value
					value := TransactionAmtDataType(j % initialNodes)
					_ = bst.Search(value)

					// Small sleep to simulate processing
					time.Sleep(time.Microsecond)
				}
			}()
		}

		// Start writer goroutines
		for i := 0; i < numWriters; i++ {
			go func(id int) {
				defer wg.Done()

				// Perform write operations
				for j := 0; j < numOperations; j++ {
					// Create a unique value for this goroutine and operation
					value := TransactionAmtDataType(initialNodes + id*numOperations + j)
					txn := NewTransaction(BuyTransactionType, value)

					// Insert the value
					bst.Insert(txn)

					// Small sleep to simulate processing
					time.Sleep(time.Millisecond)

					// Remove the value
					bst.Remove(txn)
				}
			}(i)
		}

		// Wait for all goroutines to finish
		wg.Wait()

		// Verify the final state
		result := bst.InorderTraversal()
		if len(result) != initialNodes {
			t.Errorf("Expected %d nodes in the BST, got %d", initialNodes, len(result))
		}

		// Check memory stats
		allocated, recycled := bst.GetStats()
		t.Logf("Memory stats: allocated=%d, recycled=%d", allocated, recycled)
	})
}
