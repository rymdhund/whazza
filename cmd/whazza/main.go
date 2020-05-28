package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/rymdhund/whazza/internal/wserver"
)

func main() {
	key, err := readPrivateKey()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Private key: %s", key)

	wserver.StartServer()
}

func readPrivateKey() (envKey string, err error) {
	envKey = os.Getenv("WHAZZA_KEY")
	if envKey != "" {
		return
	}

	envFile := os.Getenv("WHAZZA_KEY_FILE")

	if envFile == "" {
		envFile = "./secret.key"
	}

	dat, err := ioutil.ReadFile(envFile)
	if err != nil {
		return
	}
	envKey = string(dat)
	return
}
