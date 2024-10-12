package info

import (
	"fmt"
	"os"
	"strings"

	"github.com/jackpal/bencode-go"
)

type Data struct {
	url         string
	length      int64
	infoHash    string
	pieceLength int64
	pieceHash   string
}

func GetInfo() Data {
	data, error := os.ReadFile(os.Args[2])
	var url string
	var length int64
	var infoHash string
	var pieceLength int64
	var pieceHash string
	if error != nil {
		fmt.Println("Error fetchinf file:", error)
	}
	reader := strings.NewReader(string(data))
	// Decode the bencoded data
	decoded, err := bencode.Decode(reader)

	if err != nil {
		fmt.Println("Error decoding:", err)
	}
	output := decoded

	// Ensure the decoded data is a map
	decodedMap, ok := output.(map[string]interface{})
	if !ok {
		fmt.Println("Error: Decoded data is not a map")
	}
	if announce, found := decodedMap["announce"]; found {
		url = fmt.Sprintf("%v", announce)
	} else {
		fmt.Println("Error: 'announce' field not found")
	}

	if info, found := decodedMap["info"]; found {
		infoMap, ok := info.(map[string]interface{})
		if !ok {
			fmt.Println("Error: 'info' field is not a map")
		}
		length = infoMap["length"].(int64)

		// Bencode the "info" field
		var bencodedInfo strings.Builder
		err = bencode.Marshal(&bencodedInfo, infoMap)
		if err != nil {
			fmt.Println("Error bencoding 'info' field:", err)
		}
		hashedString := Infohash(bencodedInfo.String())

		infoHash = hashedString

		pieceLength = infoMap["piece length"].(int64)

		pieceHash = fmt.Sprintf("%x", infoMap["pieces"])

	} else {
		fmt.Println("Error: 'info' field not found")
	}
	details := Data{
		url:         url,
		length:      length,
		infoHash:    infoHash,
		pieceLength: pieceLength,
		pieceHash:   pieceHash,
	}
	return details
}
