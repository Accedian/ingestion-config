# Cisco IOS XR Supplemental

This file is a mx of experiments of things I want to add to the product eventually.

Some of these have been requested/suggest by Caroline Araujo. Other by Amish Servansing.

Work in progress \<insert gif of stickman with hardhat and yellow flashing lights from geocities\>

- Caroline: ipv4_bgp, fib_vrf_summary, ip_rib_ipv4, ip_rib_ipv6, 
- Amish: ipm, power_management_producer_node

Most of these are working, the main gap is the inability to convert the IPM histogram into something hat makes sense because of the lack of identifying information for each bin and the fact that the bin themselves will vary in size and value. 

## Progress Registry

This README is the working progress registry for this supplemental integration area.

Status values:

- `not started`
- `in progress`
- `partial`
- `ready for validation`
- `blocked`
- `done`

### Overall status

- Project state: `partial`
- Current maturity: validated Telegraf shaping and golden samples for several models, incomplete PCA artifact set
- Main gap: PCA dictionaries, sensor-collector decisions where needed, and final productization for the remaining non-FIB supplemental models

### Model progress

| Model / Object | Requested by | Sample data | YANG available locally | Telegraf alias | Object identity in Telegraf | PCA dictionary | Sensor Collector logic | Validation maturity | Status | Notes |
|---|---|---|---|---|---|---|---|---|---|---|
| `ipv4_bgp` | Caroline | yes | yes | yes | yes | first-pass only | not identified yet | raw sample plus field inventory exist, but no saved transformed sample in the `l57` bundle | `partial` | first-pass dictionary exists in `pca-ingestion-dictionaries-configuration/openmetrics-cisco-telemetry-xr-ipv4-bgp.json`, but it is derived from local raw/whitelist evidence rather than a live tenant dictionary object |
| `fib_vrf_summary` | Caroline | yes | yes | yes | yes | yes | not identified yet | validated raw and transformed samples from `l57` | `partial` | first-pass dictionary copied from live `ipm-demo` reference set into `pca-ingestion-dictionaries-configuration/openmetrics-cisco-telemetry-xr-fib-vrf-summary.json` |
| `fib_drops_v4` | backlog `P7` | yes | yes | yes | yes | yes | likely not needed | validated transformed sample from `l56`; node-centric object model confirmed | `done` | path `Cisco-IOS-XR-fib-common-oper:fib-statistics/nodes/node/drops`; keeps YANG leaves `1:1` and adds `total_drop_packets` |
| `fib_drops_v6` | backlog `P7` | yes | yes | yes | yes | yes | likely not needed | validated transformed sample from `l56`; node-centric object model confirmed | `done` | path `Cisco-IOS-XR-fib-common-oper:fib-statistics/nodes/node/drops-v6`; keeps YANG leaves `1:1` and adds `total_drop_packets` |
| `fib_exceptions_v4` | backlog `P7` | yes | yes | yes | yes | yes | likely not needed | validated transformed sample from `l56`; node-centric object model confirmed | `done` | path `Cisco-IOS-XR-fib-common-oper:fib-statistics/nodes/node/exceptions-v4`; keeps YANG leaves `1:1` and adds `total_exception_packets` |
| `fib_exceptions_v6` | backlog `P7` | yes | yes | yes | yes | yes | likely not needed | validated transformed sample from `l56`; node-centric object model confirmed | `done` | path `Cisco-IOS-XR-fib-common-oper:fib-statistics/nodes/node/exceptions-v6`; keeps YANG leaves `1:1` and adds `total_exception_packets` |
| `fib_exceptions_mpls` | backlog `P7` | yes | yes | yes | yes | yes | likely not needed | validated transformed sample from `l56`; node-centric object model confirmed | `done` | path `Cisco-IOS-XR-fib-common-oper:fib-statistics/nodes/node/exceptions-mpls`; keeps YANG leaves `1:1` and adds `total_exception_packets` |
| `ip_rib_ipv4` | Caroline | yes | yes | yes | yes | yes | likely not needed | validated raw and transformed samples from `l57` | `partial` | the old `.../protocol/bgp/as/information` path was the wrong source for this object; the working source is `Cisco-IOS-XR-ip-rib-ipv4-oper:rib/rib-table-ids/rib-table-id/summary-protos/summary-proto`; first-pass dictionary copied from live `ipm-demo` reference set |
| `ip_rib_ipv6` | Caroline | yes | yes | yes | yes | yes | likely not needed | validated raw and transformed samples from `l57` | `partial` | the old `.../protocol/bgp/as/information` path was the wrong source for this object; the working source is `Cisco-IOS-XR-ip-rib-ipv6-oper:ipv6-rib/rib-table-ids/rib-table-id/summary-protos/summary-proto`; first-pass dictionary copied from live `ipm-demo` reference set |
| `ipm` | Amish | yes | yes | yes | yes | yes | likely needed | validated transformed sample from `l57`, but histogram semantics remain unresolved | `partial` | first-pass dictionary exists and is being normalized to local naming conventions; unresolved gap is still the broader histogram/object-model question rather than basic collector output |
| `power_management_producer_node` | Amish | yes | yes | yes | yes | yes | likely not needed | validated transformed sample from `l57` | `partial` | first-pass dictionary copied from live `ipm-demo` reference set; Telegraf includes embedded tags and unit conversion logic |

### Artifact status

| Artifact area | Status | Notes |
|---|---|---|
| `telemetry-collector-configuration/telegraf.conf` | `partial` | substantial working draft exists, including validated `ip_rib` `summary-proto` shaping and done-state FIB node drops/exceptions shaping |
| `pca-ingestion-dictionaries-configuration/` | `partial` | first-pass supplemental dictionaries now exist for `fib_vrf_summary`, the five FIB node object types, `interface_aggregate`, `ipm`, `ipv4_rib_summary`, `ipv6_rib_summary`, `power_management`, and a provisional `ipv4_bgp` |
| `sensor-collector-configuration/` | `not started` | not present yet |
| `supporting-files/` | `partial` | useful samples exist for several models, including dated `l57` golden samples for `ip_rib_ipv4` and `ip_rib_ipv6`, plus an `l56` bundle for the five FIB node object types |
| `working-directory/` | `not started` | currently empty |

### Next recommended steps

1. Review the newly added supplemental dictionaries against the live `ipm-demo` reference set and local transformed samples.
2. Continue the analytics-name cleanup pass so the supplemental dictionaries use intentional KPI names instead of mechanically repeated source-family names.
3. Capture a real transformed `ipv4_bgp` sample so the first-pass local dictionary can be validated or corrected.
4. Confirm whether `ipm` needs additional object splitting or successor dictionaries because of the histogram semantics gap.
5. Decide whether any supplemental model needs Sensor Collector delta/rate logic.
6. Add a repeatable validation workflow summary for the validated `l57` collector path on `57001`.
7. Promote the completed FIB node objects into whichever downstream packaging path is chosen for backlog `P7`.

### Interface Aggregate Note

`interface_aggregate` is intentionally kept in supplemental.

It is not a duplicate per-interface XR object like the product `cisco-telemetry-xr-interface` model.
It is a derived rollup built from `if_stats` that keeps only input and output data-rate fields, then computes node-wide `count`, `min`, `max`, `mean`, and `sum` over the window.
