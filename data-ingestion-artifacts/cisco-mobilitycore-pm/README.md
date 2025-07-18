# cisco-mobilitycore-pm

## Intro

This set of files represents the first integration between Provider Connectivity Assurance and the Fault & Mobility Performance Monitoring collector. The interface between both systems is a kafka bus.

## Tradeoffs and important notes

The telegraf config is build with the expectation that the payload will comply with a few requirements. It also would benefit from templating to decouple the target IP addresses for the kafka server from the actual transformation.
The telegraf config is build with the expectation that the payload will comply with a few requirements. It also would benefit from templating to decouple the target IP addresses for the kafka server from the actual transformation.

1. As it stands, target IP/port for Kafka is embedded in the config which doesnt scale and makes it more difficult to deploy.
1. The json payloads on the kafka bus need to comply with the expected format and contain a minimum required set of key:value pairs. See below for reference
1. As agreed during the planning and architecture discussions, we decided to *not* ingest any `"schema": "p2p"` data. This data set is very large (800+ objects), with only one KPI each. A second round of engineering will be done to assess how to best bring this data into the platform. 
1. The telegraf file outputs can be sent to `/tmp/metrics.out` and `/tmp/missing_schema.out`. Those files are meant to be used for troubleshooting. They are also mounted as volumes in the `docker-compose.yml` supplied as an example. In order to avoid filling your storage space needlessly, I recommend you create symlinks to `/dev/null` for `./metrics.out` and `./missing_schema.out` in your local directory before starting the `docker compose up -d`. You can always overwrite those later if you suddenly want to peek into the output.
1. Similarly a prometheus server can listen on port 9273 for troubleshooting purposes. 

## Files

The telegraf configuration file you will need to retreive and edit can be found in the collection-artifacts folder, and all the custom dictionaries we have build at this time can be found in the ingestion-artifacts folder.

A sample docker-compose.yaml file can be found in the supporting-files folder should you need it.

## Sample kafka json payload

Each interval will produce an anonymous array of anonymous objects as follows. There could be more key:value pairs in the future as we enhance the data stream. As for v1, all the keys shown in the example are required.

```
[
   {
      "device": "svceed22",
      "kpi": "fiveg_always_on_disable",
      "index": "[\"smf-svsmft22\"]",
      "value": 0,
      "timestamp": 1749213002,
      "node_id": "230",
      "schema": "smf",
      "source_ip": "0.0.0.0"
   },
   {
      "device": "svceed22",
      "kpi": "fiveg_always_on_enable",
      "index": "[\"smf-svsmft22\"]",
      "value": 0,
      "timestamp": 1749213002,
      "node_id": "230",
      "schema": "smf",
      "source_ip": "0.0.0.0"
   },
   {
      "device": "svceed22",
      "kpi": "fiveg_current_pdu_sessions",
      "index": "[\"smf-svsmft22\"]",
      "value": 0,
      "timestamp": 1749213002,
      "node_id": "230",
      "schema": "smf",
      "source_ip": "0.0.0.0"
   }
]
```