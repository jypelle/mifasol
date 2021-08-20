package tool

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"
)

func GenerateTlsCertificate(
	organization string,
	serverCommonName string,
	serverKeyFilename, serverCertFilename string,
	hostnames []string) error {

	var err error

	notBefore := time.Now()
	notAfter := notBefore.AddDate(10, 0, 0)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	err = keyToFile(serverKeyFilename, serverKey)
	if err != nil {
		return err
	}

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}
	serverTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{organization},
			CommonName:   serverCommonName,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	for _, h := range hostnames {
		if ip := net.ParseIP(h); ip != nil {
			serverTemplate.IPAddresses = append(serverTemplate.IPAddresses, ip)
		} else {
			serverTemplate.DNSNames = append(serverTemplate.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &serverTemplate, &serverTemplate, &serverKey.PublicKey, serverKey)
	if err != nil {
		return err
	}

	certOut, err := os.Create(serverCertFilename)
	if err != nil {
		return err
	}
	if err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}
	if err = certOut.Close(); err != nil {
		return err
	}

	return nil
}

func keyToFile(filename string, key *ecdsa.PrivateKey) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	b, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}
	if err = pem.Encode(file, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}); err != nil {
		return err
	}
	return nil
}

func CertToMemory(derBytes []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
}

/*
// CertToPEM is a utility function returns a PEM encoded x509 Certificate
func CertToPEM(cert *x509.Certificate) []byte {
	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})

	return pemCert
}

// CertChainToPEM is a utility function returns a PEM encoded chain of x509 Certificates, in the order they are passed
func CertChainToPEM(certChain []*x509.Certificate) ([]byte, error) {
	var pemBytes bytes.Buffer
	for _, cert := range certChain {
		if err := pem.Encode(&pemBytes, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}); err != nil {
			return nil, err
		}
	}
	return pemBytes.Bytes(), nil
}
*/
