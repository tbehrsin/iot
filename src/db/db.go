package db

import (
	"context"

	"cloud.google.com/go/datastore"
)

var db *datastore.Client

const Domain = "iot.behrsin.com."
const Organization = "Behrsin Ltd"
const CALocality = "Manchester"
const CACountry = "GB"
const CACommonName = "Behrsin IoT CA"
const ValuesKeyName = "projects/behrsin-iot/locations/europe-west2/keyRings/iot/cryptoKeys/iot-datastore"

func Initialize() (err error) {
	c := context.Background()

	if db, err = datastore.NewClient(c, "behrsin-iot"); err != nil {
		return err
	}

	return nil
}
