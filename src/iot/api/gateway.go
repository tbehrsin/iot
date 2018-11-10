package api

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"net/http"
)

type Gateway struct {
	api          *API
	ID           string
	FQDN         string
	Port         uint16
	PrivateKey   *rsa.PrivateKey
	Certificates []*x509.Certificate
	changedCerts bool
}

func (api *API) Gateway() (*Gateway, error) {
	var in map[string]interface{}

	if err := api.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Auth"))
		if buf := b.Get([]byte("gateway")); buf == nil {
			return nil
		} else if err := json.Unmarshal(buf, &in); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if in == nil {
		return nil, nil
	}

	g := &Gateway{
		api:  api,
		ID:   in["ID"].(string),
		FQDN: in["FQDN"].(string),
		Port: uint16(in["Port"].(float64)),
	}

	if in["PrivateKey"] != nil {
		s := in["PrivateKey"].(string)
		if err := g.UnmarshalPrivateKey(&s); err != nil {
			return nil, err
		}
	}

	if in["Certificates"] != nil {
		s := in["Certificates"].(string)
		if err := g.UnmarshalCertificates(&s); err != nil {
			return nil, err
		}
	}

	return g, nil
}

func (g *Gateway) Update() error {
	if buf, err := g.MarshalJSON(); err != nil {
		return err
	} else {
		if err := g.api.DB.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("Auth"))

			if err := b.Put([]byte("gateway"), buf); err != nil {
				return err
			}
			if g.changedCerts {
				g.api.httpClient = nil
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (g *Gateway) MarshalJSON() ([]byte, error) {
	out := map[string]interface{}{
		"ID":           g.ID,
		"FQDN":         g.FQDN,
		"Port":         g.Port,
		"PrivateKey":   g.MarshalPrivateKey(),
		"Certificates": g.MarshalCertificates(),
	}

	if b, err := json.Marshal(out); err != nil {
		return nil, err
	} else {
		return b, nil
	}
}

func (g *Gateway) MarshalPrivateKey() *string {
	if g.PrivateKey == nil {
		return nil
	}

	d := x509.MarshalPKCS1PrivateKey(g.PrivateKey)
	s := base64.StdEncoding.EncodeToString(d)

	return &s
}

func (g *Gateway) UnmarshalPrivateKey(s *string) (err error) {
	if s == nil {
		g.PrivateKey = nil
		return nil
	}

	if d, err := base64.StdEncoding.DecodeString(*s); err != nil {
		return err
	} else if pk, err := x509.ParsePKCS1PrivateKey(d); err != nil {
		return err
	} else {
		g.changedCerts = true
		g.PrivateKey = pk
		return nil
	}
}

func (g *Gateway) MarshalCertificates() *string {
	if g.Certificates == nil || len(g.Certificates) == 0 {
		return nil
	}

	var pemcert []byte
	for _, b := range g.Certificates {
		buf := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: b.Raw})
		pemcert = append(pemcert, buf...)
	}

	g.changedCerts = true

	s := string(pemcert)
	return &s
}

func (g *Gateway) MarshalTLSCertificate() tls.Certificate {
	var certs [][]byte
	for _, b := range g.Certificates {
		certs = append(certs, b.Raw)
	}

	cert := tls.Certificate{
		Certificate: certs,
		PrivateKey:  g.PrivateKey,
		Leaf:        g.Certificates[0],
	}
	return cert
}

func (g *Gateway) HTTPClient() *http.Client {
	if g.api.httpClient != nil {
		return g.api.httpClient
	}

	cert := g.MarshalTLSCertificate()

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	g.api.httpClient = client
	return client
}

func (g *Gateway) UnmarshalCertificates(in *string) error {
	g.Certificates = []*x509.Certificate{}

	if in == nil {
		return nil
	}

	var block *pem.Block
	data := []byte(*in)
	for {
		block, data = pem.Decode(data)
		if block == nil {
			break
		}

		if c, err := x509.ParseCertificates(block.Bytes); err != nil {
			return err
		} else {
			g.Certificates = append(g.Certificates, c...)
		}
	}

	return nil
}

func (gw *Gateway) PublicKey() *rsa.PublicKey {
	if gw.PrivateKey == nil {
		return nil
	}

	k := &rsa.PublicKey{
		N: gw.PrivateKey.N,
		E: gw.PrivateKey.E,
	}

	return k
}

type CreateGatewayRequest struct {
}

type CreateGatewayResponse struct {
	ID   string `json:"id"`
	FQDN string `json:"fqdn"`
}

func (a *API) CreateGateway(in map[string]interface{}) (map[string]interface{}, error) {
	var r CreateGatewayResponse

	if gw, err := a.Gateway(); err != nil {
		return nil, err
	} else if gw != nil {
		return nil, fmt.Errorf("existing gateway already created")
	}

	if b, err := json.Marshal(map[string]interface{}{"Port": 0}); err != nil {
		return nil, err
	} else if req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/gateway/", Server), bytes.NewReader(b)); err != nil {
		return nil, err
	} else {
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

	gw := Gateway{
		api:  a,
		ID:   r.ID,
		FQDN: r.FQDN,
		Port: 0,
	}

	if err := gw.Update(); err != nil {
		return nil, err
	}

	return a.CreateCertificate(in)
}

func (a *API) UpdateGateway() error {
	if gw, err := a.Gateway(); err != nil {
		return err
	} else if gw == nil {
		return fmt.Errorf("no gateway created")
	} else if httpClient := gw.HTTPClient(); httpClient == nil {
		return fmt.Errorf("failed to create https client")
	} else if b, err := json.Marshal(map[string]interface{}{"Port": gw.Port}); err != nil {
		return err
	} else if req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/gateway/", Server), bytes.NewReader(b)); err != nil {
		return err
	} else {
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("cache-control", "no-cache")

		if res, err := httpClient.Do(req); err != nil {
			return err
		} else {
			res.Body.Close()
		}
	}

	return nil
}
