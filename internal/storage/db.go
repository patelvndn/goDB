package storage

import (
	"github.com/patelvndn/goDB/internal/btree"
)

const (
    OpSet    = "SET"
    OpDelete = "DELETE"
	OpGet	 = "GET"
)

type bt = btree.Btree

type DB struct {
	btree *bt
}

func NewDatabase() (*DB, error) {

	bt, err := btree.New(10)

	if err != nil {
		panic("Error creating the database")
	}
	database := &DB{
		btree: bt,
	}
	return database, nil 
}

func (database *DB) Set(key uint32, value string) (error) {
	return database.btree.InsertNode(key)
}

func (database *DB) Get(key uint32) (string, bool) {
	node, i, found := database.btree.SearchNode(key)
	if ! found {
		return "", found
	}
	return node.Values[i], true
}

// func (database *DB) Delete(key string) (bool) {

// }