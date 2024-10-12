package info

import (
	"fmt"
	"os"
	"strings"

	"github.com/jackpal/bencode-go"
)

func CommandInfo() {
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
		hashedString := Infohash(bencodedInfo.String())
		fmt.Printf("Info Hash: %v\n", hashedString)
		fmt.Printf("Piece Length: %v\n", infoMap["piece length"])

		// hashedPieces := infoHash.Infohash((infoMap["pieces"]))
		fmt.Printf("Piece Hashes: %x", infoMap["pieces"])

	} else {
		fmt.Println("Error: 'info' field not found")
	}
}
