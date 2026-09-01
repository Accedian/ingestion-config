# l57 Golden Sample Bundle

## Purpose

This bundle captures a replayable validation snapshot for the `l57` IOS XR source in SJC Jakub's lab.

The intended golden sample structure is:

1. source-near raw payloads, as close as possible to the router output and before Telegraf business logic
2. transformed outputs after Telegraf aliasing, filtering, renaming, tag construction, and object/session logic
3. metadata describing where, when, and under what conditions the capture was taken

## Capture Metadata

- Capture date: `2026-03-15`
- Capture window:
  - first transformed metric timestamp: `2026-03-15 18:28:57 UTC`
  - last transformed metric timestamp: `2026-03-15 18:42:00 UTC`
- Lab: `SJC Jakub's lab`
- Router identity: `l57`
- Collector host: `sjc_ipm_jakubs_lab_pca_collectors`
- Validation path:
  - vanilla Telegraf container
  - host port `57001`
  - source router `l57` streamed directly to the validation collector
- Important lab condition:
  - `pca-collectors` bridge `br0` had to reacquire global IPv6 after setting `net.ipv6.conf.br0.accept_ra = 2`
  - without that, `l57` could not reach the collector over IPv6

## Router Metadata

Confirmed:

- Source router hostname: `l57`
- OS family: `Cisco IOS XR`
- OS version: `25.2.1.18I LNT`
- Build version: `25.2.1.18I`
- Platform family: `cisco 8000`
- Platform model: `Cisco 8201-24H8FH`
- Hardware description: `Cisco 8200 1RU w/24x100G & 8x400G, XR SW & HBM`
- CPU: `Intel(R) Xeon(R) CPU D-1530 @ 2.40GHz`
- Memory: `32GB`
- Router uptime at capture time: `26 weeks, 4 days, 3 hours, 1 minute`
- Telemetry source used model-driven telemetry over gRPC/TCP dialout

Source of truth:

```text
RP/0/RP0/CPU0:l57#show version
Cisco IOS XR Software, Version 25.2.1.18I LNT
Version      : 25.2.1.18I
cisco 8000
cisco 8201-24H8FH processor with 32GB of memory
Cisco 8200 1RU w/24x100G & 8x400G, XR SW & HBM
```

## Source Paths Intended on l57

The active source configuration for this work included:

- `Cisco-IOS-XR-pfi-im-cmd-oper:interfaces/interface-xr/interface`
- `Cisco-IOS-XR-envmon-oper:power-management/rack/producers/producer-nodes/producer-node`
- `Cisco-IOS-XR-fib-common-oper:fib/nodes/node/protocols/protocol/vrfs/vrf/summary`
- `Cisco-IOS-XR-ipv4-bgp-oper:bgp/instances/instance/instance-active/default-vrf/process-info`
- `Cisco-IOS-XR-perf-meas-oper:performance-measurement/nodes/node/endpoints/ep-ipm-receivers/ep-ipm-receiver-metrics`
- `Cisco-IOS-XR-ip-rib-ipv4-oper:rib/rib-table-ids/rib-table-id/summary-protos/summary-proto`
- `Cisco-IOS-XR-ip-rib-ipv6-oper:ipv6-rib/rib-table-ids/rib-table-id/summary-protos/summary-proto`

## Raw Artifacts Present

These are the source-near raw artifacts currently included in `raw/`:

- `ep-ipm-receiver-metric.raw`
- `producer-node.raw`
- `Cisco-IOS-XR-ip-rib-ipv4-oper_summary-proto.raw`
- `Cisco-IOS-XR-ip-rib-ipv6-oper_summary-proto.raw`
- `interface.raw`
- `fib_vrf_summary.raw`

## Raw Gaps

The following source paths do not yet have matching raw artifacts inside this bundle:

- `ipv4_bgp`

Notes:

- `aggregate_interface` is not a router source path. It is a derived Telegraf aggregate and therefore does not need a raw router artifact.
- `ipv4_bgp` was configured on the router but did not appear in the transformed capture window saved here.

## Transformed Artifacts Present

These transformed outputs were captured from the validation Telegraf pipeline:

- [ep_ipm_receiver_metric.transformed.jsonl](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l57-2026-03-15/ep_ipm_receiver_metric.transformed.jsonl)
- [power_management_producer_node.transformed.jsonl](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l57-2026-03-15/power_management_producer_node.transformed.jsonl)
- [fib_vrf_summary.transformed.jsonl](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l57-2026-03-15/fib_vrf_summary.transformed.jsonl)
- [if_stats.transformed.jsonl](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l57-2026-03-15/if_stats.transformed.jsonl)
- [ip_rib_ipv4.transformed.jsonl](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l57-2026-03-15/ip_rib_ipv4.transformed.jsonl)
- [ip_rib_ipv6.transformed.jsonl](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l57-2026-03-15/ip_rib_ipv6.transformed.jsonl)
- [aggregate_interface.transformed.jsonl](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l57-2026-03-15/aggregate_interface.transformed.jsonl)

Batch source:

- [l57-validation-metrics.jsonl](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l57-2026-03-15/l57-validation-metrics.jsonl)

## Counts

- `ep_ipm_receiver_metric`: 11 lines
- `power_management_producer_node`: 22 lines
- `fib_vrf_summary`: 55 lines
- `if_stats`: 374 lines
- `ip_rib_ipv4`: 99 lines
- `ip_rib_ipv6`: 99 lines
- `aggregate_interface`: 11 lines

## Replay Intent

This bundle is meant to support future offline replay and skill-building work.

For a fully replayable end-to-end bundle, the remaining improvements are:

1. capture raw router-side payload for `ipv4_bgp`
2. preserve the exact Telegraf config revision used for this transformed capture

## if_stats Validation Note

`if_stats` is already implemented in the product XR collector config at:

- [cisco-telemetry-mdt.conf](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-pca-product/xr/telemetry-collector-configuration/cisco-telemetry-mdt.conf)

Validation against `l57` shows the raw sample is consistent with the product assumptions:

- source path: `Cisco-IOS-XR-pfi-im-cmd-oper:interfaces/interface-xr/interface`
- transformed rate fields come from `content.data-rates`
- interface identity exists in raw via `interface-name` and related fields

Current conclusion:

- `if_stats` is a validation/reference path right now, not a new build task in the XR supplemental workflow
