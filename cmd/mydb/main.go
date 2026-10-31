package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
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
			key, err := strconv.ParseUint(args[0], 10, 32)

			if err != nil {
				fmt.Println("There was an error converting the string to an integer value: ", err.Error())
			}
			db.Set(uint32(key), args[2])
			fmt.Println("OK")

		case storage.OpGet:
			if len(args) != 2 {
				fmt.Println("Usage: GET <key>")
				continue
			}
			key, err := strconv.ParseUint(args[0], 10, 32)

			if err != nil {
				fmt.Println("There was an error converting the string to an integer value: ", err.Error())
			}

			value, ok := db.Get(uint32(key))
			if !ok {
				fmt.Println("(nil)")
			} else {
				fmt.Println(value)
			}

		case storage.OpDelete:
			fmt.Println("Not yet implemented")
			// if len(args) != 2 {
			// 	fmt.Println("Usage: DELETE <key>")
			// 	continue
			// }

			// ok := db.Delete(args[1])

			// if ok {
			// 	fmt.Println("Deleted")
			// } else {
			// 	fmt.Println("No key " + args[1] + " exists do delete")
			// }

		case "EXIT":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Unknown command")
		}
	}

	
}
