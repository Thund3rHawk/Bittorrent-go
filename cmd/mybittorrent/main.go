package main

import (
	"encoding/json"
	"fmt"
	"os"

	info "github.com/codecrafters-io/bittorrent-starter-go/internal/command"
)

var _ = json.Marshal

func main() {
	command := os.Args[1]
	if command == "decode" {
		info.CommandDecode()
	} else if command == "info" {
		info.CommandInfo()
	} else if command == "peers" {
		info.CommandPeers(os.Args[2])
	} else if command == "handshake" {
		info.CommandHandshake()
	} else if command == "download_piece" {
		info.CommandDownloadPiece()
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
