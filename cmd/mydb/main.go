package main

import (
	"fmt"

	"github.com/patelvndn/goDB/internal/storage"
)

func main() {
    fmt.Println("My Database")
	myDB := storage.DB{}
	fmt.Println(myDB)
}