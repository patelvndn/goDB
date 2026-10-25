package btree

import (
	"github.com/patelvndn/goDB/internal/btree/btreeNode"
	"github.com/patelvndn/goDB/internal/pager"
)

type pgr = pager.Pager
var pageSize int = 4096

type Node = btreeNode.Node
type Btree struct {
	root *Node
	t int
	pager *pgr
}

func New(t int) *Btree {

	pager, err := pager.New("database", pageSize)

	if err != nil {
		return nil
	}

	return &Btree{
		root: btreeNode.New(),
		t: t,
		pager: pager,
	}
}

func (bt *Btree) InsertNode(key uint32) (error){
	// this logic needs to be moved
	root, err := bt.root.Insert(key, bt.t, bt.pager)

	if err != nil {
		return err
	}

	bt.root = root
	return nil 
}


func (bt *Btree) SearchNode(key uint32) (*Node, int, bool) {
	return bt.root.Search(key, bt.t, bt.pager)
}

func (bt *Btree) RemoveNode(key uint32) error {
	root, err := bt.root.Remove(key, bt.t, bt.pager)

	if err != nil {
		return err
	}

	bt.root = root

	return nil
}

