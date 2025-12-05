package telegram

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"time"
)

// GenerateCallbackHash generates unique hash for callback data
func GenerateCallbackHash(index int) string {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(time.Now().UnixNano())^uint64(index))

	hash := sha1.New()
	hash.Write(buf)
	return hex.EncodeToString(hash.Sum(nil))
}
