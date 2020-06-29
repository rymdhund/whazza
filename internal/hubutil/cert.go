package hubutil

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"time"
)

func InitCert() error {
	_, err := os.Stat("key.pem")
	if !os.IsNotExist(err) {
		// key exists
		return nil
	}

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * 10 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
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

	certOut, err := os.OpenFile("cert.pem", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("Failed to open cert.pem for writing: %v", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("Failed to write data to cert.pem: %v", err)
	}
	if err := certOut.Close(); err != nil {
		return fmt.Errorf("Error closing cert.pem: %v", err)
	}
	log.Print("wrote cert.pem\n")

	keyOut, err := os.OpenFile("key.pem", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
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
	log.Print("wrote key.pem\n")

	return nil
}

func ReadCertFingerprint() (string, error) {
	dat, err := ioutil.ReadFile("cert.pem")
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(dat)
	if block == nil || block.Type != "CERTIFICATE" {
		fmt.Printf("block: %+v\ntype: %+v\n", block, block.Type)
		return "", errors.New("Invalid cert")
	}

	hash := sha256.Sum256(block.Bytes)
	hexified := make([][]byte, len(hash))
	for i, data := range hash {
		hexified[i] = []byte(fmt.Sprintf("%02X", data))
	}
	return string(bytes.Join(hexified, []byte(":"))), nil
}
