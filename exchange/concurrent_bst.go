package exchange

import (
	"sync"
)

// ConcurrentTxnBST is a thread-safe wrapper around TxnBST with memory optimization
type ConcurrentTxnBST struct {
	tree     TxnBST
	rwLock   sync.RWMutex
	nodePool *NodePool
}

// NewConcurrentTxnBST creates a new concurrent transaction binary search tree
func NewConcurrentTxnBST() *ConcurrentTxnBST {
	return &ConcurrentTxnBST{
		tree:     TxnBST{},
		nodePool: NewNodePool(),
	}
}

// Insert adds a transaction to the tree in a thread-safe manner
func (ct *ConcurrentTxnBST) Insert(value Transaction) {
	ct.rwLock.Lock()
	defer ct.rwLock.Unlock()
	
	ct.tree.Root = ct.insertNodeWithPool(ct.tree.Root, value)
}

// insertNodeWithPool is similar to insertNode but uses the node pool
func (ct *ConcurrentTxnBST) insertNodeWithPool(node *treeNode, value Transaction) *treeNode {
	// Standard BST insertion
	if node == nil {
		newNode := ct.nodePool.Get()
		newNode.Value = value
		newNode.Height = 0
		return newNode
	}

	if value.Amount <= node.Value.Amount {
		node.Left = ct.insertNodeWithPool(node.Left, value)
	} else {
		node.Right = ct.insertNodeWithPool(node.Right, value)
	}

	// Update height of this node
	node.updateHeight()

	// Get the balance factor to check if this node became unbalanced
	balance := node.balanceFactor()

	// Left-Left Case
	if balance > 1 && value.Amount <= node.Left.Value.Amount {
		return rotateRight(node)
	}

	// Right-Right Case
	if balance < -1 && value.Amount > node.Right.Value.Amount {
		return rotateLeft(node)
	}

	// Left-Right Case
	if balance > 1 && value.Amount > node.Left.Value.Amount {
		node.Left = rotateLeft(node.Left)
		return rotateRight(node)
	}

	// Right-Left Case
	if balance < -1 && value.Amount <= node.Right.Value.Amount {
		node.Right = rotateRight(node.Right)
		return rotateLeft(node)
	}

	// No balancing needed
	return node
}

// Search finds a transaction with the given amount in a thread-safe manner
func (ct *ConcurrentTxnBST) Search(value TransactionAmtDataType) *Transaction {
	ct.rwLock.RLock()
	defer ct.rwLock.RUnlock()
	
	return ct.tree.Root.searchNode(value)
}

// InorderTraversal returns all transactions in sorted order in a thread-safe manner
func (ct *ConcurrentTxnBST) InorderTraversal() []Transaction {
	ct.rwLock.RLock()
	defer ct.rwLock.RUnlock()
	
	result := []Transaction{}
	ct.tree.Root.inorder(&result)
	return result
}

// Remove removes a transaction from the tree in a thread-safe manner
func (ct *ConcurrentTxnBST) Remove(value Transaction) {
	ct.rwLock.Lock()
	defer ct.rwLock.Unlock()
	
	// Track nodes to be recycled
	nodesToRecycle := make([]*treeNode, 0)
	ct.tree.Root = ct.removeNodeWithRecycling(ct.tree.Root, value, &nodesToRecycle)
	
	// Recycle nodes
	for _, node := range nodesToRecycle {
		ct.nodePool.Put(node)
	}
}

// removeNodeWithRecycling is similar to removeNode but tracks nodes to be recycled
func (ct *ConcurrentTxnBST) removeNodeWithRecycling(node *treeNode, value Transaction, nodesToRecycle *[]*treeNode) *treeNode {
	if node == nil {
		return nil
	}

	// Standard BST deletion
	if value.Amount < node.Value.Amount {
		// Value is in the left subtree.
		node.Left = ct.removeNodeWithRecycling(node.Left, value, nodesToRecycle)
	} else if value.Amount > node.Value.Amount {
		// Value is in the right subtree.
		node.Right = ct.removeNodeWithRecycling(node.Right, value, nodesToRecycle)
	} else {
		// Node to be deleted found.
		// Check if it's the exact transaction (by ID) or just same amount
		if node.Value.ID != value.ID {
			// If IDs don't match, look for the exact transaction in the right subtree
			// (since we might have multiple transactions with the same amount)
			node.Right = ct.removeNodeWithRecycling(node.Right, value, nodesToRecycle)
		} else {
			// This is the exact transaction to remove
			if node.Left == nil && node.Right == nil {
				// Case 1: Node has no children.
				*nodesToRecycle = append(*nodesToRecycle, node)
				return nil
			} else if node.Left == nil {
				// Case 2: Node has only a right child.
				rightChild := node.Right
				*nodesToRecycle = append(*nodesToRecycle, node)
				return rightChild
			} else if node.Right == nil {
				// Case 2: Node has only a left child.
				leftChild := node.Left
				*nodesToRecycle = append(*nodesToRecycle, node)
				return leftChild
			} else {
				// Case 3: Node has both left and right children.
				// Find the minimum value in the right subtree (inorder successor).
				minValue := findMinValue(node.Right)
				node.Value = minValue
				// Remove the inorder successor.
				node.Right = ct.removeNodeWithRecycling(node.Right, minValue, nodesToRecycle)
			}
		}
	}

	// If the tree had only one node, return
	if node == nil {
		return nil
	}

	// Update height of this node
	node.updateHeight()

	// Get the balance factor to check if this node became unbalanced
	balance := node.balanceFactor()

	// Left-Left Case
	if balance > 1 && node.Left.balanceFactor() >= 0 {
		return rotateRight(node)
	}

	// Left-Right Case
	if balance > 1 && node.Left.balanceFactor() < 0 {
		node.Left = rotateLeft(node.Left)
		return rotateRight(node)
	}

	// Right-Right Case
	if balance < -1 && node.Right.balanceFactor() <= 0 {
		return rotateLeft(node)
	}

	// Right-Left Case
	if balance < -1 && node.Right.balanceFactor() > 0 {
		node.Right = rotateRight(node.Right)
		return rotateLeft(node)
	}

	// No balancing needed
	return node
}

// GetStats returns statistics about the node pool
func (ct *ConcurrentTxnBST) GetStats() (allocated, recycled int64) {
	return ct.nodePool.Stats()
}
