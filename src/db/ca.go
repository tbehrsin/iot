package db

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

var caParameterName = "iot-ca"

type CA struct {
	Certificate *x509.Certificate
	PrivateKey  *rsa.PrivateKey
}

func CreateCA() error {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{Organization},
			Country:      []string{CACountry},
			Locality:     []string{CALocality},
			CommonName:   CACommonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	if privateKey, err := rsa.GenerateKey(rand.Reader, 2048); err != nil {
		return err
	} else if caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &privateKey.PublicKey, privateKey); err != nil {
		return err
	} else {
		b := make([]byte, 0, 4096)
		b = append(b, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caBytes})...)
		b = append(b, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})...)
		s := string(b)

		if err := SetValueSecure(caParameterName, s); err != nil {
			return err
		}
	}

	return nil
}

func GetCA() (*CA, error) {
	if v, err := GetValueSecure(caParameterName); err != nil {
		return nil, fmt.Errorf("error getting ca certificate: %+v", err)
	} else {
		ca := &CA{}
		var block *pem.Block
		data := []byte(v)
		for {
			block, data = pem.Decode(data)
			if block == nil {
				break
			}

			if block.Type == "CERTIFICATE" {
				if c, err := x509.ParseCertificates(block.Bytes); err != nil {
					return nil, err
				} else {
					ca.Certificate = c[0]
				}
			} else if block.Type == "RSA PRIVATE KEY" {
				if p, err := x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
					return nil, err
				} else {
					ca.PrivateKey = p
				}
			}
		}
		return ca, nil
	}
}

func (ca *CA) CreateServerCertificate(name string) (*tls.Certificate, error) {
	template := &x509.CertificateRequest{
		Subject: pkix.Name{CommonName: name},
	}

	if privateKey, err := rsa.GenerateKey(rand.Reader, 2048); err != nil {
		return nil, err
	} else if csrBytes, err := x509.CreateCertificateRequest(rand.Reader, template, privateKey); err != nil {
		return nil, err
	} else if csr, err := x509.ParseCertificateRequest(csrBytes); err != nil {
		return nil, err
	} else if certs, err := ca.Sign(csr); err != nil {
		return nil, err
	} else {
		var c [][]byte
		for _, b := range certs {
			c = append(c, b.Raw)
		}

		cert := &tls.Certificate{
			Certificate: c,
			PrivateKey:  privateKey,
			Leaf:        certs[0],
		}
		return cert, nil
	}
}

func (ca *CA) Sign(csr *x509.CertificateRequest) ([]*x509.Certificate, error) {
	template := x509.Certificate{
		Signature:          csr.Signature,
		SignatureAlgorithm: csr.SignatureAlgorithm,

		PublicKeyAlgorithm: csr.PublicKeyAlgorithm,
		PublicKey:          csr.PublicKey,

		SerialNumber: big.NewInt(2),
		Issuer:       ca.Certificate.Subject,
		Subject:      pkix.Name{CommonName: csr.Subject.CommonName},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Duration(90*24) * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	if bytes, err := x509.CreateCertificate(rand.Reader, &template, ca.Certificate, csr.PublicKey, ca.PrivateKey); err != nil {
		return nil, err
	} else if certs, err := x509.ParseCertificates(bytes); err != nil {
		return nil, err
	} else {
		certs = append(certs, ca.Certificate)
		return certs, nil
	}
}

func (ca *CA) CertPool() (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	pool.AddCert(ca.Certificate)
	return pool, nil
}
