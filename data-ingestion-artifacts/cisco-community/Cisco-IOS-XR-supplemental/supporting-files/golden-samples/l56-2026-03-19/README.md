# l56 FIB Node Golden Sample Bundle

## Purpose

This bundle captures the validated transformed outputs for the five node-level XR FIB telemetry object types proven on `l56` on `2026-03-19`.

The intended use is:

1. preserve the first live transformed samples for these object types
2. document the node-centric object identity and total-field logic
3. provide a replay/reference point when refining the supplemental dictionaries

## Capture Metadata

- Capture date: `2026-03-19`
- Capture window:
  - first transformed metric timestamp: `2026-03-19 18:02:48 UTC`
  - first transformed output write: `2026-03-19 18:02:50 UTC`
- Lab: `SJC Jakub's lab`
- Router identity: `l56`
- Collector host: `sjc_ipm_jakubs_lab_pca_collectors`
- Validation path:
  - standalone Telegraf container
  - host port `47002`
  - source router `l56` streamed directly to the validation collector

## Source Paths Intended on l56

- `Cisco-IOS-XR-fib-common-oper:fib-statistics/nodes/node/drops`
- `Cisco-IOS-XR-fib-common-oper:fib-statistics/nodes/node/drops-v6`
- `Cisco-IOS-XR-fib-common-oper:fib-statistics/nodes/node/exceptions-v4`
- `Cisco-IOS-XR-fib-common-oper:fib-statistics/nodes/node/exceptions-v6`
- `Cisco-IOS-XR-fib-common-oper:fib-statistics/nodes/node/exceptions-mpls`

## Raw Artifacts Intended

Paste `mdt_exec` output into these files under `raw/`:

- [Cisco-IOS-XR-fib-common-oper_drops.raw](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l56-2026-03-19/raw/Cisco-IOS-XR-fib-common-oper_drops.raw)
- [Cisco-IOS-XR-fib-common-oper_drops-v6.raw](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l56-2026-03-19/raw/Cisco-IOS-XR-fib-common-oper_drops-v6.raw)
- [Cisco-IOS-XR-fib-common-oper_exceptions-v4.raw](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l56-2026-03-19/raw/Cisco-IOS-XR-fib-common-oper_exceptions-v4.raw)
- [Cisco-IOS-XR-fib-common-oper_exceptions-v6.raw](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l56-2026-03-19/raw/Cisco-IOS-XR-fib-common-oper_exceptions-v6.raw)
- [Cisco-IOS-XR-fib-common-oper_exceptions-mpls.raw](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l56-2026-03-19/raw/Cisco-IOS-XR-fib-common-oper_exceptions-mpls.raw)

## Business Logic Confirmed

- object identity is node-centric: `source + node_name + object family`
- raw YANG leaves are preserved `1:1`
- `node_name` is normalized from XR slash format to PCA-safe underscore format
- each object gets a scoped computed total field:
  - `total_drop_packets` for `fib_drops_v4` and `fib_drops_v6`
  - `total_exception_packets` for the three exceptions objects

## Transformed Artifacts Present

- [fib_drops_v4.transformed.jsonl](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l56-2026-03-19/fib_drops_v4.transformed.jsonl)
- [fib_drops_v6.transformed.jsonl](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l56-2026-03-19/fib_drops_v6.transformed.jsonl)
- [fib_exceptions_v4.transformed.jsonl](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l56-2026-03-19/fib_exceptions_v4.transformed.jsonl)
- [fib_exceptions_v6.transformed.jsonl](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l56-2026-03-19/fib_exceptions_v6.transformed.jsonl)
- [fib_exceptions_mpls.transformed.jsonl](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l56-2026-03-19/fib_exceptions_mpls.transformed.jsonl)
- [l56-fib-node-validation-metrics.jsonl](/Users/etmartel/git/ingestion-config/data-ingestion-artifacts/cisco-community/Cisco-IOS-XR-supplemental/supporting-files/golden-samples/l56-2026-03-19/l56-fib-node-validation-metrics.jsonl)

## Counts

- `fib_drops_v4`: 1 line
- `fib_drops_v6`: 1 line
- `fib_exceptions_v4`: 1 line
- `fib_exceptions_v6`: 1 line
- `fib_exceptions_mpls`: 1 line

## Validation Notes

- `fib_drops_v4` and `fib_drops_v6` were both live and non-zero on `l56`
- the three exception object types emitted successfully but were all-zero at capture time
- this is sufficient to consider the shaping and dictionary design done for these five objects
