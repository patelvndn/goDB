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


func (bt *Btree) InsertNode(key int) (*Node, bool) {
	// inserting a node into the key... we can assume that the node doesn't exist...
	// will use a similar method of search
	return bt.root.insert(key, bt.t)
}

func (n *Node) insert(key int, t int) (*Node, bool) {
	i := findIndex(n.keys, key) 
	if n.isLeaf {
		n.keys = slices.Insert(n.keys,i, key)
		// insert key into index i 
		if len(n.keys) > (2 *t + 1) {
			// split, median is always t since we are inserting in sorted order
			median := n.keys[t]
			left := n.keys[:t]
			right := n.keys[t+1:]

			leftNode := &Node{keys: left, isLeaf: true}
			rightNode := &Node{keys: right, isLeaf: true}

			n.children = append(n.children,leftNode, rightNode)
			n.keys = []int{median}
			n.isLeaf = false
		}
		
	} else {
		return n.children[i].insert(key, t)
	}
	return nil, false
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