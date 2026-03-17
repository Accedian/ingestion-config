# L56 Golden Sample Bundle

## Purpose

This bundle captures source-near and collector-near validation artifacts for the Cisco PCA product XR pack during A-to-Z validation against `l56`.

## Context

- Date: `2026-03-16`
- Git branch: `xr-product-a2z-validation-united-l56`
- Tenant baseline used for comparison: `united.medley.npav.accedian.net`
- Source router: `l56`
- Initial validation target: `cisco-telemetry-xr-interface`

## Bundle Layout

- `raw/`
  - source-near payloads or direct router-side captures
- `transformed/`
  - collector output after Telegraf shaping
- `router-state/`
  - source-side config, operational state, and platform context

## Notes

- This bundle is intended to preserve replayable validation evidence, not just screenshots or ad hoc notes.
- For this run, the `united` tenant baseline already matches the local product XR template and the 7 product XR global dictionaries.
- Any differences found during live validation should therefore be treated first as source-side availability or runtime shaping issues, not as tenant template drift.
- `show version` and `show run telemetry model-driven` were captured cleanly from `l56` and stored under `router-state/`.
- The temporary validation wiring that mirrored only the interface sensor path to port `47001` is stored under `router-state/show-run-telemetry-model-driven-with-validation.txt`.
- The expanded validation wiring that mirrored both interface and environment source paths to port `47001` is stored under `router-state/show-run-telemetry-model-driven-with-interface-and-environment-validation.txt`.
- A live `mdt_exec` sample for the interface sensor path was confirmed on-box, but the noninteractive extraction path still needs refinement before the full raw payload is committed under `raw/`.
- A first transformed product-shaped interface sample is stored under `transformed/interface-sample.jsonl`.
- A first transformed product-shaped environment sample is stored under `transformed/environment-sample.jsonl`.
- Additional post-telemetry-collector transformed samples are now stored under `transformed/` for:
  - `ipsla-icmp-echo`
  - `ipsla-udp-jitter`
  - `slm`
  - `dmm` partial
- These are intentionally telemetry-collector-shaped artifacts only.
- They are kept separate from any future post-sensor-collector artifacts so the two boundaries do not get mixed.
- The current IPSLA/SLM/DMM transformed files are now backed by a real runtime Telegraf capture from a separate `l56` validation subscription to VM port `47002`, not only by reconstruction from raw MDT payloads.
- Environment raw source-near artifacts are now present under `raw/` for both:
  - `Cisco-IOS-XR-wdsysmon-fd-oper:system-monitoring/cpu-utilization`
  - `Cisco-IOS-XR-nto-misc-oper:memory-summary/nodes/node/summary`
- Current validated-complete coverage on `l56`:
  - `cisco-telemetry-xr-interface`
  - `cisco-telemetry-xr-environment`
  - `cisco-telemetry-xr-ipsla-icmp-echo`
  - `cisco-telemetry-xr-ipsla-udp-jitter`
  - `cisco-telemetry-xr-slm`
- Current partial coverage on `l56`:
  - `cisco-telemetry-xr-dmm`

## IPSLA Validation Result

- The live source-side IPSLA subtree on `l56` is:
  - `Cisco-IOS-XR-man-ipsla-oper:ipsla/operation-data/operations/operation/statistics/latest/target`
- One source sample covers both operation types:
  - `icmp-echo`
  - `udp-jitter`
- The validation target for these objects is the full product path:
  - raw MDT payload
  - Telegraf normalization
  - sensor-collector relabel and rename logic
  - dictionary contract
- Important result:
  - the saved IPSLA sample content matches the sensor-collector and dictionary contract for both XR IPSLA objects
  - `cisco-telemetry-xr-ipsla-icmp-echo`: required dictionary raw metrics matched `14/14`
  - `cisco-telemetry-xr-ipsla-udp-jitter`: required dictionary raw metrics matched `24/24` after applying the sensor-collector directional rename rules
- The `-` vs `_` concern was specifically checked:
  - raw MDT uses hyphenated keys such as `op-type` and `udp-jitter-stats`
  - Telegraf normalizes those into underscore-based field paths such as `specific_stats/op_type` and `specific_stats/udp_jitter_stats/...`
  - sensor-collector rules are written against that underscore form and collapse directional variants into the canonical dictionary raw names
- The saved raw artifact is now present on disk and reparses correctly from the local environment:
  - `raw/Cisco-IOS-XR-man-ipsla-oper_ipsla_operation-data_operations_operation_statistics_latest_target.raw`

## Ethernet SLA Validation Result

- The live source-side Ethernet SLA subtree on `l56` is:
  - `Cisco-IOS-XR-infra-sla-oper:sla/protocols/Cisco-IOS-XR-ethernet-cfm-oper:ethernet/statistics-historicals/statistics-historical`
- One saved source sample covers both operation types:
  - `cfm-delay-measurement`
  - `cfm-synthetic-loss-measurement`
- The validation target for these objects is the full product path:
  - raw MDT payload
  - Telegraf normalization
  - sensor-collector relabel and rename logic
  - dictionary contract
- Important result:
  - `cisco-telemetry-xr-slm`: complete through the dictionary boundary
  - `cisco-telemetry-xr-dmm`: partial only

### SLM Contract Result

- The saved raw sample includes both directional FLR metric families:
  - `metric-one-way-flr-sd`
  - `metric-one-way-flr-ds`
- The sensor-collector rules collapse those into the canonical dictionary-facing names and assign directions `0` and `1`.
- Result:
  - `cisco-telemetry-xr-slm`: required dictionary base metrics matched `5/5`
  - `cisco-telemetry-xr-slm`: required custom metric is computable `1/1`
  - `packetsLostPct` is derivable from `slmSent` and `packetsLost`
- Saved raw artifact:
  - `raw/Cisco-IOS-XR-infra-sla-oper_sla_protocols_Cisco-IOS-XR-ethernet-cfm-oper_ethernet_statistics-historicals_statistics-historical.raw`

### DMM Contract Result

- The saved raw sample includes:
  - `metric-round-trip-delay`
  - `metric-one-way-delay-sd`
  - `metric-one-way-delay-ds`
  - `metric-round-trip-jitter`
  - `metric-one-way-jitter-ds`
- The saved raw sample does not include:
  - `metric-one-way-jitter-sd`
- Consequences:
  - direction `1` one-way jitter metrics are structurally present
  - direction `0` one-way jitter metrics are not proven from this artifact
  - the Telegraf-derived `delay_var_avg_rt` helper cannot be produced from this sample because it depends on both `sd` and `ds` jitter averages
- Result:
  - `cisco-telemetry-xr-dmm` is not a `100%` dictionary-contract validation from this source sample
  - round-trip delay and the visible jitter families are structurally valid
  - one-way delay values are also clearly impacted by bad source clock state, so data quality is not trustworthy even though the subtree itself is correct
