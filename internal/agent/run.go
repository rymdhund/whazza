package agent

import (
	"fmt"
	"log"
	"os"

	"github.com/rymdhund/whazza/internal/agent/checking"
	"github.com/rymdhund/whazza/internal/base"
)

func Run(cfg Config) {
	checks, err := ReadChecksConfig("checks.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading checks.json file: %s\n", err)
		os.Exit(1)
	}

	for _, c := range checks {
		meta, _ := checking.GetCheckMeta(c)
		res := meta.DoCheck(c)
		checkResult := base.CheckResultMsg{Check: c, Result: res}
		err = send(cfg, checkResult)
		if err != nil {
			log.Printf("Error: couldn't send result: %s", err)
		}
	}
}
