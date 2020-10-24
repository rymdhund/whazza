package tofu

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

var (
	ErrInvalidFingerprint = errors.New("Invalid fingerprint")
	ErrCertExpired        = errors.New("Certificate has expired")
	ErrCertNotYetValid    = errors.New("Certificate is not yet valid")
	ErrBadNumberOfCerts   = errors.New("Bad mumber of certs from server")
)

func HttpClient(fingerprint string) *http.Client {
	fpBytes := fingerprintBytes(fingerprint)
	dial := func(network, addr string) (net.Conn, error) {
		config := &tls.Config{
			InsecureSkipVerify: true,
		}

		conn, err := tls.Dial(network, addr, config)
		if err != nil {
			return nil, err

		}
		state := conn.ConnectionState()
		now := time.Now()
		if len(state.PeerCertificates) != 1 {
			conn.Close()
			return nil, ErrBadNumberOfCerts
		}
		cert := state.PeerCertificates[0]
		if !compareFingerprint(cert.Raw, fpBytes) {
			conn.Close()
			fmt.Printf("expected: %s\n", fingerprint)
			fmt.Printf("got: %s\n", getFingerprint(cert.Raw))
			return nil, ErrInvalidFingerprint
		}
		if now.Before(cert.NotBefore) {
			conn.Close()
			return nil, ErrCertNotYetValid
		}
		if now.After(cert.NotAfter) {
			conn.Close()
			return nil, ErrCertExpired
		}
		// we're good
		return conn, nil
	}
	return &http.Client{
		Transport: &http.Transport{
			DialTLS: dial,
		},
	}
}

func getFingerprint(der []byte) string {
	hash := sha256.Sum256(der)
	hexified := make([][]byte, len(hash))
	for i, data := range hash {
		hexified[i] = []byte(fmt.Sprintf("%02X", data))
	}
	return string(bytes.Join(hexified, []byte(":")))
}

func fingerprintBytes(fp string) []byte {
	res, err := hex.DecodeString(strings.ReplaceAll(fp, ":", ""))
	if err != nil {
		panic(err)
	}
	return res
}

func compareFingerprint(der []byte, fp []byte) bool {
	hash := sha256.Sum256(der)
	return bytes.Compare(hash[:], fp) == 0
}

func FetchFingerprint(server string, port int) (string, error) {
	addr := fmt.Sprintf("%s:%d", server, port)

	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", addr, config)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	state := conn.ConnectionState()

	if len(state.PeerCertificates) != 1 {
		return "", ErrBadNumberOfCerts
	}
	cert := state.PeerCertificates[0]
	return getFingerprint(cert.Raw), nil
}
