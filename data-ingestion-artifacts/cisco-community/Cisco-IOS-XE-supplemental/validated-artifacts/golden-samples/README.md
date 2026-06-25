# IOS XE Supplemental Golden Samples

Golden samples captured from live IOS XE telemetry sources and sanitized for reusable publication. BGP/environment validation used an IOS XE lab source during the `2026-06-01` smoke test. PTP validation used a Catalyst 9300 lab source during `2026-06-12` through `2026-06-15`. A Catalyst 9300 current recapture on `2026-06-25` validates the current environment aggregate transform and the PTP correction-stats session identity. BGP was not expected in that Catalyst 9300 recapture; the earlier BGP smoke sample remains the BGP evidence.

The source uses Cisco IOS XE MDT / YANG Push dial-out into the XE telemetry collector port. For this protocol path, the preserved collector-ingress and transformed Telegraf output are the practical raw truth; no router-side `mdt_exec` style capture was available.

## Files

- `xe-supplemental-telegraf-corrected-capture-20260601T151308Z.tar.gz`
  - original corrected live capture bundle
  - includes `metrics.influx`, `telegraf.conf`, and `telegraf-container.log`
- `collector-ingress/xe-supplemental-live-metrics.influx`
  - full Telegraf-decoded live sample
  - `916` Influx line-protocol rows
- `xe-supplemental-c9300-current-capture-20260625.tar.gz`
  - current Catalyst 9300 capture bundle
  - includes `xe-supplemental-current.influx` and `xe-supplemental-current.json`
- `collector-ingress/xe-supplemental-c9300-current-live-metrics.influx`
  - full Telegraf-decoded Catalyst 9300 recapture
  - `752` Influx line-protocol rows
- `transformed/xe-bgp-neighbor.sample.influx`
  - filtered BGP neighbor rows
  - `8` rows
- `transformed/xe-environment-sensor.sample.influx`
  - filtered environment sensor rows from the `2026-06-25` Catalyst 9300 recapture
  - `48` rows
- `telegraf/xe-supplemental-standalone-capture-telegraf.conf`
  - standalone capture Telegraf configuration used for the smoke sample
- `proofs/xe-supplemental-standalone-capture-telegraf-container.log`
  - capture container log
- `xe-ptp-c9300-pca-prototype-2-capture-20260612.tar.gz`
  - preserved PTP prototype-2 capture bundle from the Catalyst 9300 lab
- `transformed/xe-ptp-normalized-new-run-only.sample.influx`
  - filtered normalized PTP PCA-shaped line protocol from the prototype-2 run
- `transformed/xe-ptp-normalized-new-run-only.sample.json`
  - same PTP normalized sample in JSON output format
- `telegraf/xe-ptp-managed-upload-telegraf.conf`
  - managed Telemetry Collector upload config used for the PTP validation path

## Measurements Observed

Original `2026-06-01` BGP/environment smoke capture:

- `xe_if_stats`: `728`
- `xe_memory_stats`: `12`
- `xe_bgp_neighbors`: `8`
- `xe_cpu_stats`: `4`

Current `2026-06-25` Catalyst 9300 recapture:

- `xe_if_stats`: `296`
- `xe_environment_sensor`: `48`
- `xe_memory_stats`: `9`
- `xe_cpu_stats`: `3`
- `xe_ptp_port`: `387`
- `xe_ptp_clock`: `3`
- `xe_ptp_parent`: `3`
- `xe_ptp_correction_stats`: `3`

PTP prototype-2 normalized sample:

- `xe_ptp_clock`
- `xe_ptp_parent`
- `xe_ptp_port`
- `xe_ptp_correction_stats`

## Object Identity Notes

BGP sessions are keyed by:

```text
source_bgpAfiSafi_bgpVrfName_bgpNeighborId
```

Environment sensors are aggregated into one session per router/source:

```text
source_xe-environment-sensor
```

The individual environment current readings are represented as metric names derived from sanitized sensor name plus location.

Environment sensor names are inventory-specific. Do not hard-code a fixed environment field whitelist in the collector transform; capture the real emitted names for the target platform and update PCA dictionary/profile artifacts from observed output only.

PTP clock and parent sessions are keyed by source plus clock domain with an object-type suffix. PTP port and correction-stat sessions are keyed by source plus `if_name`, with `subordinate_port` duplicated into `if_name` for correction stats so one PCA filter can select both port and correction-stat records for an interface.
