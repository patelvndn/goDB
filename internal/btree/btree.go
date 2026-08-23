package btree

type Node struct {
	key int
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


func (bt *Btree) InsertNode(key int) {
	// descend into correct leaf
	// insert and then split if needed

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