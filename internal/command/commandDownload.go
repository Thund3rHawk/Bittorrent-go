package info

import (
	"fmt"
	"os"
)

func CommandDownload() {
	args := os.Args[2:]
	var torrentFile, outputFile string
	if args[0] == "-o" {
		torrentFile = args[2]
		outputFile = args[1]
	} else {
		torrentFile = args[0]
		outputFile = "sample.txt"
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
	peerPort := int(peers.Peers[4])<<8 | int(peers.Peers[5])
	peerPortStr := fmt.Sprintf("%d", peerPort)

	// hashes, _ := hex.DecodeString(string(torrent.Info.hash()))
	hashes := torrent.Info.getPiecesHashes()
	buf := make([]byte, torrent.Info.Length)
	for i := 0; i < len(hashes); i++ {
		conn, _, err := connectWithPeer(peerIp, peerPortStr, makeHandshakeMsg(handshake{
			length: byte(19),
			pstr:   "BitTorrent protocol",
			resv:   [8]byte{},
			info:   torrent.Info.hash(),
			peerId: []byte("00112233445566778899"),
		}))
		if err != nil {
			fmt.Println(err)
			return
		}

		data := downloadPiece(conn, torrent, i)
		offset := i * torrent.Info.PieceLength
		copy(buf[offset:offset+len(data)], data)
	}

	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	_, err = file.Write(buf)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Downloaded %s to %s.\n", torrentFile, outputFile)
}
