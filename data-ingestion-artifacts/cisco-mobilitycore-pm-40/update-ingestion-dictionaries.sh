#!/bin/sh

GO_IMAGE="golang:1.26.0-alpine3.23"

export BASE_URL="https://pca.kajar.npav.accedian.net"
export AUTHORIZATION_HEADER="Bearer your-token-here"
export DICTIONARIES_PATH="./ingestion-artifacts"
export INSECURE_SKIP_VERIFY="true"  # for self-signed certs

docker run --rm -v "$(pwd):/work" -w /work \
  -e BASE_URL -e AUTHORIZATION_HEADER -e DICTIONARIES_PATH -e INSECURE_SKIP_VERIFY \
  "$GO_IMAGE" \
  go run dictionaryuploader/main.go

