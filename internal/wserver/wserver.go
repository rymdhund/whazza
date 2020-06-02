package wserver

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/rymdhund/whazza/internal/agent"
	"github.com/rymdhund/whazza/internal/base"
	serverdb "github.com/rymdhund/whazza/internal/server_db"
)

func generateCertIfNotExists() error {
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
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	/*
		hosts := strings.Split(*host, ",")
		for _, h := range hosts {
			if ip := net.ParseIP(h); ip != nil {
				template.IPAddresses = append(template.IPAddresses, ip)
			} else {
				template.DNSNames = append(template.DNSNames, h)
			}
		}
	*/

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

func StartServer() {
	err := generateCertIfNotExists()
	if err != nil {
		panic(err)
	}

	serverdb.Init()

	c := base.Check{CheckType: "http-up", Namespace: "net:google.com", CheckParams: agent.HttpCheckParams{"google.com", 80, nil}, Interval: 900}

	res := agent.DoCheck(c)
	checkResult := base.CheckResultMsg{Check: c, Result: res}
	handleMessage(checkResult)

	overview, err := serverdb.GetCheckOverview(c)
	if err != nil {
		panic(err)
	}
	fmt.Printf("status: %s\n", overview.Show())

	http.HandleFunc("/", notFoundHandler)
	http.HandleFunc("/agent/ping", pingHandler)
	http.HandleFunc("/agent/result", resultHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleMessage(msg interface{}) {
	fmt.Printf("Got message: %+v\n", msg)

	switch msg := msg.(type) {
	default:
		panic(fmt.Sprintf("unexpected type %T", msg))
	case base.CheckResultMsg:
		err := serverdb.AddResult(msg.Result, msg.Check)
		if err != nil {
			panic(err)
		}
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "404 Not found", http.StatusNotFound)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		log.Print("Got ping")
	default:
		http.Error(w, "405 Method Not allowed", http.StatusMethodNotAllowed)
	}
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		decoder := json.NewDecoder(r.Body)
		var checkResult base.CheckResultMsg
		err := decoder.Decode(&checkResult)
		if err != nil {
			http.Error(w, "400 Bad Request. Invalid data", http.StatusBadRequest)
			return
		}
		fmt.Printf("Got check result: %+v\n", checkResult)
		ok, _ := checkResult.Validate()
		if !ok {
			http.Error(w, "400 Bad Request. Invalid data", http.StatusBadRequest)
			return
		}
		err = serverdb.AddResult(checkResult.Result, checkResult.Check)
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "405 Method Not allowed", http.StatusMethodNotAllowed)
	}
}
