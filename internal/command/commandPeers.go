package info

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jackpal/bencode-go"
)

func CommandPeers() []string {
	baseUrl := GetInfo().url
	var peerArray []string
	infoHashBytes, _ := hex.DecodeString(GetInfo().infoHash)

	params := url.Values{}
	params.Add("info_hash", string(infoHashBytes))
	params.Add("peer_id", "00112233445566778899")
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", fmt.Sprint(GetInfo().length))
	params.Add("compact", "1")

	u, _ := url.ParseRequestURI(baseUrl)
	u.RawQuery = params.Encode()
	urlStr := fmt.Sprintf("%v", u)

	resp, _ := http.Get(urlStr)
	// if err != nil {
	// 	fmt.Println("Error getting data from the url")
	// 	return
	// }
	body, _ := io.ReadAll(resp.Body)
	reader := strings.NewReader(string(body))
	decoded, _ := bencode.Decode(reader)
	peerString, _ := decoded.(map[string]interface{})
	if peerData, found := peerString["peers"]; found {
		peers := ""
		peerBytes, _ := peerData.(string)
		// if !ok {
		// 	fmt.Println("Error: 'peers' field is not in the expected format")
		// 	return
		// }
		for k := 0; k < len(peerBytes); k += 6 {
			peers += strconv.Itoa(int(peerBytes[k])) + "." + strconv.Itoa(int(peerBytes[k+1])) + "." + strconv.Itoa(int(peerBytes[k+2])) + "." + strconv.Itoa(int(peerBytes[k+3])) + ":" + strconv.Itoa(int((binary.BigEndian.Uint16)([]byte(peerBytes[k+4:k+6])))) + "\n"
		}
		peerArray = append(peerArray, peers)
		fmt.Println(peers)
	} else {
		fmt.Println("Error: 'announce' field not found")
	}
	return peerArray
}
