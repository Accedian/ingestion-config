# Starlink

## Intro

This pipeline configuration is a little odd because it currently relies on rebuilding the Telemetry Collector. This is demo/early POC/alpha quality. To be fixed in a next version.

This was built as a proof of concept for the Starlink engagement and feeds data from the Starlink Telemetry API https://starlink.readme.io/docs/telemetry-api (PW: `flatmcdishyface`) into this PCA tenant: http://stardust.analytics.accedian.io

## Tradeoffs

This should be reworked in the future to avoid several tradeoffs that were made for speed of development.

1. As it stands API credentials are to be placed in the python script
1. The telegraf config expects to run the python script locally using execd. This should be reworked to a pair of microservices 1. Because of that local execd expectation, the Telemetry Collector container is expected to have Python along with the requetes module. This has to be removed
1. No parsing for Alerts - Although Alerts are supported by the API we arent doing anything with them
1. Code only works for `UserTerminal` configs, not the `Router` or the `IpAllocs`

## Setting up
1. Docker pull the latest `gcr.io/sky-agents/agent-telemetry-amd64` container image
1. Based on the tags/version of the latest image, update the Dockefile
1. `docker build` your new python enabled Telemetry Collector (ex.: `docker build -t agent-telemetry-amd64-python3:r24.09 .`)
1. Update the `starlink_telemetry.py` to include your `client_id` and `client_secret`
1. Review and edit your `docker-compose.yaml`
1. Create a new Telemetry Collector on PCA
1. `GET` its config through the API using Postman - `{{tenant-server}}/api/orchestrate/v3/agents/configuration/{{agentId}}`
1. Edit the telemetry -> dataCollection part of the schema to the configuration shared here
1. Edit the templateName to `Starlink-Telemetry-API`
1. `PUT` the edited config back using postman - `{{tenant-server}}/api/orchestrate/v3/agents/configuration/{{agentId}}` 
1. Get the secret for the Teleemtry Collector (API or Web)
1. Create the corresponding `secrets.yaml` file on your host
1. `docker-compose up -d`
1. Wait for the dictionary to show in PCA
1. Ask TAC/CSM to update the dictionary with the one enclosed in this repo though Intercom
1. Enable the metrics you want in your ingestion profile