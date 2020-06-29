package agent

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/rymdhund/whazza/internal/base"
	"github.com/rymdhund/whazza/internal/tofu"
)

func Ping(cfg Config) error {
	client := tofu.HttpClient(cfg.ServerCertFingerprint)

	url := fmt.Sprintf("https://%s:%d/agent/ping", cfg.ServerHost, cfg.ServerPort)
	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(cfg.AgentName, cfg.AgentToken)
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if string(body) != "pong" {
			fmt.Errorf("Unexpected response: '%s'", body)
		}
		return nil
	} else if resp.StatusCode == http.StatusForbidden {
		hash := sha256.Sum256([]byte(cfg.AgentToken))

		hexified := make([][]byte, len(hash))
		for i, data := range hash {
			hexified[i] = []byte(fmt.Sprintf("%02X", data))
		}
		hashString := string(bytes.Join(hexified, nil))

		return fmt.Errorf("Not authorized. Run `whazza register %s %s` on the server to register this agent.", cfg.AgentName, hashString)
	} else {
		return fmt.Errorf("Invalid status: %d", resp.StatusCode)
	}
}

func SendCheckResult(cfg Config, msg base.CheckResultMsg) error {
	client := tofu.HttpClient(cfg.ServerCertFingerprint)

	url := fmt.Sprintf("https://%s:%d/agent/result", cfg.ServerHost, cfg.ServerPort)
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	req.SetBasicAuth(cfg.AgentName, cfg.AgentToken)
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return fmt.Errorf("Unexpected status: %d", resp.StatusCode)
	}
}
