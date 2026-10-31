package btree

import (
	"crypto/rand"

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
	// schema (schema will have a comparator)
	/*
	schema
		[]dtype

	dtype interface
		- decode schema
			[]bytes -> type
			type -> bytes
		- comparison
		- string
		- equality

	ex String, Int, Time, Boolean
	*/
}

func New(t int) (*Btree, error) {

	pager, err := pager.New("database", pageSize)

	p, err := pager.AllocatePage() // reserve space for metadata

	if err != nil {
		return nil, err
	}

	buf := make([]byte, pageSize) // random data for now
	_, err = rand.Read(buf)
	p.Data = buf

	p, err = pager.WritePage(p.Id, buf)

	if err != nil {
		return nil, err
	}

	root := btreeNode.New(pager)
	
	return &Btree{
		root: root,
		t: t,
		pager: pager,
	}, nil 
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

