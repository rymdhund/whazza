package tofu

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

var (
	ErrInvalidFingerprint = errors.New("Invalid fingerprint")
	ErrCertExpired        = errors.New("Certificate has expired")
	ErrCertNotYetValid    = errors.New("Certificate is not yet valid")
	ErrBadNumberOfCerts   = errors.New("Bad mumber of certs from server")
)

func HttpClient(fingerprint Fingerprint) *http.Client {
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
		certFp, err := FingerprintOfCert(cert)
		if err != nil {
			return nil, err
		}

		if !fingerprint.Matches(certFp) {
			conn.Close()
			fmt.Printf("expected: %s\n", fingerprint.Encode())
			fmt.Printf("got: %s\n", certFp.Encode())
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

type Fingerprint []byte

func (fp Fingerprint) Encode() string {
	return base64.StdEncoding.EncodeToString(fp)
}

func (fp Fingerprint) Matches(fp2 Fingerprint) bool {
	return bytes.Compare(fp, fp2) == 0
}

func FingerprintOfServer(server string, port int) (Fingerprint, error) {
	addr := fmt.Sprintf("%s:%d", server, port)

	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", addr, config)
	if err != nil {
		return Fingerprint{}, err
	}
	defer conn.Close()
	state := conn.ConnectionState()

	if len(state.PeerCertificates) != 1 {
		return Fingerprint{}, ErrBadNumberOfCerts
	}

	cert := state.PeerCertificates[0]
	return FingerprintOfCert(cert)
}

func FingerprintOfCertFile(certFile string) (Fingerprint, error) {
	dat, err := ioutil.ReadFile(certFile)
	if err != nil {
		return Fingerprint{}, err
	}

	block, _ := pem.Decode(dat)
	if block == nil || block.Type != "CERTIFICATE" {
		fmt.Printf("block: %+v\ntype: %+v\n", block, block.Type)
		return Fingerprint{}, errors.New("Invalid cert")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return Fingerprint{}, err
	}

	return FingerprintOfCert(cert)
}

func FingerprintOfCert(cert *x509.Certificate) (Fingerprint, error) {
	pubKey, ok := cert.PublicKey.(ed25519.PublicKey)
	if !ok {
		return Fingerprint{}, errors.New("Invalid cert type, expected ed25519")
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return Fingerprint{}, err
	}

	hash := sha256.Sum256(pubBytes)
	return Fingerprint(hash[:]), nil
}

func FingerprintOfString(fp string) (Fingerprint, error) {
	bytes, err := base64.StdEncoding.DecodeString(fp)
	if err != nil {
		return Fingerprint{}, err
	}
	return Fingerprint(bytes), nil
}
