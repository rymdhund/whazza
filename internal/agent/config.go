package agent

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path"

	"github.com/rymdhund/whazza/internal/base"
	"github.com/rymdhund/whazza/internal/checking"
	"github.com/rymdhund/whazza/internal/tofu"
)

type Config struct {
	ServerHost            string `json:"server_host"`
	ServerPort            int    `json:"server_port"`
	ServerCertFingerprint string `json:"server_cert_fingerprint"`
	AgentName             string `json:"agent_name"`
	AgentToken            string `json:"agent_token"`
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

	err = os.MkdirAll(path.Dir(filename), 0755)
	if err != nil {
		return err
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
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = json.Unmarshal(contents, &cfg)
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

func normalizeCheck(chk *base.Check, defaultInterval int) error {
	if chk.CheckType == "" {
		return errors.New("Empty check type")
	}
	if chk.Interval <= 0 {
		return errors.New("Invalid interval")
	}
	if chk.Interval == 0 {
		chk.Interval = defaultInterval
	}
	meta, err := checking.GetCheckMeta(chk.CheckType)
	if err != nil {
		return err
	}
	_, err = meta.ParseParams(*chk)
	if err != nil {
		return err
	}
	if chk.Namespace == "" {
		chk.Namespace = meta.DefaultNamespace(*chk)
	}
	return nil
}

// ReadChecksConfig reads check configuration from a file
func ReadChecksConfig(filename string) ([]base.Check, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	type checkConfig struct {
		DefaultInterval int          `json:"default_interval"`
		Checks          []base.Check `json:"checks"`
	}

	var config checkConfig

	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&config)
	if err != nil {
		panic(err)
		return nil, err
	}

	defaultInterval := config.DefaultInterval
	if defaultInterval <= 0 {
		defaultInterval = 60
	}
	for i := range config.Checks {
		err = normalizeCheck(&config.Checks[i], defaultInterval)
		if err != nil {
			return nil, err
		}
	}
	return config.Checks, err
}
