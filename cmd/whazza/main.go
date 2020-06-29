package main

import (
	"fmt"
	"os"

	"github.com/rymdhund/whazza/internal/hubutil"
	"github.com/rymdhund/whazza/internal/persist"
)

func main() {
	args := os.Args

	if len(args) <= 1 {
		showUsage()
	} else if args[1] == "run" {
		startServer()
	} else if args[1] == "register" && len(args) == 4 {
		registerAgent(args[2], args[3])
	} else if args[1] == "fingerprint" && len(args) == 2 {
		showFingerprint()
	} else if args[1] == "show" && len(args) == 2 {
		show()
	} else {
		showUsage()
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Printf(`usage: %s <command> [<args>]
Where command is one of the following:
  run                           Start the server
  fingerprint                   Show the certificate fingerprint
  regiser <agent> <token hash>  Register the agent with a hashed token`, os.Args[0])
}

func show() {
	overviews, err := persist.GetCheckOverviews()
	if err != nil {
		panic(err)
	}
	for _, overview := range overviews {
		fmt.Println(overview.Show())
	}
}

func showFingerprint() {
	err := hubutil.InitCert()
	if err != nil {
		panic(err)
	}

	fp, err := hubutil.ReadCertFingerprint()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Cert fingerprint: %s\n", fp)
}

func registerAgent(name, tokenHash string) {
	err := persist.SetAgent(name, tokenHash)
	if err != nil {
		panic(err)
	}
}
