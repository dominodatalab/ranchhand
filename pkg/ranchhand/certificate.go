package ranchhand

import (
	"io/ioutil"
	"os"

	"github.com/dominodatalab/ranchhand/pkg/x509"
)

const (
	localCertPEM = "cert.pem"
	localKeyPEM  = "key.pem"
)

func generateCertificate(cfg *Config) (certPEM, keyPEM []byte, err error) {
	_, cerr := os.Stat(localCertPEM)
	_, kerr := os.Stat(localKeyPEM)

	if os.IsNotExist(cerr) && os.IsNotExist(kerr) {
		certPEM, keyPEM, err = x509.CreateSelfSignedCert(cfg.CertIPs, cfg.CertDNSNames)
		if err != nil {
			return
		}
		if err = ioutil.WriteFile(localCertPEM, certPEM, 0644); err != nil {
			return
		}
		if err = ioutil.WriteFile(localKeyPEM, keyPEM, 0644); err != nil {
			return
		}
	} else {
		if certPEM, err = ioutil.ReadFile(localCertPEM); err != nil {
			return
		}
		if keyPEM, err = ioutil.ReadFile(localKeyPEM); err != nil {
			return
		}
	}
	return
}
