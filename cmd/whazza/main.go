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
  run                            Start the server
  fingerprint                    Show the certificate fingerprint
  register <agent> <token hash>  Register the agent with a hashed token`, os.Args[0])
}

func show() {
	db, err := persist.Open(dbfile)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	overviews, err := db.GetCheckOverviews()
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
	db, err := persist.Open(dbfile)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	err = tx.SaveAgent(name, tokenHash)
	if err != nil {
		tx.Rollback()
		panic(err)
	} else {
		tx.Commit()
	}
}
