package btree

import "slices"

/*
	An abstraction...
	view keys as page numbers
	and nodes are somethign completely different
	the
*/
type Node struct {
	children []*Node
	keys []int
	isLeaf bool
}

type Btree struct {
	root *Node
	t int
}

func New(t int) *Btree {
	return &Btree{
		root:&Node{isLeaf: true,},
		t: t,
	}
}


func (bt *Btree) InsertNode(key int) (int, *Node, bool) {
	// inserting a node into the key... we can assume that the node doesn't exist...
	// will use a similar method of search
	return bt.root.insert(key, bt.t)
}

func (n *Node) insert(key int, t int) (int, *Node, bool) {
	i := findIndex(n.keys, key) 
	if n.isLeaf {
		n.keys = slices.Insert(n.keys, i, key)
		// insert key into index i 
	} else {
		// propogate the split back up 
		promoted, right, split := n.children[i].insert(key,t);
		if split {
			n.keys = slices.Insert(n.keys, i, promoted)
			n.children = slices.Insert(n.children, i+1, right)
		}
	}

	// split
	if len(n.keys) > 2*t-1 {
        return n.split(t)
    } 

	return 0, nil, false
}


func (n *Node)split(t int) (int, *Node, bool)  {
	// when we split a node each side should have t
	// splitting a node means half the keys get shifted to a new node (right node)
	// and the middle node gets promoted up

	promoted := n.keys[t]
	lk := n.keys[0:t]
	rk := n.keys[t+1:]
	n.keys = lk
	var right *Node

	// need to split children too, but how?
	// node could be a leaf still aka it won't have any childrn
	right = &Node{
			keys: rk,
	}

	if !n.isLeaf {
		rightChildren := n.children[t+1:]
		right.children = rightChildren
		right.isLeaf = false
	}

	return promoted, right, true
}

func (bt *Btree) SearchNode(key int) (*Node, int, bool) {
	return bt.root.search(key)
}

func (bt *Btree) RemoveNode() {
	
}


func findIndex (keys []int, key int) int {
	for i, k := range keys {
		if key <= k {
			return i
		}
	}
	return len(keys)
}

// search within the node itself to find out where the key exists
func (n *Node) search(key int) (*Node, int, bool) {
	i := findIndex(n.keys, key)		

	// this is the node with the key
	if i < len(n.keys) && n.keys[i] == key {
		return n, i, true
	}

	// this node is a leaf, no more children to search
	if n.isLeaf {
		return nil, 0, false
	}

	// recursively search the children that would have the key
	return n.children[i].search(key)
}