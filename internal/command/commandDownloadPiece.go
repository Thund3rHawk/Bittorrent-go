package info

import (
	"fmt"
	"os"
	"strconv"
)

func CommandDownloadPiece() {
	args := os.Args[2:]
	var torrentFile, outputPath string
	if args[0] == "-o" {
		torrentFile = args[2]
		outputPath = args[1]
	} else {
		torrentFile = args[0]
		outputPath = "."
	}

	torrent, err := readTorrentFile(torrentFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	trackerRequest := makeTrackerRequest(torrent)

	peers, err := requestPeers(trackerRequest)
	if err != nil {
		fmt.Println(err)
		return
	}

	peerIp := fmt.Sprintf("%d.%d.%d.%d", peers.Peers[0], peers.Peers[1], peers.Peers[2], peers.Peers[3])
	// peerIp := CommandPeers(os.Args[4])
	peerPort := int(peers.Peers[4])<<8 | int(peers.Peers[5])
	peerPortStr := fmt.Sprintf("%d", peerPort)

	handshakeMsg := makeHandshakeMsg(handshake{
		length: byte(19),
		pstr:   "BitTorrent protocol",
		resv:   [8]byte{},
		info:   torrent.Info.hash(),
		peerId: []byte("00112233445566778899"),
	})
	conn, _, err := connectWithPeer(peerIp, peerPortStr, handshakeMsg)
	if err != nil {
		fmt.Println(err)
		return
	}

	ind, _ := strconv.Atoi(args[3])
	data := downloadFile(conn, torrent, ind)

	file, err := os.Create(outputPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Piece downloaded to %s.\n", outputPath)

}
