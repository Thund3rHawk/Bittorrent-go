package info

import (
	"fmt"
	"os"
	"strings"
)

func CommandMagnetParse() {
	var magnetLink = os.Args[2]
	magnetParts := strings.Split(strings.TrimPrefix(magnetLink, "magnet:?"), "&")
	var trackerURL, infoHash string

	for _, part := range magnetParts {
		if strings.HasPrefix(part, "xt=urn:btih:") {
			infoHash = strings.TrimPrefix(part, "xt=urn:btih:")
		} else if strings.HasPrefix(part, "tr=") {
			trackerURL = strings.TrimPrefix(part, "tr=")
		}
	}

	// for _, part := range magnetParts {
	// 	if strings.HasPrefix(part, "tr=") {
	// 		trackerURL = strings.TrimPrefix(part, "tr=")
	// 	} else if strings.HasPrefix(part, "dn=") {
	// 		torrentFile = strings.TrimPrefix(part, "dn=")
	// 	}
	// }
	trackerURL = strings.ReplaceAll(trackerURL, "%3A", ":")
	trackerURL = strings.ReplaceAll(trackerURL, "%2F", "/")
	fmt.Println("Tracker URL:", trackerURL)
	fmt.Println("Info Hash:", infoHash)
	// fmt.Println("Torrent File:", torrentFile)
}
