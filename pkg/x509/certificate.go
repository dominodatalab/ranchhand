package x509

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/pkg/errors"
)

func CreateSelfSignedCert() (certPEM, keyPEM []byte, err error) {
	// generate a new key-pair
	rootKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}

	// generate cert template used to self-sign
	certTmpl, err := certTemplate()
	if err != nil {
		return
	}
	certTmpl.IsCA = true
	certTmpl.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature
	certTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}

	// generate self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, certTmpl, certTmpl, &rootKey.PublicKey, rootKey)
	if err != nil {
		return
	}

	// convert cert/key to PEM format
	certPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})
	keyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rootKey),
	})
	return
}

func certTemplate() (*x509.Certificate, error) {
	// generate a random serial number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate new serial number")
	}

	now := time.Now()
	tmpl := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Domino Data Lab, Inc."},
			CommonName:   "domino.rancher",
		},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             now,
		NotAfter:              now.AddDate(10, 0, 0),
		BasicConstraintsValid: true,
	}
	return &tmpl, nil
}
