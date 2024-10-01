# Ingestion Configuration
This repository is intended to act as the source of truth for configuration artifacts associated with the Ingestion pipeline.

## File Structure
The `data-ingestion-artifacts` folder contains sub-folders for every object type that we have artifacts for.
Each object type sub-directory contains files which respect the following naming conventions:

    `collection-{{object type}}.toml` - Telegraf toml for Telemetry Collector'
    `ingestion-{{object type}}.yaml` - Prometheus yaml for Roadrunner'
    `ingestion-{{object type}}.json' - Operations json for Roadrunner'