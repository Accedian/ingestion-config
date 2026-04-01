## Sensor Collector Overlay Notes

This folder contains overlay-style sensor-collector additions for Nokia SR OS gNMI telemetry.

### `sensor-collector-rate-delta.json`

This file is not a full connector object.
It is an overlay fragment meant to be merged into a telemetry collector's `openMetricsConfig.operations`.

Where the overlay fits:
- target path: `data.attributes.openMetricsConfig.operations`
- overlay keys:
  - `nokia_telemetry_interface`
  - `nokia_telemetry_interface_queues`

Intended API workflow:
1. `GET /api/v2/connector-configs/templates`
2. start from the `telemetry-collector` template or from a live collector object
3. merge the operations from this file into `openMetricsConfig.operations`
4. `PATCH /api/v2/connector-configs/{connectorId}` with the merged config, preserving `_rev`

Base template note:
- use the live `telemetry-collector` template from `GET /api/v2/connector-configs/templates`
- the current repo XR sample was refreshed from template `_rev = 5-0577a6a64178c7dc159940cfc6089b59` with `metricMappingsVersion = 0.10.0`

Use this overlay when the collector must process Nokia interface and queue metrics with the expected delta/rate behavior.
