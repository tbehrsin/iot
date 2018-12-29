package main

import (
	"crypto/rsa"
	"crypto/x509"
	"db"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

type CreateCertificateRequest struct {
	CSR string `json:"csr"`
}

func (csr *CreateCertificateRequest) GetCSRBytes() ([]byte, error) {
	return base64.StdEncoding.DecodeString(csr.CSR)
}

func (csr *CreateCertificateRequest) GetCSR() (*x509.CertificateRequest, error) {
	if c, err := csr.GetCSRBytes(); err != nil {
		return nil, fmt.Errorf("failed to parse certificate request: %+v", err)
	} else {
		return x509.ParseCertificateRequest(c)
	}
}

func (csr *CreateCertificateRequest) Validate(fqdn string) error {
	if c, err := csr.GetCSR(); err != nil {
		return err
	} else if c.Subject.CommonName != fqdn {
		return fmt.Errorf("invalid certificate common name")
	} else if len(c.ExtraExtensions) != 0 || len(c.Attributes) != 1 || len(c.DNSNames) != 2 || c.DNSNames[0] != fqdn || c.DNSNames[1] != fmt.Sprintf("local.%s", fqdn) || len(c.EmailAddresses) != 0 || len(c.IPAddresses) != 0 || len(c.URIs) != 0 {
		return fmt.Errorf("invalid certificate extensions")
	}

	return nil
}

type CreateCertificateResponse struct {
	Certificate string `json:"certificate"`
}

func CreateCertificate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var request CreateCertificateRequest

	if body, err := ioutil.ReadAll(r.Body); err != nil {
		APIError(w, err)
		return
	} else if err := json.Unmarshal([]byte(body), &request); err != nil {
		APIError(w, err)
		return
	} else if csr, err := request.GetCSR(); err != nil {
		APIError(w, err)
		return
	} else if publicKey, ok := csr.PublicKey.(*rsa.PublicKey); !ok {
		APIError(w, fmt.Errorf("could not find an rsa public key to sign"))
		return
	} else if gateway, err := db.GetGateway(vars["id"]); err != nil {
		APIError(w, err)
		return
	} else if gateway == nil {
		APIErrorWithStatus(w, fmt.Errorf("unknown gateway"), http.StatusBadRequest)
		return
	} else if ca, err := db.GetCA(); err != nil {
		APIError(w, err)
		return
	} else if err := request.Validate(gateway.FQDNWithoutDot()); err != nil {
		APIError(w, err)
		return
	} else {
		if certs, err := ca.Sign(csr); err != nil {
			APIError(w, err)
			return
		} else if err := gateway.MarshalPublicKey(publicKey); err != nil {
			APIError(w, err)
			return
		} else {
			if err := gateway.Update(); err != nil {
				APIError(w, err)
				return
			}

			var pemcert []byte
			for _, cert := range certs {
				b := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
				pemcert = append(pemcert, b...)
			}

			APIJSON(w, CreateCertificateResponse{
				Certificate: string(pemcert),
			})
		}
	}
}
