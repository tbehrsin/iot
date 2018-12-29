package api

import (
	"encoding/json"
	"gateway/errors"

	"github.com/boltdb/bolt"
)

type PINCode struct {
	Hash string `json:"hash"`
	Seed string `json:"seed"`
}

func (api *API) GetPINCodeSeed(in map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	var buf []byte

	if err := api.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Auth"))
		if buf = b.Get([]byte("pin")); buf == nil {
			return errors.NewNotFound("no pin code has been set")
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var p PINCode
	if err := json.Unmarshal(buf, &p); err != nil {
		return nil, err
	}

	out["seed"] = p.Seed

	return out, nil
}

func (api *API) SetPINCode(in map[string]interface{}) (map[string]interface{}, error) {
	var hash, seed string

	if v, ok := in["hash"]; !ok {
		return nil, errors.NewBadRequest("pin hash not found")
	} else if hash, ok = v.(string); !ok {
		return nil, errors.NewBadRequest("pin hash not a string")
	}

	if v, ok := in["seed"]; !ok {
		return nil, errors.NewBadRequest("pin seed not found")
	} else if seed, ok = v.(string); !ok {
		return nil, errors.NewBadRequest("pin seed not a string")
	}

	p := PINCode{Hash: hash, Seed: seed}
	if buf, err := json.Marshal(p); err != nil {
		return nil, err
	} else {
		if err := api.DB.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("Auth"))
			if err := b.Put([]byte("pin"), buf); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return api.CreateTokenBLE(map[string]interface{}{})
}

func (api *API) VerifyPINCode(in map[string]interface{}) (map[string]interface{}, error) {
	var hash string

	if v, ok := in["hash"]; !ok {
		return nil, errors.NewBadRequest("pin hash not found")
	} else if hash, ok = v.(string); !ok {
		return nil, errors.NewBadRequest("pin hash not a string")
	}

	var buf []byte

	if err := api.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Auth"))
		if buf = b.Get([]byte("pin")); buf == nil {
			return errors.NewNotFound("no pin code has been set")
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var p PINCode
	if err := json.Unmarshal(buf, &p); err != nil {
		return nil, err
	}

	if hash == p.Hash {
		return api.CreateTokenBLE(map[string]interface{}{})
	} else {
		return nil, errors.NewUnauthorized("incorrect pin code")
	}
}
