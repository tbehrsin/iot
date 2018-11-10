package db

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/rs/xid"
	"log"
	"strings"
)

type Gateway struct {
	ID        string
	Address   string
	Port      uint16
	PublicKey string
}

func (g *Gateway) FQDN() string {
	return fmt.Sprintf("%s.%s", g.ID, Domain)
}

func (g *Gateway) FQDNWithoutDot() string {
	return strings.TrimSuffix(g.FQDN(), ".")
}

func (g *Gateway) Update() error {
	if av, err := dynamodbattribute.MarshalMap(g); err != nil {
		return err
	} else {
		input := &dynamodb.PutItemInput{
			TableName: aws.String("Gateways"),
			Item:      av,
		}

		log.Printf("updating gateway %s\n", g.ID)

		if _, err := db.PutItem(input); err != nil {
			return err
		}
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
	input := &dynamodb.GetItemInput{
		TableName: aws.String("Gateways"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(id),
			},
		},
	}

	gw := new(Gateway)

	if result, err := db.GetItem(input); err != nil {
		return nil, err
	} else if result.Item == nil {
		return nil, nil
	} else if err := dynamodbattribute.UnmarshalMap(result.Item, gw); err != nil {
		return nil, err
	}

	return gw, nil
}

func CreateGateway(address string, port uint16) (*Gateway, error) {
	id := xid.New()

	input := &dynamodb.PutItemInput{
		TableName: aws.String("Gateways"),
		Item: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(id.String()),
			},
			"Address": {
				S: aws.String(address),
			},
			"Port": {
				N: aws.String(fmt.Sprintf("%d", port)),
			},
		},
	}

	gw := new(Gateway)

	log.Printf("creating gateway %s\n", id.String())

	if _, err := db.PutItem(input); err != nil {
		return nil, err
	} else if err := dynamodbattribute.UnmarshalMap(input.Item, gw); err != nil {
		return nil, err
	}

	return gw, nil
}
