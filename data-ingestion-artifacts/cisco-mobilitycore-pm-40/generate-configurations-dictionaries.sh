#!/bin/sh

GO_IMAGE="golang:1.26.0-alpine3.23"

COLLECTOR_OUTPUT_DIR="collector_artifacts"    
DICTIONARIES_DIR="ingestion-artifacts"
# Create the output directory if it doesn't exist
mkdir -p "$COLLECTOR_OUTPUT_DIR" "$DICTIONARIES_DIR"

docker run --rm -v "$(pwd):/work" -w /work "$GO_IMAGE" \
  go run generate_telegraf_configs/generate_telegraf_configs.go \
  -csv Kpi_calc_Kpicatalog-updatedGrouping.csv \
  -output "$COLLECTOR_OUTPUT_DIR" \
  -dictionaries "$DICTIONARIES_DIR"

