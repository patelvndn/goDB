package btree

import (
	"github.com/patelvndn/goDB/internal/btree/btreeNode"
	"github.com/patelvndn/goDB/internal/pager"
)

type node = btreeNode.Node
type pgr = pager.Pager
type page = pager.Page

var pageSize int = 4024

type Btree struct {
	root *node
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

