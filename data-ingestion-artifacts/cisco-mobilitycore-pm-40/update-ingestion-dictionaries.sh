#!/bin/sh

export BASE_URL="https://pca.kajar.npav.accedian.net"
#export AUTHORIZATION_HEADER="Bearer your-token-here"
export AUTHORIZATION_HEADER="Bearer qI5TsYU1AJ7AEm4affRBvT8TVeYMcZ_C8EYwpBM3aFONuIR_TyBjyTZJUD3GIRZuBsiDxeFpEnDzde0bZfTHbDfS8_euxUnUdvmSQpJF"
export DICTIONARIES_PATH="./ingestion-artifacts"
export INSECURE_SKIP_VERIFY="true"  # for self-signed certs

go run dictionaryuploader/main.go 

