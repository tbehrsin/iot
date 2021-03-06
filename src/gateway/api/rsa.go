package api

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gateway/errors"
	"io/ioutil"
	"net/http"
)

type CreateCertificateRequest struct {
	CSR string `json:"csr"`
}

type CreateCertificateResponse struct {
	Certificate string `json:"certificate"`
}

func (a *API) CreateCertificate(in map[string]interface{}) (map[string]interface{}, error) {
	var r CreateCertificateResponse

	if g, err := a.Gateway(); err != nil {
		return nil, err
	} else if g == nil {
		return nil, errors.NewNotFound("no existing gateway found")
	} else {
		if kp, err := rsa.GenerateKey(rand.Reader, 2048); err != nil {
			return nil, err
		} else {
			g.PrivateKey = kp
		}

		req := &x509.CertificateRequest{
			Subject:  pkix.Name{CommonName: g.FQDN},
			DNSNames: []string{g.FQDN, fmt.Sprintf("local.%s", g.FQDN)},
		}

		if csr, err := x509.CreateCertificateRequest(rand.Reader, req, g.PrivateKey); err != nil {
			return nil, errors.NewInternalServerError("failed to create certificate request: %+v", err)
		} else {
			c := CreateCertificateRequest{CSR: base64.StdEncoding.EncodeToString(csr)}

			if b, err := json.Marshal(c); err != nil {
				return nil, err
			} else {
				req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/gateway/%s/certificate/", Server, g.ID), bytes.NewReader(b))

				req.Header.Add("Content-Type", "application/json")
				req.Header.Add("cache-control", "no-cache")

				if res, err := http.DefaultClient.Do(req); err != nil {
					return nil, err
				} else {
					defer res.Body.Close()
					if body, err := ioutil.ReadAll(res.Body); err != nil {
						return nil, err
					} else if err := json.Unmarshal(body, &r); err != nil {
						return nil, err
					}
				}
			}
		}

		if err := g.UnmarshalCertificates(&r.Certificate); err != nil {
			return nil, errors.NewInternalServerError("failed to unmarshal certificates: %+v", err)
		}

		if err := g.Update(); err != nil {
			return nil, err
		} else {
			if a.running {
				a.Stop()
			}

			if err := a.Start(); err != nil {
				return nil, err
			}
			return make(map[string]interface{}), nil
		}
	}
}
