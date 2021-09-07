package main

import (
	"fmt"
	"os"
	"path"

	"github.com/rymdhund/whazza/internal/hubutil"
	"github.com/rymdhund/whazza/internal/persist"
	"github.com/rymdhund/whazza/internal/sectoken"
	"github.com/rymdhund/whazza/internal/tofu"
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
	} else if args[1] == "register-external" && len(args) == 3 {
		initConf()

		token := sectoken.New()
		name := args[2]
		registerAgent(name, token.Hash())

		fingerprint, err := tofu.FingerprintOfCertFile(Config.CertFile())
		if err != nil {
			fmt.Printf("Couldn't get fingerprint: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated new agent\n")
		fmt.Printf("Name: %s\n", name)
		fmt.Printf("Token: %s\n", token)
		fmt.Printf("curl --insecure --pinnedpubkey 'sha256//%s' https://%s:%s@localhost:%d/agent/result\n", fingerprint.Encode(), name, token.Hash(), Config.Port)
	} else {
		showUsage()
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Printf(`usage: %s <command> [<args>]
Where command is one of the following:
  run                               Start the server
  fingerprint                       Show the certificate fingerprint
  register <agent> <token hash>     Register the agent with a hashed token
  register-external <name>          Register a new external agent and generate a token
  show                              Show status of checks
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

	fp, err := tofu.FingerprintOfCertFile(Config.CertFile())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Cert fingerprint: %s\n", fp.Encode())
}

func registerAgent(name, tokenHash string) {
	db, err := persist.Open(Config.Database())
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.SaveAgent(name, tokenHash)
	if err != nil {
		panic(err)
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
