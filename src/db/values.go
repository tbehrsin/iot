package db

import (
	"context"
	"encoding/base64"

	"cloud.google.com/go/datastore"
	kms "cloud.google.com/go/kms/apiv1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

type Value struct {
	Value string `datastore:"value,noindex"`
}

func SetValue(key string, value string) error {
	v := &Value{value}
	k := datastore.NameKey("Values", key, nil)

	c := context.Background()
	if _, err := db.Put(c, k, v); err != nil {
		return err
	}

	return nil
}

func SetValueSecure(key string, value string) error {
	ctx := context.Background()
	if c, err := kms.NewKeyManagementClient(ctx); err != nil {
		return err
	} else if resp, err := c.Encrypt(ctx, &kmspb.EncryptRequest{
		Name:      ValuesKeyName,
		Plaintext: []byte(value),
	}); err != nil {
		return err
	} else {
		ct64 := base64.StdEncoding.EncodeToString(resp.Ciphertext)
		return SetValue(key, ct64)
	}
}

func GetValue(key string) (string, error) {
	k := datastore.NameKey("Values", key, nil)
	v := new(Value)

	c := context.Background()
	if err := db.Get(c, k, v); err != nil {
		return "", err
	}

	return v.Value, nil
}

func GetValueSecure(key string) (string, error) {
	ctx := context.Background()
	if ct64, err := GetValue(key); err != nil {
		return "", err
	} else if ct, err := base64.StdEncoding.DecodeString(ct64); err != nil {
		return "", err
	} else if c, err := kms.NewKeyManagementClient(ctx); err != nil {
		return "", err
	} else if resp, err := c.Decrypt(ctx, &kmspb.DecryptRequest{
		Name:       ValuesKeyName,
		Ciphertext: []byte(ct),
	}); err != nil {
		return "", err
	} else {
		return string(resp.Plaintext), nil
	}
}
