package btreeNode

import (
	"errors"
	"slices"
)

type Node struct {
	Children []*Node
	Keys []int
	IsLeaf bool
}

func New() *Node {
	return &Node{IsLeaf: true}
}


func findIndex (keys []int, key int) int {
	for i, k := range keys {
		if key <= k {
			return i
		}
	}
	return len(keys)
}

func (n *Node) Insert(key int, t int) (int, *Node, bool, error) {
	i := findIndex(n.Keys, key) 
	

	// this is the node with the key
	if 0 <= i && i < len(n.Keys) && n.Keys[i] == key {
		return -1, nil, false, errors.New("Key already exists in node")
	}

	if n.IsLeaf {
		n.Keys = slices.Insert(n.Keys, i, key)
		// insert key into index i 
	} else {
		// propogate the split back up, if it happened
		promoted, right, split, err := n.Children[i].Insert(key,t);
		if err != nil {
			return -1, nil, false, err
		}

		if split {
			n.Keys = slices.Insert(n.Keys, i, promoted)
			n.Children = slices.Insert(n.Children, i+1, right)
		}

	}

	// split
	if len(n.Keys) > 2*t-1 {
        return n.split(t)
    } 
	return 0, nil, false, nil
}


// search within the node itself to find out where the key exists
func (n *Node) Search(key int) (*Node, int, bool) {
	i := findIndex(n.Keys, key)		
	// this is the node with the key
	if i < len(n.Keys) && n.Keys[i] == key {
		return n, i, true
	}

	// this node is a leaf, no more children to search
	if n.IsLeaf {
		return nil, 0, false
	}

	// recursively search the children that would have the key
	return n.Children[i].Search(key)
}

func (n *Node)split(t int) (int, *Node, bool, error)  {
	// when we split a node each side should have t
	// splitting a node means half the keys get shifted to a new node (right node)
	// and the middle node gets promoted up
	// n turns into left
	promoted := n.Keys[t]

	lk := make([]int, t)
	copy(lk, n.Keys[:t])

	rk := make([]int, len(n.Keys)-(t+1))
	copy(rk, n.Keys[t+1:])

	n.Keys = lk
	
	var right *Node

	// need to split children too, but how?
	// node could be a leaf still aka it won't have any childrn
	right = &Node{
			Keys: rk,
			IsLeaf: true,
	}

	if !n.IsLeaf {
		rc := make([]*Node, len(n.Children)-(t+1))
		copy(rc, n.Children[t+1:],)
		right.Children = rc

		lc := make([]*Node, t+1)
		copy(lc, n.Children[:t+1])
		n.Children = lc

		right.IsLeaf = false
	}

	return promoted, right, true, nil
}

func (n *Node) Remove(key int, t int) error {

	i := findIndex(n.Keys, key)
	found := false
	if i < len(n.Keys) && n.Keys[i] == key {
		found = true
	}

	if n.IsLeaf {
		// key is not in internal node
		if found {
			n.Keys = slices.Delete(n.Keys,i,i+1)
			return nil
		} 
		return errors.New("Key not found error")
	}

	if found {
		// key is in this internal node
		left := n.Children[i]
		right := n.Children[i+1]

		// removing a node from nodee with size t => 
		if len(left.Keys) >= t {
			pred := slices.Max(left.Keys)
			n.Keys[i] = pred
			left.Remove(pred, t)
		} else if len(right.Keys) >= t {
			succ := slices.Min(right.Keys)
			n.Keys[i] = succ
			right.Remove(succ, t)
		} else {
			merge(n,i)
			left.Remove(key, t)
		}

		// 
	} else {
		child := n.Children[i]
		// fix merge
		child.Remove(i,t)
	}

	return nil 
}


func merge (n *Node, i int) error {
	return nil 
}