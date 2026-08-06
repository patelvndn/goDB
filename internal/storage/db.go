package storage

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

const (
    OpSet    = "SET"
    OpDelete = "DELETE"
	OpGet	 = "GET"
)

type DB struct {
	fileName string
	separatingOperator string
	db map[string]string
	file *os.File

}

func NewDatabase() (*DB, error) {

	database := &DB{
		db: make(map[string]string),
		fileName: "database.db",
		separatingOperator: ":",
	}

	database.readFromFile()
	return database, nil 
}

func (database *DB) Set(key string, value string) (error) {

	if (strings.Contains(key, database.separatingOperator)) {
		return errors.New("<key> cannot contain a semicolon \":\"")
	}

	database.db[key] = value
	database.file.WriteString(OpSet + database.separatingOperator + key + database.separatingOperator + value + "\n")
	return nil 
}

func (database *DB) Get(key string) (string, bool) {
	value, ok := database.db[key]
	return value, ok
}

func (database *DB) Delete(key string) (bool) {
	_, ok := database.db[key]

	if ok {
		delete(database.db, key)
		database.file.WriteString(OpDelete + database.separatingOperator + key + "\n" ) // tombstone
		return true
	}
	return false
}

func (database *DB) readFromFile() error {
	file, err := os.OpenFile(database.fileName, os.O_CREATE | os.O_RDWR, 0600)

	database.file = file
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)


	for scanner.Scan() {
		line := scanner.Text() 
		args := strings.Split(line, database.separatingOperator)


		if len(args) >= 3 && args[0] != OpDelete{
			database.db[args[1]] = args[2]
		}
	}
	return nil 
}

func (db *DB) Close() error {
    return db.file.Close()
}