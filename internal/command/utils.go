package info

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/jackpal/bencode-go"
)

type Torrent struct {
	Announce string `bencode:"announce"`
	Info     Info   `bencode:"info"`
}

type Info struct {
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

type trackerRequest struct {
	URL        string
	InfoHash   string
	PeerID     string
	Port       int
	Uploaded   int
	Downloaded int
	Left       int
	Compact    int
}

type trackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

type handshake struct {
	length byte
	pstr   string
	resv   [8]byte
	info   []byte
	peerId []byte
}

type PeerMessage struct {
	lengthPrefix uint32
	id           uint8
	index        uint32
	begin        uint32
	length       uint32
}

func decodeBencode(bencodedString string) (interface{}, int, error) {
	switch bencodedString[0] {
	case 'i':
		return decodeInt(bencodedString, 0)
	case 'l':
		return decodeList(bencodedString, 0)
	case 'd':
		return decodeDict(bencodedString, 0)
	default:
		if unicode.IsDigit(rune(bencodedString[0])) {
			return decodeString(bencodedString, 0)
		} else {
			return "", -1, fmt.Errorf("invalid bencoded string")
		}
	}
}

func decodeString(bencodedString string, pos int) (string, int, error) {
	firstColonIndex := strings.Index(bencodedString[pos:], ":") + pos
	lengthStr := bencodedString[pos:firstColonIndex]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", 0, err
	}
	return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], firstColonIndex + length, nil
}

func decodeInt(bencodedString string, pos int) (int, int, error) {
	for i := pos; i < len(bencodedString); i++ {
		if bencodedString[i] == 'e' {
			decodedInt, err := strconv.Atoi(bencodedString[pos+1 : i])
			if err != nil {
				return 0, 0, err
			}
			return decodedInt, i, nil
		}
	}
	return 0, 0, fmt.Errorf("invalid bencoded string")
}

func decodeList(bencodedString string, pos int) ([]interface{}, int, error) {
	list := []interface{}{}
	end := 0
	for i := pos + 1; i < len(bencodedString); i++ {
		ch := bencodedString[i]
		if ch == 'e' {
			end = i
			break
		} else if ch == 'i' {
			decodedInt, endPos, err := decodeInt(bencodedString, i)
			if err != nil {
				return nil, -1, err
			}
			list = append(list, decodedInt)
			i = endPos
		} else if ch == 'l' {
			decodedList, endPos, err := decodeList(bencodedString, i)
			if err != nil {
				return nil, -1, err
			}
			list = append(list, decodedList)
			i = endPos
		} else if unicode.IsDigit(rune(ch)) {
			decodedString, endPos, err := decodeString(bencodedString, i)
			if err != nil {
				return nil, -1, err
			}
			list = append(list, decodedString)
			i = endPos
		}
	}
	return list, end, nil
}

func decodeDict(bencodedString string, pos int) (map[string]interface{}, int, error) {
	dict := make(map[string]interface{})
	key := ""
	end := 0
	for i := pos + 1; i < len(bencodedString); i++ {
		ch := bencodedString[i]
		if ch == 'e' {
			end = i
			break
		} else if ch == 'i' {
			decodedInt, endPos, err := decodeInt(bencodedString, i)
			if err != nil {
				return nil, -1, err
			}
			i = endPos
			if key == "" {
				key = strconv.Itoa(decodedInt)
			} else {
				dict[key] = decodedInt
				key = ""
			}
		} else if unicode.IsDigit(rune(ch)) {
			decodedString, endPos, err := decodeString(bencodedString, i)
			if err != nil {
				return nil, -1, err
			}
			i = endPos
			if key == "" {
				key = decodedString
			} else {
				dict[key] = decodedString
				key = ""
			}
		} else if ch == 'l' {
			decodedList, endPos, err := decodeList(bencodedString, i)
			if err != nil {
				return nil, -1, err
			}
			i = endPos
			if key == "" {
				key = "list"
			} else {
				dict[key] = decodedList
				key = ""
			}
		} else if ch == 'd' {
			decodedDict, endPos, err := decodeDict(bencodedString, i)
			if err != nil {
				return nil, -1, err
			}
			i = endPos
			if key == "" {
				key = "dict"
			} else {
				dict[key] = decodedDict
				key = ""
			}
		} else {
			return nil, -1, fmt.Errorf("invalid bencoded string")
		}
	}
	return dict, end, nil
}

func encodeBencode(data interface{}) (string, error) {
	switch v := data.(type) {
	case string:
		return EncodeString(v)
	case int:
		return EncodeNumber(v)
	case []interface{}:
		return EncodeList(v)
	case map[string]interface{}:
		return EncodeDictionary(v)
	default:
		return "", fmt.Errorf("unsupported type: %T", v)
	}
}

func EncodeString(data string) (string, error) {
	return strconv.Itoa(len(data)) + ":" + data, nil
}

func EncodeNumber(data int) (string, error) {
	return "i" + strconv.Itoa(data) + "e", nil
}

func EncodeList(data []interface{}) (string, error) {
	encodedList := "l"
	for _, v := range data {
		encoded, err := encodeBencode(v)
		if err != nil {
			return "", err
		}
		encodedList += encoded
	}
	encodedList += "e"
	return encodedList, nil
}

func EncodeDictionary(data map[string]interface{}) (string, error) {
	encodedDict := "d"
	for k, v := range data {
		encodedKey, err := EncodeString(k)
		if err != nil {
			return "", err
		}
		encodedDict += encodedKey
		encodedValue, err := encodeBencode(v)
		if err != nil {
			return "", err
		}
		encodedDict += encodedValue
	}
	encodedDict += "e"
	return encodedDict, nil
}

func readTorrentFile(torrentFile string) (Torrent, error) {
	file, err := os.Open(torrentFile)
	if err != nil {
		return Torrent{}, err
	}
	defer file.Close()

	var torrent Torrent
	err = bencode.Unmarshal(file, &torrent)
	if err != nil {
		return Torrent{}, err
	}
	return torrent, nil
}

func (info Info) hash() []byte {
	var b bytes.Buffer
	bencode.Marshal(&b, info)
	sha := sha1.Sum(b.Bytes())
	return sha[:]
}

func makeTrackerRequest(torrent Torrent) trackerRequest {
	return trackerRequest{
		URL:        torrent.Announce,
		InfoHash:   string(torrent.Info.hash()),
		PeerID:     "00112233445566778899",
		Port:       6881,
		Uploaded:   0,
		Downloaded: 0,
		Left:       torrent.Info.Length,
		Compact:    1,
	}
}

func requestPeers(req trackerRequest) (trackerResponse, error) {
	client := &http.Client{}
	url, err := url.Parse(req.URL)
	if err != nil {
		return trackerResponse{}, err
	}
	q := url.Query()
	q.Add("info_hash", req.InfoHash)
	q.Add("peer_id", req.PeerID)
	q.Add("port", strconv.Itoa(req.Port))
	q.Add("uploaded", strconv.Itoa(req.Uploaded))
	q.Add("downloaded", strconv.Itoa(req.Downloaded))
	q.Add("left", strconv.Itoa(req.Left))
	q.Add("compact", strconv.Itoa(req.Compact))
	url.RawQuery = q.Encode()

	resp, err := client.Get(url.String())
	if err != nil {
		return trackerResponse{}, err
	}
	defer resp.Body.Close()

	var trackerResponse trackerResponse
	err = bencode.Unmarshal(resp.Body, &trackerResponse)
	if err != nil {
		return trackerResponse, err
	}

	return trackerResponse, nil
}

func makeHandshakeMsg(hadnshake handshake) []byte {
	var msg []byte
	msg = append(msg, hadnshake.length)
	msg = append(msg, hadnshake.pstr...)
	msg = append(msg, hadnshake.resv[:]...)
	msg = append(msg, hadnshake.info[:]...)
	msg = append(msg, hadnshake.peerId[:]...)
	return msg
}

func connectWithPeer(peerIp string, peerPort string, msg []byte) (net.Conn, handshake, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", peerIp, peerPort))
	if err != nil {
		return nil, handshake{}, err
	}

	_, err = conn.Write(msg)
	if err != nil {
		return nil, handshake{}, err
	}

	// Read the handshake response
	resp := make([]byte, 68)
	_, err = conn.Read(resp)
	if err != nil {
		return nil, handshake{}, err
	}

	// fmt.Println("Handshake response:", string(resp))
	return conn, handshake{
		length: resp[0],
		pstr:   string(resp[1:20]),
		resv:   [8]byte{},
		info:   resp[28:48],
		peerId: resp[48:68],
	}, nil
}

func downloadFile(conn net.Conn, torrent Torrent, index int) []byte {
	defer conn.Close()

	// wait for the bitfield message (id = 5)
	buf := make([]byte, 4)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	peerMessage := PeerMessage{}
	peerMessage.lengthPrefix = binary.BigEndian.Uint32(buf)
	payloadBuf := make([]byte, peerMessage.lengthPrefix)
	_, err = conn.Read(payloadBuf)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	peerMessage.id = payloadBuf[0]

	fmt.Printf("Received message: %v\n", peerMessage)
	if peerMessage.id != 5 {
		fmt.Println("Expected bitfield message")
		return nil
	}

	// send interested message (id = 2)
	_, err = conn.Write([]byte{0, 0, 0, 1, 2})
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// wait for unchoke message (id = 1)
	buf = make([]byte, 4)
	_, err = conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	peerMessage = PeerMessage{}
	peerMessage.lengthPrefix = binary.BigEndian.Uint32(buf)
	payloadBuf = make([]byte, peerMessage.lengthPrefix)
	_, err = conn.Read(payloadBuf)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	peerMessage.id = payloadBuf[0]

	fmt.Printf("Received message: %v\n", peerMessage)
	if peerMessage.id != 1 {
		fmt.Println(buf)
		fmt.Println("Expected unchoke message")
		return nil
	}

	// send request message (id = 6) for each block
	// Break the piece into blocks of 16 kiB (16 * 1024 bytes) and send a request message for each block
	pieceSize := torrent.Info.PieceLength
	pieceCnt := int(math.Ceil(float64(torrent.Info.Length) / float64(pieceSize)))
	if index == pieceCnt-1 {
		pieceSize = torrent.Info.Length % torrent.Info.PieceLength
	}
	blockSize := 16 * 1024
	blockCnt := int(math.Ceil(float64(pieceSize) / float64(blockSize)))
	fmt.Printf("File Length: %d, Piece Length: %d, Piece Count: %d, Block Size: %d, Block Count: %d\n", torrent.Info.Length, torrent.Info.PieceLength, pieceCnt, blockSize, blockCnt)
	var data []byte
	for i := 0; i < blockCnt; i++ {
		blockLength := blockSize
		if i == blockCnt-1 {
			blockLength = pieceSize - ((blockCnt - 1) * int(blockSize))
		}
		peerMessage := PeerMessage{
			lengthPrefix: 13,
			id:           6,
			index:        uint32(index),
			begin:        uint32(i * int(blockSize)),
			length:       uint32(blockLength),
		}

		var buf bytes.Buffer
		binary.Write(&buf, binary.BigEndian, peerMessage)
		_, err = conn.Write(buf.Bytes())
		if err != nil {
			fmt.Println(err)
			return nil
		}
		fmt.Println("Sent request message", peerMessage)

		// wait for piece message (id = 7)
		resBuf := make([]byte, 4)
		_, err = conn.Read(resBuf)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		peerMessage = PeerMessage{}
		peerMessage.lengthPrefix = binary.BigEndian.Uint32(resBuf)
		payloadBuf := make([]byte, peerMessage.lengthPrefix)
		_, err = io.ReadFull(conn, payloadBuf)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		peerMessage.id = payloadBuf[0]
		fmt.Printf("Received message: %v\n", peerMessage)

		data = append(data, payloadBuf[9:]...)
	}

	return data
}
