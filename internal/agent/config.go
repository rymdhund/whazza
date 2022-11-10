package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/rymdhund/whazza/internal/chk"
	"github.com/rymdhund/whazza/internal/sectoken"
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
	var fingerprint tofu.Fingerprint

	if serverFingerprint != "" {
		fingerprint, err = tofu.FingerprintOfString(serverFingerprint)
		if err != nil {
			return Config{}, fmt.Errorf("Invalid fingerprint %w", err)
		}
	} else {
		fingerprint, err = tofu.FingerprintOfServer(serverHost, serverPort)
		if err != nil {
			return Config{}, errors.New("Could not connect to server, please manually provide a certificate fingerprint")
		}
		fmt.Printf("Fetched fingerprint from server: '%s'\n", fingerprint)
		fmt.Printf("Verify that it is correct on server by 'whazza fingerprint'\n")
	}

	agentToken := sectoken.New()
	if err != nil {
		return Config{}, err
	}

	return Config{
		ServerHost:            serverHost,
		ServerPort:            serverPort,
		ServerCertFingerprint: fingerprint.Encode(),
		AgentName:             agentName,
		AgentToken:            agentToken.String(),
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

func ParseChecksConfig(input io.Reader) ([]chk.Check, error) {
	type checkConfig struct {
		DefaultInterval int         `json:"default_interval"`
		Checks          []chk.Check `json:"checks"`
	}

	var config checkConfig

	decoder := json.NewDecoder(input)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return config.Checks, nil
}
