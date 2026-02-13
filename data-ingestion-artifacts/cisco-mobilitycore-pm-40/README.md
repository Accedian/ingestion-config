# Cisco Mobility Core PM Configuration Tools

This directory contains tools for generating Telegraf collection configurations and ingestion dictionaries from a KPI catalog CSV file.

## Workflow

### Step 1: Update the CSV File

Edit `Kpi_calc_Kpicatalog-updatedGrouping.csv` to define or modify the KPI metrics and their groupings. This file serves as the source of truth for generating all configuration artifacts.

### Step 2: Generate Configuration Files

Run the generation script to produce Telegraf configuration files and corresponding ingestion dictionaries:

```bash
./generate-configurations-dictionaries.sh
```

This script will:
- Read the CSV file (`Kpi_calc_Kpicatalog-updatedGrouping.csv`)
- Generate Telegraf collector configuration files in the `collector_artifacts/` directory
- Generate ingestion dictionary JSON files in the `ingestion-artifacts/` directory

> **Warning:** Files in `collector_artifacts/` and `ingestion-artifacts/` directories are **overwritten** each time this script runs.

### Step 3: Update Telemetry Collector Configuration

If you have an existing Telemetry Collector deployment, update its configuration to use the newly generated Telegraf configs from `collector_artifacts/`.

Refer to the [Deploying Telemetry Collector in Docker](https://docs.accedian.io/docs/deploying-telemetry-collector-in-docker) guide for deployment instructions.

### Step 4: Upload Dictionaries to PCA

After generating the dictionaries, upload them to PCA using:

```bash
./update-ingestion-dictionaries.sh
```

Before running, ensure the following environment variables are properly configured in the script:
- `BASE_URL` - The PCA instance URL
- `AUTHORIZATION_HEADER` - Bearer token for authentication
- `INSECURE_SKIP_VERIFY` - Set to `true` for self-signed certificates

## Directory Structure

| Directory | Description |
|-----------|-------------|
| `collector_artifacts/` | Generated Telegraf configuration files (overwritten on generation) |
| `ingestion-artifacts/` | Generated ingestion dictionary JSON files (overwritten on generation) |
| `generate_telegraf_configs/` | Go source code for the configuration generator |
| `dictionaryuploader/` | Go source code for the dictionary uploader tool |

## Troubleshooting

### Duplicate ObjectType Error

When uploading dictionaries, you may encounter a `duplicateObjectType` error if an OpenMetrics dictionary with the same objectType already exists. To resolve this:

1. **Disable metrics** from the ingestion profile in PCA that reference the conflicting objectType
2. **Delete the OpenMetrics dictionary** via the PCA API:
   ```bash
   curl -X DELETE "${BASE_URL}/api/v3/ingestion-dictionaries/<dictionary-id>" \
     -H "Authorization: ${AUTHORIZATION_HEADER}" \
     -H "Accept: application/vnd.api+json"
   ```
3. **Re-run** the dictionary uploader to create the new Cisco dictionaries

## Prerequisites

- Go runtime (minimum version: **1.16**, tested with go1.25.7)
- Access to the target PCA instance
- Valid authentication token (see [API Authentication Guide](https://api.accedian.io/session.html#section/Quick-setup/Authenticate-to-obtain-an-Authorized-Token))
