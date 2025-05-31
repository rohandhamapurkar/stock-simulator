package exchange

// TxnBST represents a self-balancing AVL Tree for transactions.
type TxnBST struct {
    Root *treeNode
}

// treeNode represents a node in the TxnBST.
type treeNode struct {
    Value  Transaction
    Left   *treeNode
    Right  *treeNode
    Height int // Height of the node for AVL balancing
}

// height returns the height of the node.
// A nil node has height -1, a leaf node has height 0.
func height(node *treeNode) int {
    if node == nil {
        return -1
    }
    return node.Height
}

// updateHeight updates the height of the node based on its children's heights.
func (node *treeNode) updateHeight() {
    leftHeight := height(node.Left)
    rightHeight := height(node.Right)
    
    // Height is 1 + the maximum height of the children
    if leftHeight > rightHeight {
        node.Height = leftHeight + 1
    } else {
        node.Height = rightHeight + 1
    }
}

// balanceFactor returns the balance factor of the node.
// Balance factor = height of left subtree - height of right subtree
func (node *treeNode) balanceFactor() int {
    if node == nil {
        return 0
    }
    return height(node.Left) - height(node.Right)
}

// rotateRight performs a right rotation on the given node.
func rotateRight(y *treeNode) *treeNode {
    x := y.Left
    T2 := x.Right

    // Perform rotation
    x.Right = y
    y.Left = T2

    // Update heights
    y.updateHeight()
    x.updateHeight()

    // Return new root
    return x
}

// rotateLeft performs a left rotation on the given node.
func rotateLeft(x *treeNode) *treeNode {
    y := x.Right
    T2 := y.Left

    // Perform rotation
    y.Left = x
    x.Right = T2

    // Update heights
    x.updateHeight()
    y.updateHeight()

    // Return new root
    return y
}

// Insert inserts a value into the TxnBST.
func (bst *TxnBST) Insert(value Transaction) {
    bst.Root = bst.Root.insertNode(value)
}

// insertNode inserts a value into the subtree rooted at the given node.
// Returns the new root of the subtree after insertion and balancing.
func (node *treeNode) insertNode(value Transaction) *treeNode {
    // Standard BST insertion
    if node == nil {
        return &treeNode{Value: value, Height: 0}
    }

    if value.Amount <= node.Value.Amount {
        node.Left = node.Left.insertNode(value)
    } else {
        node.Right = node.Right.insertNode(value)
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

// Search searches for a value in the TxnBST and returns pointer to a transaction if found.
func (bst *TxnBST) Search(value TransactionAmtDataType) *Transaction {
    return bst.Root.searchNode(value)
}

// searchNode searches for a value in the subtree rooted at the given node.
func (node *treeNode) searchNode(value TransactionAmtDataType) *Transaction {
    if node == nil {
        return nil
    }

    if value == node.Value.Amount {
        return &node.Value
    } else if value < node.Value.Amount {
        if node.Left == nil {
            return nil
        }
        return node.Left.searchNode(value)
    } else {
        if node.Right == nil {
            return nil
        }
        return node.Right.searchNode(value)
    }
}

// InorderTraversal performs an inorder traversal of the TxnBST and returns the values in sorted order.
func (bst *TxnBST) InorderTraversal() []Transaction {
    result := []Transaction{}
    bst.Root.inorder(&result)
    return result
}

// inorder appends the values of the subtree rooted at the given node to the result slice in sorted order.
func (node *treeNode) inorder(result *[]Transaction) {
    if node != nil {
        if node.Left != nil {
            node.Left.inorder(result)
        }
        *result = append(*result, node.Value)
        if node.Right != nil {
            node.Right.inorder(result)
        }
    }
}

// Remove removes a node with the given value from the TxnBST.
func (bst *TxnBST) Remove(value Transaction) {
    bst.Root = bst.Root.removeNode(value)
}

// removeNode removes a node with the given value from the subtree rooted at the given node.
// Returns the new root of the subtree after removal and balancing.
func (node *treeNode) removeNode(value Transaction) *treeNode {
    if node == nil {
        return nil
    }

    // Standard BST deletion
    if value.Amount < node.Value.Amount {
        // Value is in the left subtree.
        node.Left = node.Left.removeNode(value)
    } else if value.Amount > node.Value.Amount {
        // Value is in the right subtree.
        node.Right = node.Right.removeNode(value)
    } else {
        // Node to be deleted found.
        // Check if it's the exact transaction (by ID) or just same amount
        if node.Value.ID != value.ID {
            // If IDs don't match, look for the exact transaction in the right subtree
            // (since we might have multiple transactions with the same amount)
            node.Right = node.Right.removeNode(value)
        } else {
            // This is the exact transaction to remove
            if node.Left == nil && node.Right == nil {
                // Case 1: Node has no children.
                return nil
            } else if node.Left == nil {
                // Case 2: Node has only a right child.
                return node.Right
            } else if node.Right == nil {
                // Case 2: Node has only a left child.
                return node.Left
            } else {
                // Case 3: Node has both left and right children.
                // Find the minimum value in the right subtree (inorder successor).
                minValue := findMinValue(node.Right)
                node.Value = minValue
                // Remove the inorder successor.
                node.Right = node.Right.removeNode(minValue)
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

// findMinValue returns the minimum value in the subtree rooted at the given node.
func findMinValue(node *treeNode) Transaction {
    current := node
    for current.Left != nil {
        current = current.Left
    }
    return current.Value
}
