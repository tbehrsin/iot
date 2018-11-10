#!/bin/bash

export AWS_PROFILE=behrsin

#ENDPOINT_URL="--endpoint-url=http://localhost:8000"
ENDPOINT_URL=""

aws dynamodb create-table $ENDPOINT_URL --table-name Gateways \
  --attribute-definitions AttributeName=ID,AttributeType=S \
  --key-schema AttributeName=ID,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5

aws dynamodb put-item $ENDPOINT_URL --table-name Gateways \
  --item '{"ID": {"S": "test"}, "Address": {"S": "192.168.2.13"}, "Port":  {"N": "443"}}'
