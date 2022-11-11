package hubutil

import (
	"os"

	. "github.com/rymdhund/whazza/internal/logging"
	"github.com/rymdhund/whazza/internal/tofu"
)

func InitCert(keyFile, certFile string) error {
	_, err := os.Stat(keyFile)
	if !os.IsNotExist(err) {
		// key exists
		return nil
	}

	if err := tofu.GenerateCert(keyFile, certFile); err != nil {
		return err
	}

	InfoLog.Printf("Generated %s and %s", keyFile, certFile)
	return nil
}
