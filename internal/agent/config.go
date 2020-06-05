package agent

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"

	"github.com/rymdhund/whazza/internal/tofu"
)

type Config struct {
	ServerHost            string
	ServerPort            int
	ServerCertFingerprint string
	AgentName             string
	AgentToken            string
}

func GenerateConfig(agentName string, serverHost string, serverPort int, serverFingerprint string) (Config, error) {
	var err error

	if serverFingerprint == "" {
		serverFingerprint, err = tofu.FetchFingerprint(serverHost, serverPort)
		if err != nil {
			return Config{}, errors.New("Could not connect to server, please manually provide a certificate fingerprint")
		}
		fmt.Printf("Fetched fingerprint from server: %s\n", serverFingerprint)
		fmt.Printf("Verify that it is correct on server by 'whazza fingerprint'\n")
	}

	agentToken, err := generateToken()
	if err != nil {
		return Config{}, err
	}

	return Config{
		ServerHost:            serverHost,
		ServerPort:            serverPort,
		ServerCertFingerprint: serverFingerprint,
		AgentName:             agentName,
		AgentToken:            agentToken,
	}, nil
}

func SaveConfig(cfg Config, filename string) error {
	var _, err = os.Stat(filename)
	if !os.IsNotExist(err) {
		return errors.New("Config file already exists")
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(cfg)
	if err != nil {
		return err
	}

	return nil
}

func ReadConfig(filename string) (Config, error) {
	byteValue, err := ioutil.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = json.Unmarshal(byteValue, &cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func generateToken() (string, error) {
	var symbols = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	const n = 64

	key := make([]rune, n)
	r := rand.Reader
	sl := big.NewInt(int64(len(symbols)))
	for i := range key {
		v, err := rand.Int(r, sl)
		if err != nil {
			return "", err
		}
		key[i] = symbols[v.Int64()]
	}
	return string(key), nil
}
