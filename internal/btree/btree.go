package btree

import (
	"github.com/patelvndn/goDB/internal/btree/btreeNode"
)

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
	// this logic needs to be moved
	root, err := bt.root.Insert(key, bt.t)

	if err != nil {
		return err
	}

	bt.root = root
	return nil 
}


func (bt *Btree) SearchNode(key int) (*node, int, bool) {
	return bt.root.Search(key)
}

func (bt *Btree) RemoveNode(key int) error {
	root, err := bt.root.Remove(key, bt.t)

	if err != nil {
		return err
	}

	bt.root = root

	return nil
}

