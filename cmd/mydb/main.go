package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/patelvndn/goDB/internal/storage"
)

func main() {


	db := storage.NewDatabase()
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

		case "SET":
			if len(args) != 3 {
				fmt.Println("Usage: SET <key> <value>")
				continue
			}

			db.SET(args[1], args[2])
			fmt.Println("OK")

		case "GET":
			if len(args) != 2 {
				fmt.Println("Usage: GET <key>")
				continue
			}

			value, ok := db.GET(args[1])
			if !ok {
				fmt.Println("(nil)")
			} else {
				fmt.Println(value)
			}

		case "DELETE":
			if len(args) != 2 {
				fmt.Println("Usage: DELETE <key>")
				continue
			}

			db.DELETE(args[1])
			fmt.Println("Deleted")

		case "EXIT":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Unknown command")
		}
	}
}
