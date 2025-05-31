package exchange

import (
	"sync"
	"testing"
)

func TestNodePool(t *testing.T) {
	// Create a new node pool
	pool := NewNodePool()

	// Test basic operations
	t.Run("Basic Operations", func(t *testing.T) {
		// Get a node from the pool
		node := pool.Get()
		if node == nil {
			t.Errorf("Expected non-nil node from pool")
		}

		// Check initial stats
		allocated, recycled := pool.Stats()
		if allocated != 1 {
			t.Errorf("Expected 1 allocated node, got %d", allocated)
		}
		if recycled != 0 {
			t.Errorf("Expected 0 recycled nodes, got %d", recycled)
		}

		// Return the node to the pool
		pool.Put(node)

		// Check updated stats
		allocated, recycled = pool.Stats()
		if allocated != 1 {
			t.Errorf("Expected 1 allocated node, got %d", allocated)
		}
		if recycled != 1 {
			t.Errorf("Expected 1 recycled node, got %d", recycled)
		}

		// Get another node from the pool (should reuse the one we put back)
		node2 := pool.Get()
		if node2 == nil {
			t.Errorf("Expected non-nil node from pool")
		}

		// Check that allocated count didn't increase
		allocated, recycled = pool.Stats()
		if allocated != 1 {
			t.Errorf("Expected 1 allocated node, got %d", allocated)
		}
		// Recycled count should remain the same since we took a node out
		if recycled != 1 {
			t.Errorf("Expected 1 recycled node, got %d", recycled)
		}
	})

	// Test concurrent operations
	t.Run("Concurrent Operations", func(t *testing.T) {
		// Create a new node pool
		pool := NewNodePool()

		// Number of goroutines and operations
		const numGoroutines = 10
		const numOperations = 100

		// Use a wait group to wait for all goroutines to finish
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Start goroutines to perform concurrent operations
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()

				// Create a slice to hold nodes temporarily
				nodes := make([]*treeNode, 0, numOperations)

				// Get nodes from the pool
				for j := 0; j < numOperations; j++ {
					node := pool.Get()
					if node == nil {
						t.Errorf("Expected non-nil node from pool")
					}
					nodes = append(nodes, node)
				}

				// Return nodes to the pool
				for _, node := range nodes {
					pool.Put(node)
				}
			}()
		}

		// Wait for all goroutines to finish
		wg.Wait()

		// Check final stats
		allocated, recycled := pool.Stats()
		t.Logf("Memory stats: allocated=%d, recycled=%d", allocated, recycled)

		// Verify that the number of allocated nodes is reasonable
		// In the worst case, each goroutine might allocate all its nodes before any are recycled
		maxExpectedAllocated := numGoroutines * numOperations
		if allocated > int64(maxExpectedAllocated) {
			t.Errorf("Expected at most %d allocated nodes, got %d", maxExpectedAllocated, allocated)
		}

		// Verify that nodes were recycled
		if recycled == 0 {
			t.Errorf("Expected some nodes to be recycled, got 0")
		}

		// Verify that recycling was effective
		recyclingEfficiency := float64(recycled) / float64(allocated) * 100
		t.Logf("Recycling efficiency: %.2f%%", recyclingEfficiency)
		if recyclingEfficiency < 50 {
			t.Errorf("Expected recycling efficiency to be at least 50%%, got %.2f%%", recyclingEfficiency)
		}
	})

	// Test node reset
	t.Run("Node Reset", func(t *testing.T) {
		// Create a new node pool
		pool := NewNodePool()

		// Get a node and set its fields
		node := pool.Get()
		node.Value = NewTransaction(BuyTransactionType, 100)
		node.Left = &treeNode{}
		node.Right = &treeNode{}
		node.Height = 5

		// Return the node to the pool
		pool.Put(node)

		// Get the node back from the pool
		node2 := pool.Get()

		// Check that the fields were reset
		if node2.Left != nil {
			t.Errorf("Expected Left to be nil after reset")
		}
		if node2.Right != nil {
			t.Errorf("Expected Right to be nil after reset")
		}
		if node2.Height != 0 {
			t.Errorf("Expected Height to be 0 after reset, got %d", node2.Height)
		}
		// Value is not reset as it will be overwritten when the node is reused
	})
}
