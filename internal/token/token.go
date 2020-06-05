package token

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

type Token string

func (t Token) Hash() string {
	hash := sha256.Sum256([]byte(t))

	hexified := make([][]byte, len(hash))
	for i, data := range hash {
		hexified[i] = []byte(fmt.Sprintf("%02X", data))
	}
	return string(bytes.Join(hexified, nil))
}
