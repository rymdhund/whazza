package tofu

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"
)

var (
	ErrInvalidFingerprint = errors.New("Invalid fingerprint")
	ErrCertExpired        = errors.New("Certificate has expired")
	ErrCertNotYetValid    = errors.New("Certificate is not yet valid")
	ErrBadNumberOfCerts   = errors.New("Bad number of certs from server")
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
			return nil, fmt.Errorf("Incorrect server fingerprint. Expected: \"%s\", Got: \"%s\"", fingerprint.Encode(), certFp.Encode())
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

func GenerateCert(keyFile, certFile string) error {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * 10 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("Failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"whazza"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, priv.Public(), priv)
	if err != nil {
		return fmt.Errorf("Failed to create certificate: %v", err)
	}

	certOut, err := os.OpenFile(certFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("Failed to open cert.pem for writing: %v", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("Failed to write data to cert.pem: %v", err)
	}
	if err := certOut.Close(); err != nil {
		return fmt.Errorf("Error closing cert.pem: %v", err)
	}

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("Failed to open key.pem for writing: %v", err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("Failed to write data to key.pem: %v", err)
	}
	if err := keyOut.Close(); err != nil {
		return fmt.Errorf("Error closing key.pem: %v", err)
	}

	return nil
}
