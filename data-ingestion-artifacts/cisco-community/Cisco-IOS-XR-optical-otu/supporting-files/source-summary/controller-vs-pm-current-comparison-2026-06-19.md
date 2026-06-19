# Cisco IOS XR Optical / OTU: Controller State vs PM Current

Date: 2026-06-19

## Summary

This package currently contains tenant-derived controller operational telemetry
for IOS XR optics and OTU, and a locally proven PM-current model for OTU/optics
performance-management data.

The two data families are complementary:

- Controller operational paths describe current object state, inventory,
  controller properties, alarms, instantaneous optical readings, and some BER/FEC
  values.
- PM-current paths describe the current 30-second performance bin: interval
  counters, min/average/max values, validity flags, individual TCA flags, and an
  aggregate per-session TCA flag.

## Source Families

| Family | Current package object type | Source path | Collection proven so far | Main use |
|---|---|---|---|---|
| Controller optics state | `cisco_ios_xr_optics_info` | `Cisco-IOS-XR-controller-optics-oper:optics-oper/optics-ports/optics-port/optics-info` | Tenant-derived MDT dial-out config | Optics inventory/state, frequency/wavelength, power, temperature, voltage, OSNR/CD/PMD style controller readings |
| Controller OTU state | `cisco_ios_xr_otu_info` | `Cisco-IOS-XR-controller-otu-oper:otu/controllers/controller/info` | Tenant-derived MDT dial-out config | OTU controller state, bandwidth, OTU alarm state/counters, FEC/BER controller values |
| PM OTU current FEC | `cisco-telemetry-xr-pm-otu-current-fec` | `Cisco-IOS-XR-pmengine-oper:performance-management/otu/.../otu-second30-fecs/otu-second30-fec` | Locally proven gNMI dial-in probe | Current 30s FEC counters and BER/Q/Q-margin min/average/max values |
| PM OTU current OTN | `cisco-telemetry-xr-pm-otu-current-otn` | `Cisco-IOS-XR-pmengine-oper:performance-management/otu/.../otu-second30-otns/otu-second30-otn` | Locally proven gNMI dial-in probe | Current 30s OTN counters: BBE, BBER, ES, ESR, FC, LBC, SES, SESR, UAS, near/far-end |
| PM optics current | `cisco-telemetry-xr-pm-optics-current-optics` | `Cisco-IOS-XR-pmengine-oper:performance-management/optics/.../optics-second30-optics/optics-second30-optic` | Locally proven gNMI dial-in probe | Current 30s optics min/average/max performance readings: CD, DGD, OPR/OPT, OSNR/SNR, PDL, PMD/SOPMD, Rx signal power |

## Configuration / Platform Notes

| Question | Controller optics/OTU state | PM-current OTU/optics |
|---|---|---|
| Requires a specific collector input? | Current package is MDT dial-out with `inputs.cisco_telemetry_mdt`. The same YANG paths may also be collectable by gNMI on suitable devices, but that is not what this package currently proves. | The local proof used gNMI dial-in with `inputs.gnmi`. Packaging with MDT dial-out has not yet been proven for the PM paths in this workstream. |
| Requires device subscriptions? | Yes for MDT dial-out: IOS XR must be configured to dial out the controller optics/OTU paths to the collector listener. | Yes for gNMI dial-in: collector must know the target address, credentials, encoding, and exact port/controller path keys. |
| Requires specific hardware? | Requires IOS XR platforms/interfaces that expose controller optics and/or OTU operational YANG. Non-optical ports will not emit meaningful optics/OTU objects. | More hardware-dependent. Requires optics/OTU ports that expose `Cisco-IOS-XR-pmengine-oper` current PM containers for the desired optics/OTU port. In the RCDN proof, coherent DSP/optics ports produced the data. |
| Requires PM feature enabled? | The controller optics dictionary includes `pmEnable`, but the controller-state path itself is not the PM-current bin. | Likely depends on PM engine support and the port having current PM counters active. The local proof confirms it on one RCDN IOS XR device/port; it does not prove all XR optics/OTU platforms expose the same PM-current paths by default. |
| Requires a license? | No license requirement was proven or identified from local YANG/package evidence. Needs Cisco platform/release confirmation before making a customer-facing claim. | No license requirement was proven or identified from local YANG/package evidence. Needs Cisco platform/release confirmation before making a customer-facing claim. |
| Best operational role | Baseline topology/object state, present/down/admin status, configured optical properties, current controller alarm state. | Performance operations and degradation tracking: current-bin OTN/FEC counters, optical min/avg/max, validity, TCA flags, aggregate TCA signal. |

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

## Package Direction

The safest package shape is to keep the controller-state and PM-current objects
as separate object families in one IOS XR optical/OTU package, but not to merge
their metrics into a single object type.

Recommended object families:

- `cisco_ios_xr_optics_info`
- `cisco_ios_xr_otu_info`
- `cisco-telemetry-xr-pm-otu-current-fec`
- `cisco-telemetry-xr-pm-otu-current-otn`
- `cisco-telemetry-xr-pm-optics-current-optics`

Open decision before live rollout:

- Whether the deployed collector should combine MDT dial-out controller state
  and gNMI dial-in PM-current collection in one managed agent config.
- Whether PM-current should be packaged as a separate collector config variant
  to keep rollback and credentials boundaries simpler.

## SMX / XpertsHub Answer Reconciliation

On 2026-06-19, Etienne asked SMX/XpertsHub in parallel and pasted the answer
into Codex. The SMX answer broadly aligns with this package direction:

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

Important nuance from SMX:

- SMX cited examples of `controller optics ... fastpoll enable` for high-rate
  controller optics sampling.
- SMX also cited examples of `pm 15-min optics report ...` and threshold CLI
  under controller optics.
- Those examples prove platform configuration knobs exist for fastpoll,
  PM reports, and threshold/TCA behavior, but they do not by themselves prove
  that 30s PM-current telemetry requires a synthetic test/probe or PRBS session.

Current local interpretation:

- The PM-current paths modeled here should not be described as active probing
  or testing configuration. They are PM engine operational data.
- Data availability still depends on platform, optics/OTU hardware, controller
  support, and whether PM is enabled/populated for the target port.
- The RCDN live probe proved that one IOS XR coherent optics/OTU port emitted
  OTU FEC, OTU OTN, and optics current 30s PM data without Codex configuring a
  new test/probe in that session.
- The same probe did not get optics FEC rows in the short run, which reinforces
  that PM subtree availability is per-platform/per-port/per-profile rather than
  universally guaranteed.

Licensing boundary:

- SMX did not find a license requirement in its retrieved sources.
- Local YANG/package evidence also did not identify a license gate.
- Therefore the safe statement is: no separate license requirement is proven
  from current evidence, but customer-facing claims need platform/release
  confirmation.

Hardware/platform boundary:

- SMX explicitly recommended platform-specific validation for NCS1004,
  NCS1010, NCS1014, or other targets.
- That matches this package's open validation item: exact hardware/card/optic
  support needs platform documentation or live target probing.
