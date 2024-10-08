package infoHash

import (
	"crypto/sha1"
	"encoding/hex"
)

func Infohash(s string) string {
	hash := sha1.New()
	hash.Write([]byte(s))
	sha1_hash := hex.EncodeToString(hash.Sum(nil))
	return sha1_hash
}
