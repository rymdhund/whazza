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
	} else if args[1] == "fingerprint" && len(args) == 2 {
		wserver.ShowFingerprint()
	} else {
		ShowUsage()
		os.Exit(1)
	}
}

func ShowUsage() {
	fmt.Printf(`usage: %s <command> [<args>]
Where command is one of the following:
  run          Start the server
  fingerprint  Show the certificate fingerprint`, os.Args[0])
}
