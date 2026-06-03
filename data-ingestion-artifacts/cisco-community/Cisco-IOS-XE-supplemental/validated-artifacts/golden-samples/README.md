# IOS XE Supplemental Golden Samples

Golden sample captured from a live IOS XE telemetry source during the 2026-06-01 smoke test. The first validation source was the Open Fiber PCA lab tenant.

The source uses Cisco IOS XE MDT / YANG Push dial-out into the XE telemetry collector port. For this protocol path, the preserved collector-ingress and transformed Telegraf output are the practical raw truth; no router-side `mdt_exec` style capture was available.

## Files

- `xe-supplemental-telegraf-corrected-capture-20260601T151308Z.tar.gz`
  - original corrected live capture bundle
  - includes `metrics.influx`, `telegraf.conf`, and `telegraf-container.log`
- `collector-ingress/xe-supplemental-live-metrics.influx`
  - full Telegraf-decoded live sample
  - `916` Influx line-protocol rows
- `transformed/xe-bgp-neighbor.sample.influx`
  - filtered BGP neighbor rows
  - `8` rows
- `transformed/xe-environment-sensor.sample.influx`
  - filtered environment sensor rows
  - `164` rows
- `telegraf/xe-supplemental-standalone-capture-telegraf.conf`
  - standalone capture Telegraf configuration used for the smoke sample
- `proofs/xe-supplemental-standalone-capture-telegraf-container.log`
  - capture container log

## Measurements Observed

- `xe_if_stats`: `728`
- `xe_environment_sensor`: `164`
- `xe_memory_stats`: `12`
- `xe_bgp_neighbors`: `8`
- `xe_cpu_stats`: `4`

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
