package agent

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/rymdhund/whazza/internal/messages"
	"github.com/rymdhund/whazza/internal/tofu"
)

type HubConnection struct {
	client *http.Client
	cfg    Config
}

func NewHubConnection(cfg Config) *HubConnection {
	client := tofu.HttpClient(cfg.ServerCertFingerprint)
	return &HubConnection{
		client: client,
		cfg:    cfg,
	}
}

func (conn *HubConnection) request(method, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("https://%s:%d%s", conn.cfg.ServerHost, conn.cfg.ServerPort, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err) // we will only get err if url is malformed or invalid method
	}
	req.SetBasicAuth(conn.cfg.AgentName, conn.cfg.AgentToken)
	return conn.client.Do(req)
}

func (conn *HubConnection) Ping() error {
	resp, err := conn.request("GET", "/agent/ping", nil)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		if string(body) != "pong" {
			return fmt.Errorf("Unexpected response: '%s'", body)
		}
		return nil
	} else if resp.StatusCode == http.StatusForbidden {
		hash := sha256.Sum256([]byte(conn.cfg.AgentToken))

		hexified := make([][]byte, len(hash))
		for i, data := range hash {
			hexified[i] = []byte(fmt.Sprintf("%02X", data))
		}
		hashString := string(bytes.Join(hexified, nil))

		return fmt.Errorf("Not authorized. Run `whazza register %s %s` on the server to register this agent.", conn.cfg.AgentName, hashString)
	} else {
		return fmt.Errorf("Invalid status: %d", resp.StatusCode)
	}
}

func (conn *HubConnection) SendCheckResult(cfg Config, msg messages.CheckResultMsg) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	resp, err := conn.request("POST", "/agent/result", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status: %d", resp.StatusCode)
	}
	return nil
}
