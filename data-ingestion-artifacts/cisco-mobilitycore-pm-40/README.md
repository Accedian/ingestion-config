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

### Step 3: Upload Dictionaries to PCA

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

## Prerequisites

- Go runtime (minimum version: **1.16**, tested with go1.25.7)
- Access to the target PCA instance
- Valid authentication token
