package btree

import (
	"github.com/patelvndn/goDB/internal/btree/btreeNode"
)

/*
	An abstraction...
	view keys as page numbers
	and nodes are somethign completely different
	the

	will handle the logic concerned with pager, parsing etc.
		actual tree logic and insert logic will be handled by node.go in btreeNode module
*/

type node = btreeNode.Node


type Btree struct {
	root *node
	t int
}

func New(t int) *Btree {
	return &Btree{
		root: btreeNode.New(),
		t: t,
	}
}


func (bt *Btree) InsertNode(key int) (error){
	
	promoted, right, split, err := bt.root.Insert(key, bt.t);
	if err != nil {
		return err
	}

	var newRoot *node
	var left *node

	if split {
		left = bt.root
		newRoot = &node{
			Keys: []int{promoted},
			Children: []*node{left,right},
		}
		bt.root = newRoot
	}
	return nil 
}



func (bt *Btree) SearchNode(key int) (*node, int, bool) {
	return bt.root.Search(key)
}

// eventually abstract away all the node methods
// btree module
// btreeNode module
// data encoding module 
func (bt *Btree) RemoveNode(key int) error {
	return bt.root.Remove(key, bt.t)
}


// func Merge(n *Node, i int){

// }

