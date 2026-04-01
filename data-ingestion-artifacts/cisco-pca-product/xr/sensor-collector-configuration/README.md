## Sensor Collector Notes

This folder contains sensor-collector artifacts for Cisco PCA XR work.

### `sensor-collector-rate-delta.json`

This file is a template-derived sample, not a frozen live tenant object.

Template provenance:
- derived from the live `telemetry-collector` template fetched on `2026-04-01`
- source endpoint: `GET /api/v2/connector-configs/templates`
- selected object: `id = telemetry-collector`
- template `_rev`: `5-0577a6a64178c7dc159940cfc6089b59`
- `openMetricsConfig.metricMappingsVersion`: `0.10.0`

How it was built:
- start from `GET /api/v2/connector-configs/templates`
- select the object with `id = telemetry-collector`
- preserve the template's `openMetricsConfig` baseline
- keep the richer Git-managed `cisco-telemetry-xe-interface` delta/rate set if the live template is narrower
- add the XR policy overlay under `openMetricsConfig.operations.cisco-telemetry-xr-policy`

Where the overlay fits:
- base template path: `data.attributes.openMetricsConfig.operations`
- overlay key: `cisco-telemetry-xr-policy`

Intended API workflow:
1. `GET /api/v2/connector-configs/templates`
2. take the `telemetry-collector` template as the base
3. merge this XR policy delta block into `data.attributes.openMetricsConfig.operations`
4. use the generated result when creating or patching a telemetry collector

How to refresh this sample later:
1. log in to a tenant with access to connector-config templates
2. call `GET /api/v2/connector-configs/templates`
3. extract the object with `id = telemetry-collector`
4. compare `_rev`, `metricMappingsVersion`, and baseline `openMetricsConfig.operations`
5. preserve the repo's richer `cisco-telemetry-xe-interface` delta/rate set if the fresh template contains fewer XE metrics
6. rebuild this sample from that fresh template before reapplying the XR policy overlay

Current intended XR policy delta metrics:
- `policy_stats_class_stats_general_stats_transmit_packets`
- `policy_stats_class_stats_general_stats_transmit_bytes`
- `policy_stats_class_stats_general_stats_total_drop_packets`
- `policy_stats_class_stats_general_stats_total_drop_bytes`
- `policy_stats_class_stats_general_stats_pre_policy_matched_bytes`
- `policy_stats_class_stats_general_stats_pre_policy_matched_packets`

Current XE interface note:
- the live `telemetry-collector` template currently carries a reduced XE interface delta list
- this repo sample intentionally preserves the broader Git-managed XE interface delta/rate set so that reconciliation with the live template does not drop previously curated XE handling
