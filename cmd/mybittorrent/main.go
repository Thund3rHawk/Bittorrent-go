package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jackpal/bencode-go"
)

// Ensures gofmt doesn't remove the "os" encoding/json import (feel free to remove this!)
var _ = json.Marshal

func main() {

	command := os.Args[1]

	if command == "decode" {

		bencodedValue := os.Args[2]
		reader := strings.NewReader(bencodedValue)

		// Create a variable to hold the decoded data
		var decoded interface{}

		// Decode the bencoded data
		decoded, err := bencode.Decode(reader)

		if err != nil {
			fmt.Println("Error decoding:", err)
			return
		}

		// Convert the decoded data to JSON format
		jsonOutput, err := json.Marshal(decoded)
		if err != nil {
			fmt.Println("Error converting to JSON:", err)
			return
		}

		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
