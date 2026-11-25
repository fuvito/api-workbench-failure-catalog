package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Every port sends a certificate that is in date, issued for the host name
// clients use, together with the intermediate that links it to the root.
//
// All keys are created at start-up and kept in memory. Only the test root's
// public certificate is written to disk, for clrsig to trust.

const caFile = "failures/fm-032-01-tls-certificate/.certs/ca.pem"

type authority struct {
	cert *x509.Certificate
	der  []byte
	key  *ecdsa.PrivateKey
}

// issue creates a certificate signed by parent, or self-signed if parent is nil.
func issue(parent *authority, template *x509.Certificate) authority {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	if err != nil {
		log.Fatal(err)
	}
	template.SerialNumber = serial
	issuerCert, signer := template, crypto.Signer(key)
	if parent != nil {
		issuerCert, signer = parent.cert, parent.key
	}
	der, err := x509.CreateCertificate(rand.Reader, template, issuerCert, &key.PublicKey, signer)
	if err != nil {
		log.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		log.Fatal(err)
	}
	return authority{cert: cert, der: der, key: key}
}

func ca(name string) *x509.Certificate {
	return &x509.Certificate{
		Subject:   pkix.Name{CommonName: name},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(24 * time.Hour),
		IsCA: true, BasicConstraintsValid: true,
		KeyUsage: x509.KeyUsageCertSign,
	}
}

func leaf(host string, notBefore, notAfter time.Time) *x509.Certificate {
	return &x509.Certificate{
		Subject:   pkix.Name{CommonName: host},
		DNSNames:  []string{host},
		NotBefore: notBefore, NotAfter: notAfter,
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
}

func serveTLS(address string, server authority, intermediates ...authority) {
	chain := [][]byte{server.der}
	for _, i := range intermediates {
		chain = append(chain, i.der)
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	config := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{{Certificate: chain, PrivateKey: server.key}},
	}
	go func() { log.Fatal(http.Serve(tls.NewListener(listener, config), mux)) }()
}

func main() {
	root := issue(nil, ca("FM-032 Test Root"))
	intermediate := issue(&root, ca("FM-032 Test Intermediate"))
	now := time.Now()

	valid := func() authority {
		return issue(&intermediate, leaf("localhost", now.Add(-time.Hour), now.Add(12*time.Hour)))
	}
	for _, address := range []string{":19077", ":19085", ":19086"} {
		serveTLS(address, valid(), intermediate)
	}

	if err := os.MkdirAll(filepath.Dir(caFile), 0o755); err != nil {
		log.Fatal(err)
	}
	rootPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: root.der})
	if err := os.WriteFile(caFile, rootPEM, 0o644); err != nil {
		log.Fatal(err)
	}

	// Plain HTTP, for the readiness check only. It starts after the CA is written.
	health := http.NewServeMux()
	health.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	log.Fatal(http.ListenAndServe(":19076", health))
}
