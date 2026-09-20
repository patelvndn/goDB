package btree

import (
	"errors"
	"slices"
)

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
		root:&Node{
				isLeaf: true,
				keys: make([]int, 0),
				children: make([]*Node, 0)},
		t: t,
	}
}


func (bt *Btree) InsertNode(key int) (error){
	// inserting a node into the key... we can assume that the node doesn't exist...
	// will use a similar method of search	
	// if the root splits we create a new root
	
	promoted, right, split, err := bt.root.insert(key, bt.t);
	if err != nil {
		return err
	}

	var newRoot *Node
	var left *Node

	if split {
		left = bt.root

		newRoot = &Node{
			keys: []int{promoted},
			children: []*Node{left,right},
		}
		bt.root = newRoot
	}
	return nil 
}

func (n *Node) insert(key int, t int) (int, *Node, bool, error) {
	i := findIndex(n.keys, key) 
	

	// this is the node with the key
	if 0 <= i && i < len(n.keys) && n.keys[i] == key {
		return -1, nil, false, errors.New("Key already exists in node")
	}

	if n.isLeaf {
		n.keys = slices.Insert(n.keys, i, key)
		// insert key into index i 
	} else {
		// propogate the split back up, if it happened
		promoted, right, split, err := n.children[i].insert(key,t);
		if err != nil {
			return -1, nil, false, err
		}

		if split {
			n.keys = slices.Insert(n.keys, i, promoted)
			n.children = slices.Insert(n.children, i+1, right)
		}

	}

	// split
	if len(n.keys) > 2*t-1 {
        return n.split(t)
    } 
	return 0, nil, false, nil
}


func (n *Node)split(t int) (int, *Node, bool, error)  {
	// when we split a node each side should have t
	// splitting a node means half the keys get shifted to a new node (right node)
	// and the middle node gets promoted up
	// n turns into left
	promoted := n.keys[t]

	lk := make([]int, t)
	copy(lk, n.keys[:t])

	rk := make([]int, len(n.keys)-(t+1))
	copy(rk, n.keys[t+1:])

	n.keys = lk
	
	var right *Node

	// need to split children too, but how?
	// node could be a leaf still aka it won't have any childrn
	right = &Node{
			keys: rk,
			isLeaf: true,
	}

	if !n.isLeaf {
		rc := make([]*Node, len(n.children)-(t+1))
		copy(rc, n.children[t+1:],)
		right.children = rc

		lc := make([]*Node, t+1)
		copy(lc, n.children[:t+1])
		n.children = lc

		right.isLeaf = false
	}

	return promoted, right, true, nil
}

func (bt *Btree) SearchNode(key int) (*Node, int, bool) {
	return bt.root.search(key)
}

// eventually abstract away all the node methods
// btree module
// btreeNode module
// data encoding module 
func (bt *Btree) RemoveNode(key int) error {
	return bt.root.remove(key)
}

func (n *Node) remove(key int) error {
	i := findIndex(n.keys, key)
	found := false
	if i < len(n.keys) && n.keys[i] == key {
		found = true
	}

	if found {
		// delete logic since we found it
		if n.isLeaf{
			n.keys = slices.Delete(n.keys, i,i+1)
			return nil
		}

		// now it gets a bit harder with merges and such, there were three cases overall


	} else {
		// havent found it
		return n.children[i].remove(key)
	}

	return nil 
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