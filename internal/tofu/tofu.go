package tofu

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
)

var (
	ErrInvalidFingerprint = errors.New("Invalid fingerprint")
	ErrCertExpired        = errors.New("Certificate has expired")
	ErrCertNotYetValid    = errors.New("Certificate is not yet valid")
	ErrBadNumberOfCerts   = errors.New("Bad mumber of certs from server")
)

func HttpClientFromFile(server string, port int) (*http.Client, error) {
	fp, exists, err := ReadFingerprint()
	if err != nil {
		return nil, err
	}
	if !exists {
		fp, err = FetchFingerprint(server, port)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Got fingerprint from server: %s\n", fp)
		fmt.Printf("Verify that it is correct on server by 'whazza fingerprint'\n")
		WriteFingerprint(fp)
	}
	return HttpClient(fp), nil
}

func HttpClient(fingerprint string) *http.Client {
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
			return nil, ErrBadNumberOfCerts
		}
		cert := state.PeerCertificates[0]
		if fingerprint != getFingerprint(cert.Raw) {
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

func FetchFingerprint(server string, port int) (string, error) {
	addr := fmt.Sprintf("%s:%d", server, port)

	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", addr, config)
	if err != nil {
		return "", err
	}
	state := conn.ConnectionState()

	if len(state.PeerCertificates) != 1 {
		return "", ErrBadNumberOfCerts
	}
	cert := state.PeerCertificates[0]
	return getFingerprint(cert.Raw), nil
}

func ReadFingerprint() (string, bool, error) {
	dat, err := ioutil.ReadFile("cert.fingerprint")
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(dat), true, err
}

func WriteFingerprint(fp string) error {
	return ioutil.WriteFile("cert.fingerprint", []byte(fp), 0644)
}
