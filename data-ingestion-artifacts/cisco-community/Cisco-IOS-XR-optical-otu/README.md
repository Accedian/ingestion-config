# Cisco IOS XR Optical / OTU Telemetry

## Scope

This package preserves and sanitizes tenant-proven Cisco IOS XR optical and
OTU telemetry collector configuration, then extends the package direction with
a locally proven PM-current model for IOS XR optics and OTU performance data.

- Source tenant evidence: `cisco-labs.analytics.accedian.io`
- Source collector: `ofc2025-lab-telemetry-collector-10_89_205_253`
- Source agent ID: `48ab384c-2832-4991-8983-281716d3d31f`
- Source template: `Cisco-Telemetry-IOS-XR-custom-optical` v2
- Original collection mode: IOS XR MDT dial-out to
  `inputs.cisco_telemetry_mdt`
- PM-current proof mode: local standalone Telegraf gNMI dial-in probe against
  the same RCDN router family used during package research
- Artifact status: initial sanitized package, not yet fully
  dictionary-validated

## Source Paths

The original tenant collector aliases these IOS XR controller operational
paths:

- `Cisco-IOS-XR-controller-optics-oper:optics-oper/optics-ports/optics-port/optics-info`
- `Cisco-IOS-XR-controller-otu-oper:otu/controllers/controller/info`

The original collector emits two object families:

- `cisco_ios_xr_optics_info`
- `cisco_ios_xr_otu_info`

The PM-current research adds these IOS XR PM engine 30-second current families:

- `Cisco-IOS-XR-pmengine-oper:performance-management/otu/.../otu-second30-fecs/otu-second30-fec`
- `Cisco-IOS-XR-pmengine-oper:performance-management/otu/.../otu-second30-otns/otu-second30-otn`
- `Cisco-IOS-XR-pmengine-oper:performance-management/optics/.../optics-second30-optics/optics-second30-optic`

The intended PM object families are:

- `cisco-telemetry-xr-pm-otu-current-fec`
- `cisco-telemetry-xr-pm-otu-current-otn`
- `cisco-telemetry-xr-pm-optics-current-optics`

## Data Family Summary

The controller and PM-current families are complementary:

- Controller operational paths describe current object state, inventory,
  controller properties, alarms, instantaneous optical readings, and some
  BER/FEC values.
- PM-current paths describe the current 30-second performance bin: interval
  counters, min/average/max values, validity flags, individual TCA flags, and
  an aggregate per-session TCA flag.

| Family | Package object type | Source path | Collection proven so far | Main use |
|---|---|---|---|---|
| Controller optics state | `cisco_ios_xr_optics_info` | `Cisco-IOS-XR-controller-optics-oper:optics-oper/optics-ports/optics-port/optics-info` | Tenant-derived MDT dial-out config | Optics inventory/state, frequency/wavelength, power, temperature, voltage, OSNR/CD/PMD style controller readings |
| Controller OTU state | `cisco_ios_xr_otu_info` | `Cisco-IOS-XR-controller-otu-oper:otu/controllers/controller/info` | Tenant-derived MDT dial-out config | OTU controller state, bandwidth, OTU alarm state/counters, FEC/BER controller values |
| PM OTU current FEC | `cisco-telemetry-xr-pm-otu-current-fec` | `Cisco-IOS-XR-pmengine-oper:performance-management/otu/.../otu-second30-fecs/otu-second30-fec` | Locally proven gNMI dial-in probe | Current 30s FEC counters and BER/Q/Q-margin min/average/max values |
| PM OTU current OTN | `cisco-telemetry-xr-pm-otu-current-otn` | `Cisco-IOS-XR-pmengine-oper:performance-management/otu/.../otu-second30-otns/otu-second30-otn` | Locally proven gNMI dial-in probe | Current 30s OTN counters: BBE, BBER, ES, ESR, FC, LBC, SES, SESR, UAS, near/far-end |
| PM optics current | `cisco-telemetry-xr-pm-optics-current-optics` | `Cisco-IOS-XR-pmengine-oper:performance-management/optics/.../optics-second30-optics/optics-second30-optic` | Locally proven gNMI dial-in probe | Current 30s optics min/average/max performance readings: CD, DGD, OPR/OPT, OSNR/SNR, PDL, PMD/SOPMD, Rx signal power |

## Controller State vs PM Current

| Question | Controller optics/OTU state | PM-current OTU/optics |
|---|---|---|
| Primary purpose | Real-time controller state, inventory, alarms, and sensor readings. | Performance operations and degradation tracking from current 30s PM bins. |
| Collector input proven here | MDT dial-out with `inputs.cisco_telemetry_mdt`. | gNMI dial-in with `inputs.gnmi` in a standalone local probe. |
| Device-side requirement | IOS XR must be configured to stream the controller optics/OTU paths to the collector when using MDT dial-out. | Collector must know target address, credentials, encoding, and exact port/controller path keys when using gNMI dial-in. |
| Hardware dependency | Requires IOS XR platforms/interfaces that expose controller optics and/or OTU operational YANG. Non-optical ports will not emit meaningful optics/OTU objects. | More hardware-dependent. Requires optics/OTU ports that expose `Cisco-IOS-XR-pmengine-oper` current PM containers. In the RCDN proof, coherent DSP/optics ports produced the data. |
| PM enablement dependency | The controller optics dictionary includes `pmEnable`, but the controller-state path itself is not the PM-current bin. | Likely depends on PM engine support and the port having current PM counters active. The local proof confirms one RCDN IOS XR device/port; it does not prove all XR optics/OTU platforms expose the same PM-current paths by default. |
| License dependency | No license requirement was proven or identified from local YANG/package evidence or the SMX answer. Platform/release confirmation is still required before making customer-facing claims. | No license requirement was proven or identified from local YANG/package evidence or the SMX answer. Platform/release confirmation is still required before making customer-facing claims. |
| Best operational role | Baseline topology/object state, present/down/admin status, configured optical properties, current controller alarm state. | Current-bin OTN/FEC counters, optical min/avg/max values, validity, TCA flags, and aggregate TCA signal. |

## Metric Overlap And Differences

| Area | Present in controller optics/OTU | Present in PM-current OTU/optics | Interpretation |
|---|---|---|---|
| Object identity | Yes: `sessionId`, `sessionName`, `objectType`, source/name/controller dimensions. | Yes: same identity pattern, with `cisco-telemetry-xr-*` object types and sanitized session IDs. | Common PCA modeling ground. |
| Operational state | Strong: `state`, `controllerState`, `laserState`, presence/config flags, OTU alarms. | Weak: PM validity flags only; not a controller state source. | Keep controller objects for availability/state. |
| Bandwidth / inventory | Strong on OTU: `bandwidth`, local/remote IDs, support/config flags. | Not present in current PM model. | PM should not replace controller OTU inventory. |
| Optical power/current readings | Strong instantaneous/controller readings: Tx/Rx power, mW, dBm, voltage, temperature, wavelength/frequency. | Strong interval readings for selected performance parameters: OPR, OPT, Rx signal power min/avg/max. | PM adds interval behavior; controller adds broader physical details. |
| Optical quality | Controller optics includes CD, OSNR, PMD/SOPMD-style values, Q/Q margin through OTU/controller paths. | PM optics adds current 30s min/avg/max for CD, DGD, OSNR/SNR, PDL, PCR, PMD/SOPMD, low signal frequency offset, LBC. | PM is better for trend/degradation analysis. |
| FEC / BER | Controller OTU/optics includes pre/post FEC BER and uncorrected BER style current values. | PM OTU FEC adds EC bits, UC words, pre/post FEC BER min/avg/max, Q and Q-margin min/avg/max, validity and TCA flags. | PM FEC is the richer operational performance source. |
| OTN counters | Controller OTU has many OTU alarm counters/states. | PM OTU OTN has current-bin BBE/BBER/ES/ESR/FC/LBC/SES/SESR/UAS near/far-end counters and TCA flags. | These are largely unique to PM-current and are important for service performance monitoring. |
| Alarms / TCA | Controller OTU has detailed alarm detected/asserted/reporting state for many OTU alarm types. | PM-current has individual `*_tca_report` flags and aggregate `tca_report`; thresholds are modeled as dynamic metadata candidates. | Alarm state and PM threshold crossing are different signals; both are useful. |
| Threshold values | Controller package mostly treats configuration/state metrics as normal metrics. | PM model moves threshold/configured-threshold values out of metrics and treats them as dynamic metadata candidates. | Avoids polluting KPIs while preserving context. |
| Time-bin behavior | No. Mostly current controller state. | Yes. Current 30s bin only; 15m intentionally excluded because PCA can aggregate. | PM-current is the right source for time-windowed performance KPIs. |

## SMX / XpertsHub Reconciliation

On 2026-06-19, SMX/XpertsHub was asked in parallel about these YANG families.
The answer broadly aligns with this package direction:

- `controller-optics-oper:.../optics-info` is the real-time optics state and
  sensor source.
- `controller-otu-oper:.../controller/info` is the OTU controller operational
  source.
- `pmengine-oper:performance-management/optics` and
  `pmengine-oper:performance-management/otu` are the PM sources for intervaled
  optics and OTU counters/readings.
- The PM engine exposes 30s, 15m, and 24h interval families; this package
  intentionally models only the 30s current family because PCA can handle
  downstream aggregation.

SMX cited examples of `controller optics ... fastpoll enable` for high-rate
controller optics sampling and examples of `pm 15-min optics report ...` plus
threshold CLI under controller optics. Those examples prove platform
configuration knobs exist for fastpoll, PM reports, and threshold/TCA behavior,
but they do not by themselves prove that 30s PM-current telemetry requires a
synthetic test, probe, or PRBS session.

Current package interpretation:

- The modeled PM-current paths are PM engine operational data, not active
  probing or testing configuration.
- Data availability still depends on platform, optics/OTU hardware, controller
  support, and whether PM is enabled/populated for the target port.
- The RCDN live probe proved that one IOS XR coherent optics/OTU port emitted
  OTU FEC, OTU OTN, and optics current 30s PM data without Codex configuring a
  new test/probe in that session.
- The same probe did not get optics FEC rows in the short run, which reinforces
  that PM subtree availability is per-platform/per-port/per-profile rather than
  universally guaranteed.
- No separate license requirement is proven from current local or SMX evidence,
  but customer-facing claims need platform/release confirmation.

SMX recommended platform-specific validation for NCS1004, NCS1010, NCS1014, or
other targets. That remains an open package validation item: exact
hardware/card/optic support needs platform documentation or live target probing.

## Included Files

- `telemetry-collector-configuration/cisco-telemetry-xr-optical-otu.conf`
  - sanitized Telegraf configuration for MDT dial-out collection;
  - tenant-specific listener and source-host mappings replaced with placeholders.
- `pca-ingestion-dictionaries-configuration/cisco-ios-xr-optics-info.json`
  - initial PCA dictionary for optics operational metrics.
- `pca-ingestion-dictionaries-configuration/cisco-ios-xr-otu-info.json`
  - initial PCA dictionary for OTU/FEC metrics.
- `supporting-files/source-summary/cisco-labs-optical-otu-source-summary.md`
  - sanitized provenance and proof-boundary note.
- `supporting-files/source-summary/controller-vs-pm-current-comparison-2026-06-19.md`
  - supporting detail for the controller-state vs PM-current comparison and
    SMX reconciliation summarized in this README.

## Deployment Notes

Configure IOS XR model-driven telemetry to dial out to the telemetry
collector listener configured in `service_address`.

The cisco-labs source configuration used plain TCP MDT dial-out. It did not
use gNMI dial-in, device username/password values, or client-side TLS material.

The `processors.enum` source-to-host-name mapping is intentionally shown as an
example. Replace it with customer/lab source addresses or remove the mapping if
the original source tag is the desired host identity.

The PM-current work was proven separately with standalone Telegraf gNMI
dial-in. Packaging the PM-current paths into the same managed collector
configuration as the controller-state MDT dial-out paths is an open rollout
decision, not yet a validated deployment pattern in this package.

## Validation Boundary

This package starts from a current tenant-side collector configuration. The
collector config is tenant-proven, but the dictionary JSON files are an initial
repo normalization pass and still need PCA ingestion-staging validation before
they should be treated as final production dictionaries.

The PM-current model is locally proven from a short RCDN probe, but has not yet
been converted into final package dictionaries or validated across target
platforms. Before customer-facing use, confirm platform, IOS XR release,
card/optic support, PM enablement behavior, and any licensing constraints for
the intended deployment platform.
