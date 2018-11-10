#!/bin/bash -ex

TABLE_NAME=$1
aws dynamodb scan --table-name $TABLE_NAME | jq -r '.Items[].ID.S' | xargs -Iid aws dynamodb delete-item --table-name $TABLE_NAME --key '{ "ID": { "S": "id" }}'
