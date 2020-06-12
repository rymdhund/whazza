package main

import (
	"fmt"
	"os"

	"github.com/rymdhund/whazza/internal/wserver"
)

func main() {
	args := os.Args

	if len(args) <= 1 {
		ShowUsage()
	} else if args[1] == "run" {
		wserver.StartServer()
	} else if args[1] == "register" && len(args) == 4 {
		wserver.RegisterAgent(args[2], args[3])
	} else if args[1] == "fingerprint" && len(args) == 2 {
		wserver.ShowFingerprint()
	} else if args[1] == "show" && len(args) == 2 {
		wserver.Show()
	} else {
		ShowUsage()
		os.Exit(1)
	}
}

func ShowUsage() {
	fmt.Printf(`usage: %s <command> [<args>]
Where command is one of the following:
  run                           Start the server
  fingerprint                   Show the certificate fingerprint
  regiser <agent> <token hash>  Register the agent with a hashed token`, os.Args[0])
}
