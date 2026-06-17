# Cisco IOS XE Supplemental

Community extension package for supplemental IOS XE PCA telemetry coverage.

This package is intentionally kept under `cisco-community` because it is a demo/PoC extension of the out-of-the-box IOS XE collector, not a product baseline change yet.

## Scope

The bundled collector configuration starts from the out-of-the-box IOS XE MDT configuration and adds support for:

- `Cisco-IOS-XE-bgp-oper:bgp-state-data/neighbors/neighbor`
- `Cisco-IOS-XE-environment-oper:environment-sensors/environment-sensor`
- `Cisco-IOS-XE-switch-ptp-oper:switch-ptp-oper-data`

The requested IOS XE subscription path for environment telemetry is the parent container:

```text
/environment-ios-xe-oper:environment-sensors
```

The Telegraf alias targets the keyed `environment-sensor` list below that container. The supplemental transform aggregates those sensor rows into one PCA object per router/source, with each sensor reading represented as a named metric.

The PTP subscription path should use the IOS XE YANG prefix, not the module name:

```text
/switch-ptp-ios-xe-oper:switch-ptp-oper-data
```

For a `60` second / `1` minute PTP cadence on the Catalyst 9300 lab switch, `update-policy periodic 6000` was the corrected setting. The earlier `10000` setting produced `100` second spacing.

## Included Artifacts

- `telemetry-collector-configuration/input-cisco-telemetry-xe-supplemental-mdt.conf`: full demo collector config, based on the product IOS XE MDT config plus supplemental BGP/environment/PTP additions.
- `telemetry-collector-configuration/transformation-cisco-telemetry-xe-bgp-neighbor.conf`: standalone BGP shaping block for review or later merge.
- `telemetry-collector-configuration/transformation-cisco-telemetry-xe-environment-sensor.conf`: standalone environment-sensor shaping block for review or later merge.
- `telemetry-collector-configuration/transformation-cisco-telemetry-xe-ptp.conf`: standalone PTP shaping block for review or later merge.
- `sensor-collector-configuration/cisco-telemetry-xe-supplemental.yaml`: job list with the three out-of-the-box IOS XE objects plus BGP, environment, and four PTP extension objects.
- `sensor-collector-configuration/cisco-telemetry-xe-bgp-neighbor-delta-operations.json`: Sensor Collector operations fragment to merge under `openMetricsConfig.operations` for BGP churn-counter deltas.
- `pca-ingestion-specifications-configuration/`: PCA ingestion-staging specifications. This is the forward default artifact format.
- `pca-ingestion-dictionaries-configuration/`: PCA dictionary templates. These remain in the package during the transition from dictionary-first to ingestion-spec-first workflows.
- `validated-artifacts/`: live-validated tenant dictionaries, deployed collector configuration, and golden samples captured during first lab validation.

## Object Model

### BGP Neighbor

Object type: `cisco-telemetry-xe-bgp-neighbor`

Identity:

- `source`
- `bgpAfiSafi`
- `bgpVrfName`
- `bgpNeighborId`

The stable `sessionId` uses the neighbor key: `source_bgpAfiSafi_bgpVrfName_bgpNeighborId`.

The human-facing `sessionName` uses the neighbor description when present: `source_bgpAfiSafi_bgpVrfName_bgpDescription`.

Metadata kept as dimensions:

- neighbor AS, link, description
- session state and connection state
- reset reason
- local and foreign host

Metrics focus on BGP message-count spikes, state/change context, and neighbor metadata:

- sent and received opens, updates, notifications, keepalives, and route refreshes
- input and output queue depth
- total established and total dropped connections
- installed prefixes and prefix activity counters
- negotiated keepalive timers
- transport ports and MSS

The Sensor Collector delta configuration converts only BGP churn counters to per-interval deltas:

- sent and received opens, updates, notifications, keepalives, and route refreshes
- sent and received implicit and explicit withdraw counters

It intentionally leaves current-state gauges as raw values, including current prefixes, bestpaths, multipaths, queue depths, timers, ports, and MSS. It also leaves `totalPrefixes` out of the delta list unless a later design explicitly chooses to expose route-arrival churn.

### Environment Sensor

Object type: `cisco-telemetry-xe-environment-sensor`

Identity:

- `source`

The stable `sessionId` and `sessionName` are `source_xe-environment-sensor`, for example `RM_01-CC-IT-01_xe-environment-sensor`.

Metadata kept as dimensions:

- source

Sensor identity is moved into gauge-style metric names. Each metric is built as `<sanitized name>_<sanitized location>`, where the router-emitted sensor `name` and `location` have spaces, colons, slashes, hyphens, and dots replaced by underscores. Location is always included, even when the sample data has no visible collision, because uniqueness can vary by router.

Examples:

- `Temp_UADP_0_0_R0`
- `Temp_InltFrnt_1_0`
- `HotSwap_Power_R0`

Threshold leaves are intentionally ignored in this supplemental demo model so the dashboard receives one aggregated environment object per router instead of one object per physical sensor.

### PTP

Object types:

- `cisco-telemetry-xe-ptp-clock`
- `cisco-telemetry-xe-ptp-parent`
- `cisco-telemetry-xe-ptp-port`
- `cisco-telemetry-xe-ptp-correction-stats`

Identities:

- clock and parent: stable `source + clock_domain` with type suffixes in `sessionId` and `sessionName`
- port: stable `source + if_name`
- correction-stats: stable `source + if_name`; do not include correction sample time in object identity

The correction-stat transform duplicates IOS XE `subordinate_port` into `if_name` so one PCA filter can select both PTP port and correction-stat objects for an interface.

Boolean PTP leaves are normalized upstream to `0` or `1`. IOS XE nanosecond leaves are kept as raw nanosecond fields in Telegraf and normalized in the ingestion spec/dictionary with explicit backtick casts. Final published PTP specs use microseconds for the nanosecond metrics:

```sql
cast(`raw_metric_name_ns` as DOUBLE) / 1000
```

The double-quoted expression pattern shown by the staging UI, for example `cast("field" as DOUBLE)`, did not apply correctly during validation.

## Validation Notes

This package has been validated against a live IOS XE lab and deployed to an existing `telemetry-collector_XE` managed collector.

The validated artifact bundle lives in:

```text
validated-artifacts/
```

Deployment result:

- existing managed collector: `telemetry-collector_XE`
- managed agent id: `1c94fd13-157b-4029-b8bf-92dd55fc8bc6`
- deployed field: `telemetry.dataTransformation`
- deployed config: `validated-artifacts/collector-deployment/telemetry-collector_XE.deployed.dataTransformation.conf`
- tenant dictionaries were published successfully for BGP neighbor and environment sensor object types
- tenant ingestion-staging specs were published successfully for the four PTP object types in the United Medley lab
- user confirmed the configuration works with a live data source

Standalone capture result:

The package was also smoke-tested by temporarily replacing the running `telemetry-collector_XE` listener on port `57000` with a disposable Telegraf 1.34 container, then restoring the original collector.

Golden sample result:

- source device observed: `RM_01-CC-IT-01`
- BGP subscription observed: `600`
- environment subscription observed: `700`
- environment path emitted as `Cisco-IOS-XE-environment-oper:environment-sensors/environment-sensor`
- BGP path emitted as `Cisco-IOS-XE-bgp-oper:bgp-state-data/neighbors/neighbor`
- PTP path emitted as `Cisco-IOS-XE-switch-ptp-oper:switch-ptp-oper-data`
- original `telemetry-collector_XE` was restored healthy after capture
- environment sample contained 41 sensor name/location pairs from one router/source; the aggregate model creates 1 environment session for that router with 41 current-reading metrics
- preserved sample archive: `validated-artifacts/golden-samples/xe-supplemental-telegraf-corrected-capture-20260601T151308Z.tar.gz`

Telegraf normalizes emitted IOS XE leaf/tag names by replacing hyphens with underscores. For example:

- YANG key `afi-safi` is emitted as `afi_safi`
- YANG key `neighbor-id` is emitted as `neighbor_id`
- `bgp-version` is emitted as `bgp_version`
- `current-reading` is emitted as `current_reading`

The collector field filters and Starlark metadata extraction use the emitted Telegraf names, not the raw YANG leaf spelling.

This is still community/demo status, not product promotion.

Before promoting beyond community/demo status:

1. Decide whether string leaves such as `up-time`, `last-read`, `last-write`, and `connection/last-reset` need a separate representation.
2. Decide whether the PCA dictionary should retain only the compact metric set or add additional diagnostic policy/capability fields from BGP.
3. Re-test with the final production packaging path before promoting into `cisco-pca-product/xe`.
