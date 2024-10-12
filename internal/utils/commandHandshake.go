package info

import (
	"encoding/hex"
	"fmt"
	"net"
	"os"
)

func CommandHandshake() {
	peerAddress := os.Args[3]
	listner, err := net.Dial("tcp", peerAddress)
	if err != nil {
		fmt.Println("tcp connection error")
	}
	defer listner.Close()

	pstrlen := byte(19) // The length of the string "BitTorrent protocol"
	pstr := []byte("BitTorrent protocol")
	reserved := make([]byte, 8) // Eight zeros
	infoHashBytes, _ := hex.DecodeString(GetInfo().infoHash)
	handshake := append([]byte{pstrlen}, pstr...)
	handshake = append(handshake, reserved...)
	handshake = append(handshake, infoHashBytes...)
	handshake = append(handshake, []byte{0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9}...)

	_, err = listner.Write(handshake)

	if err != nil {
		fmt.Println("tcp message error")
	}

	buf := make([]byte, 68)
	_, err = listner.Read(buf)
	if err != nil {
		fmt.Println("failed:", err)
		return
	}
	fmt.Printf("Peer ID: %s\n", hex.EncodeToString(buf[48:]))

}
