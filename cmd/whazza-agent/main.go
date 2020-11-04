package main

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/rymdhund/whazza/internal/agent"
)

func main() {
	args := os.Args

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
			err = agent.SaveConfig(cfg, configFile())
			if err != nil {
				fmt.Printf("Error saving config: %s\n", err)
				os.Exit(1)
			}
			fmt.Printf("Successfully created %s\n", configFile())
		} else {
			ShowUsage()
			os.Exit(1)
		}
	} else if args[1] == "ping" {
		ping()
	} else if args[1] == "run" {
		run()
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
  run 																	 Run the agent continously
`, os.Args[0])
}

func configFile() string {
	dir := os.Getenv("WHAZZA_AGENT_DIR")
	if dir == "" {
		dir = path.Join(os.Getenv("HOME"), ".whazza-agent")
	}
	return path.Join(dir, "config.json")
}

func checksConfigFile() string {
	dir := os.Getenv("WHAZZA_AGENT_DIR")
	if dir == "" {
		dir = path.Join(os.Getenv("HOME"), ".whazza-agent")
	}
	return path.Join(dir, "checks.json")
}

func readConf() agent.Config {
	file := configFile()
	cfg, err := agent.ReadConfig(file)
	if err != nil {
		fmt.Printf("Couldn't read config file: %s (%s)\n", file, err)
		os.Exit(1)
	}
	return cfg
}

func ping() {
	cfg := readConf()
	hubConn, err := agent.NewHubConnection(cfg)
	if err != nil {
		fmt.Printf("Error connecting to server: %s\n", err)
		os.Exit(1)
	}
	err = hubConn.Ping()
	if err != nil {
		fmt.Printf("Error pinging server: %s\n", err)
		os.Exit(1)
	} else {
		fmt.Println("Server: Ok")
	}
}
