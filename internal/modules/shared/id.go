package shared

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateConnID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
