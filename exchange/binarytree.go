package exchange

// TxnBST represents a Binary Search Tree for transactions.
type TxnBST struct {
    Root *treeNode
}

// treeNode represents a node in the TxnBST.
type treeNode struct {
    Value Transaction
    Left  *treeNode
    Right *treeNode
}

// Insert inserts a value into the TxnBST.
func (bst *TxnBST) Insert(value Transaction) {
    if bst.Root == nil {
        bst.Root = &treeNode{Value: value}
    } else {
        bst.Root.insertNode(value)
    }
}

// insertNode inserts a value into the subtree rooted at the given node.
func (node *treeNode) insertNode(value Transaction) {
    if value.Amount <= node.Value.Amount {
        if node.Left == nil {
            node.Left = &treeNode{Value: value}
        } else {
            node.Left.insertNode(value)
        }
    } else {
        if node.Right == nil {
            node.Right = &treeNode{Value: value}
        } else {
            node.Right.insertNode(value)
        }
    }
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
func (node *treeNode) removeNode(value Transaction) *treeNode {
    if node == nil {
        return nil
    }

    if value.Amount < node.Value.Amount {
        // Value is in the left subtree.
        node.Left = node.Left.removeNode(value)
    } else if value.Amount > node.Value.Amount {
        // Value is in the right subtree.
        node.Right = node.Right.removeNode(value)
    } else {
        // Node to be deleted found.
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
    return node
}

// findMinValue returns the minimum value in the subtree rooted at the given node.
func findMinValue(node *treeNode) Transaction {
    for node.Left != nil {
        node = node.Left
    }
    return node.Value
}
