package checker

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rymdhund/whazza/internal/base"
)

type CertChecker struct {
	Host            string `json:"host"`
	Port            int    `json:"port,omitempty"`
	ExpiresSoonDays int    `json:"expires_soon_days,omitempty"`
}

func (c CertChecker) Title() string {
	if c.portOrDefault() != 443 {
		return fmt.Sprintf("cert:%s:%d", c.Host, c.portOrDefault())
	}
	return fmt.Sprintf("cert:%s", c.Host)
}

func (c CertChecker) Type() string {
	return "cert"
}

func (c CertChecker) Validate() error {
	if c.Host == "" {
		return errors.New("Empty host in cert check")
	}
	return nil
}

func (c CertChecker) AsJson() []byte {
	b, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return b
}

func (c CertChecker) portOrDefault() int {
	if c.Port == 0 {
		return 443
	}
	return c.Port
}

func (c CertChecker) expiresSoonDaysOrDefault() int {
	if c.ExpiresSoonDays == 0 {
		return 20
	}
	return c.ExpiresSoonDays
}

func (c CertChecker) Run() base.Result {
	addr := fmt.Sprintf("%s:%d", c.Host, c.portOrDefault())
	conn, err := tls.Dial("tcp", addr, &tls.Config{})
	if err != nil {
		return base.FailResult(niceTlsError(err))
	}
	defer conn.Close()

	err = conn.VerifyHostname(c.Host)
	if err != nil {
		return base.FailResult(err.Error())
	}

	errorMsg := "No certificate found"
	for _, cert := range conn.ConnectionState().PeerCertificates {
		if !cert.IsCA {
			ok, msg := c.verifyExpiry(cert, time.Now())
			if ok {
				return base.GoodResult()
			}
			errorMsg = msg
		}
	}
	return base.FailResult(errorMsg)
}

func (c CertChecker) verifyExpiry(crt *x509.Certificate, now time.Time) (bool, string) {
	toExpiry := crt.NotAfter.Sub(now)
	if toExpiry < time.Duration(c.expiresSoonDaysOrDefault()*24)*time.Hour {
		return false, fmt.Sprintf("Cert expires in %d days", int(toExpiry.Hours()/24))
	}
	return true, ""
}

func niceTlsError(err error) string {
	switch e := err.(type) {
	case x509.UnknownAuthorityError:
		return "Cert not signed by trusted authority"
	case x509.CertificateInvalidError:
		switch e.Reason {
		case x509.Expired:
			return "Cert expired or not yet valid"
		}
	}
	return err.Error()
}
