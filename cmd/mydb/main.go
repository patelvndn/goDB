package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/patelvndn/goDB/internal/storage"
)

func main() {


	db, err := storage.NewDatabase()

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		input = strings.TrimSpace(input)
		args := strings.Fields(input)

		if len(args) == 0 {
			continue
		}

		switch strings.ToUpper(args[0]) {

		case storage.OpSet:
			if len(args) != 3 {
				fmt.Println("Usage: SET <key> <value>")
				continue
			}

			db.Set(args[1], args[2])
			fmt.Println("OK")

		case storage.OpGet:
			if len(args) != 2 {
				fmt.Println("Usage: GET <key>")
				continue
			}

			value, ok := db.Get(args[1])
			if !ok {
				fmt.Println("(nil)")
			} else {
				fmt.Println(value)
			}

		case storage.OpDelete:
			if len(args) != 2 {
				fmt.Println("Usage: DELETE <key>")
				continue
			}

			ok := db.Delete(args[1])

			if ok {
				fmt.Println("Deleted")
			} else {
				fmt.Println("No key " + args[1] + " exists do delete")
			}

		case "EXIT":
			db.Close()
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Unknown command")
		}
	}
}
