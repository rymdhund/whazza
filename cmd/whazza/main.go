package main

import (
	"fmt"
	"os"
	"path"

	"github.com/rymdhund/whazza/internal/hubutil"
	"github.com/rymdhund/whazza/internal/persist"
)

// Config is  the global configuration object. Initialized by initConf() function
var Config hubutil.HubConfig

func main() {
	args := os.Args

	if len(args) <= 1 {
		showUsage()
	} else if args[1] == "run" {
		initConf()
		startServer()
	} else if args[1] == "register" && len(args) == 4 {
		initConf()
		registerAgent(args[2], args[3])
	} else if args[1] == "fingerprint" && len(args) == 2 {
		initConf()
		showFingerprint()
	} else if args[1] == "show" && len(args) == 2 {
		initConf()
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
  register <agent> <token hash>  Register the agent with a hashed token
  show                           Show status of checks
`, os.Args[0])
}

func show() {
	db, err := persist.Open(Config.Database())
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
	err := hubutil.InitCert(Config.KeyFile(), Config.CertFile())
	if err != nil {
		panic(err)
	}

	fp, err := hubutil.ReadCertFingerprint(Config.CertFile())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Cert fingerprint: %s\n", fp)
}

func registerAgent(name, tokenHash string) {
	db, err := persist.Open(Config.Database())
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

func initConf() {
	cfgFile := os.Getenv("WHAZZA_CONFIG_FILE")
	if cfgFile == "" {
		cfgFile = path.Join(os.Getenv("HOME"), ".whazza", "hub.json")
	}

	cfg, err := hubutil.ReadConfig(cfgFile)
	if err != nil {
		fmt.Printf("Couldn't read config file %s\n", cfgFile)
		os.Exit(1)
	}

	Config = cfg
}
