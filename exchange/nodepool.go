package exchange

import (
	"sync"
	"sync/atomic"
)

// NodePool manages a pool of tree nodes to reduce memory allocations
type NodePool struct {
	pool      sync.Pool
	allocated int64
	recycled  int64
}

// NewNodePool creates a new node pool
func NewNodePool() *NodePool {
	np := &NodePool{}
	np.pool = sync.Pool{
		New: func() interface{} {
			atomic.AddInt64(&np.allocated, 1)
			return &treeNode{}
		},
	}
	return np
}

// Get retrieves a node from the pool or creates a new one if the pool is empty
func (np *NodePool) Get() *treeNode {
	return np.pool.Get().(*treeNode)
}

// Put returns a node to the pool for reuse
func (np *NodePool) Put(node *treeNode) {
	if node == nil {
		return
	}
	
	// Reset node state
	node.Left = nil
	node.Right = nil
	node.Height = 0
	// We don't reset Value as it will be overwritten when the node is reused
	
	// Return to pool
	atomic.AddInt64(&np.recycled, 1)
	np.pool.Put(node)
}

// Stats returns statistics about the node pool
func (np *NodePool) Stats() (allocated, recycled int64) {
	return atomic.LoadInt64(&np.allocated), atomic.LoadInt64(&np.recycled)
}
