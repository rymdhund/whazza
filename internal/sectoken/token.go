package sectoken

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
)

type SecToken string

func New() SecToken {
	var symbols = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	const n = 64

	key := make([]rune, n)
	r := rand.Reader
	sl := big.NewInt(int64(len(symbols)))
	for i := range key {
		v, err := rand.Int(r, sl)
		if err != nil {
			panic(err)
		}
		key[i] = symbols[v.Int64()]
	}
	return SecToken(string(key))
}

func (t SecToken) Hash() string {
	hash := sha256.Sum256([]byte(t))

	hexified := make([][]byte, len(hash))
	for i, data := range hash {
		hexified[i] = []byte(fmt.Sprintf("%02X", data))
	}
	return string(bytes.Join(hexified, nil))
}

func (t SecToken) String() string {
	return string(t)
}
