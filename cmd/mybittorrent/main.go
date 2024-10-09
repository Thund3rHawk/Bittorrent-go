package main

import (
	"encoding/json"
	"fmt"
	"os"

	info "github.com/codecrafters-io/bittorrent-starter-go/internal/utils"
)

var _ = json.Marshal

func main() {
	command := os.Args[1]
	if command == "decode" {
		info.CommandDecode(command)
	} else if command == "info" {
		info.CommandInfo(command)
	} else if command == "peers" {
		info.CommandPeers(command)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
