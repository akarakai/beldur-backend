package campaign

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

func generateAccessCode() string {
	const size = 3
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return strings.ToUpper(hex.EncodeToString(b))
}
