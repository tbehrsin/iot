package db

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/rs/xid"
)

type Gateway struct {
	ID           string `datastore:"id"`
	IsOrigin     bool   `datastore:"isOrigin,noindex"`
	Address      string `datastore:"address,noindex"`
	Port         int    `datastore:"port,noindex"`
	LocalAddress string `datastore:"localAddress,noindex"`
	PublicKey    string `datastore:"publicKey,noindex"`
}

func (g *Gateway) FQDN() string {
	return fmt.Sprintf("%s.%s", g.ID, Domain)
}

func (g *Gateway) FQDNWithoutDot() string {
	return strings.TrimSuffix(g.FQDN(), ".")
}

func (g *Gateway) Update() error {
	k := datastore.NameKey("Gateway", g.ID, nil)

	c := context.Background()
	if _, err := db.Put(c, k, g); err != nil {
		return err
	}

	return nil
}

func (g *Gateway) UnmarshalPublicKey() (*rsa.PublicKey, error) {
	if b, err := base64.StdEncoding.DecodeString(g.PublicKey); err != nil {
		return nil, err
	} else if publicKey, err := x509.ParsePKCS1PublicKey(b); err != nil {
		return nil, err
	} else {
		return publicKey, nil
	}
}

func (g *Gateway) MarshalPublicKey(publicKey *rsa.PublicKey) error {
	b := x509.MarshalPKCS1PublicKey(publicKey)
	s := base64.StdEncoding.EncodeToString(b)
	g.PublicKey = s
	return nil
}

func GetGateway(id string) (*Gateway, error) {
	k := datastore.NameKey("Gateway", id, nil)
	gw := new(Gateway)

	c := context.Background()
	if err := db.Get(c, k, gw); err != nil {
		return nil, err
	}

	return gw, nil
}

func CreateGateway(address string, port int) (*Gateway, error) {
	id := xid.New()
	gw := &Gateway{
		ID:      id.String(),
		Address: address,
		Port:    port,
	}
	k := datastore.NameKey("Gateway", gw.ID, nil)

	c := context.Background()
	if _, err := db.Put(c, k, gw); err != nil {
		return nil, err
	}

	return gw, nil
}
