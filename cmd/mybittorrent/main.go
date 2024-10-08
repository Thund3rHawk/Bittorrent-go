package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	infoHash "github.com/codecrafters-io/bittorrent-starter-go/utils"
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
	} else if command == "info" {
		data, error := os.ReadFile(os.Args[2])

		if error != nil {
			fmt.Println("Error fetchinf file:", error)
			return
		}
		reader := strings.NewReader(string(data))
		// Decode the bencoded data
		decoded, err := bencode.Decode(reader)

		if err != nil {
			fmt.Println("Error decoding:", err)
			return
		}
		output := decoded

		// Ensure the decoded data is a map
		decodedMap, ok := output.(map[string]interface{})
		if !ok {
			fmt.Println("Error: Decoded data is not a map")
			return
		}

		// Print the "announce" field
		if announce, found := decodedMap["announce"]; found {
			fmt.Printf("Tracker URL: %v\n", announce)
		} else {
			fmt.Println("Error: 'announce' field not found")
		}

		// Print the "info" field
		if info, found := decodedMap["info"]; found {
			infoMap, ok := info.(map[string]interface{})
			if !ok {
				fmt.Println("Error: 'info' field is not a map")
				return
			}
			fmt.Printf("Length: %v\n", infoMap["length"])
			// Bencode the "info" field
			var bencodedInfo strings.Builder
			err = bencode.Marshal(&bencodedInfo, infoMap)
			if err != nil {
				fmt.Println("Error bencoding 'info' field:", err)
				return
			}
			hashedString := infoHash.Infohash(bencodedInfo.String())
			fmt.Printf("Info Hash: %v\n", hashedString)

		} else {
			fmt.Println("Error: 'info' field not found")
		}
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
