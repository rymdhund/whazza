package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/rymdhund/whazza/internal/agent"
	"github.com/rymdhund/whazza/internal/tofu"
)

func main() {
	args := os.Args

	cfgFile := "config.json"

	if len(args) <= 1 {
		ShowUsage()
	} else if args[1] == "init" {
		initArgs := args[2:]
		if len(initArgs) >= 3 && len(initArgs) <= 4 {
			fp := ""
			if len(initArgs) == 4 {
				fp = initArgs[3]
			}
			port, err := strconv.Atoi(initArgs[2])
			if err != nil {
				fmt.Printf("Invalid port number: '%s'", initArgs[2])
				os.Exit(1)
			}
			cfg, err := agent.GenerateConfig(initArgs[0], initArgs[1], port, fp)
			if err != nil {
				fmt.Printf("Error generating config: %s\n", err)
				os.Exit(1)
			}
			err = agent.SaveConfig(cfg, cfgFile)
			if err != nil {
				fmt.Printf("Error saving config: %s\n", err)
				os.Exit(1)
			}
			fmt.Printf("Successfully created %s\n", cfgFile)
		} else {
			ShowUsage()
			os.Exit(1)
		}
	} else if args[1] == "ping" {
		cfg, err := agent.ReadConfig(cfgFile)
		if err != nil {
			fmt.Printf("Couldn't read config file: %s", err)
			os.Exit(1)
		}
		err = ping(cfg)
		if err != nil {
			fmt.Printf("Error pinging server: %s\n", err)
			os.Exit(1)
		} else {
			fmt.Println("Server: Ok")
		}
	} else {
		ShowUsage()
		os.Exit(1)
	}

}

func ShowUsage() {
	fmt.Printf(`usage: %s <command> [<args>]
Where command is one of the following:
  init <agentname> <serverhost> <serverport> [server cert fingerprint]   Create config file
  ping                                                                   Ping the configured whazza server
`, os.Args[0])
}

func ping(cfg agent.Config) error {
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
