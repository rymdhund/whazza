package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rymdhund/whazza/internal/agent"
)

const cfgFile = "config.json"

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
		cfg := readConf()
		agent.Ping(cfg)
	} else if args[1] == "run" {
		cfg := readConf()
		agent.Run(cfg)
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

func readConf() agent.Config {
	cfg, err := agent.ReadConfig(cfgFile)
	if err != nil {
		fmt.Printf("Couldn't read config file: %s", err)
		os.Exit(1)
	}
	return cfg
}
