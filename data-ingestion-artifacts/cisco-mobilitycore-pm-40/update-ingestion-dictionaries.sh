#!/bin/sh

export BASE_URL="https://pca.kajar.npav.accedian.net"
export AUTHORIZATION_HEADER="Bearer your-token-here"
export DICTIONARIES_PATH="./ingestion-artifacts"
export INSECURE_SKIP_VERIFY="true"  # for self-signed certs

go run dictionaryuploader/main.go 

