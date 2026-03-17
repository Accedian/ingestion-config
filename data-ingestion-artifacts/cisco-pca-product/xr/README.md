# Cisco PCA Product XR

## Progress Note

This pack now has real A-to-Z validation evidence for most XR product flows in the `l56` / `l57` lab pairing.

Current validation status:

- complete:
  - `cisco-telemetry-xr-interface`
  - `cisco-telemetry-xr-environment`
  - `cisco-telemetry-xr-ipsla-icmp-echo`
  - `cisco-telemetry-xr-ipsla-udp-jitter`
  - `cisco-telemetry-xr-slm`
- partial:
  - `cisco-telemetry-xr-dmm`
- not yet validated:
  - `cisco-telemetry-xr-policy`

## Artifact Discipline

For this XR pack, artifact boundaries are intentional and should not be mixed:

- raw source-near captures live under `supporting-files/golden-samples/.../raw/`
- real post-telemetry-collector captures live under `supporting-files/golden-samples/.../transformed/`
- sensor-collector and dictionary validation may be documented from those artifacts, but must not be confused with telemetry-collector output

The current `l56-2026-03-16` bundle contains:

- raw source captures for interface, environment, IPSLA, and Ethernet SLA
- real runtime telemetry-collector output for:
  - interface
  - environment
  - ipsla-icmp-echo
  - ipsla-udp-jitter
  - slm
  - dmm partial

## Current Lab Constraint

`dmm` is only partial in this lab because the routers have severely wrong clocks. The subtree and runtime shaping are valid, but one-way delay quality is not trustworthy and the observed runtime sample does not provide full one-way jitter coverage.

## Next Capture

Before further XR product work, refresh the saved `show run` state from both routers so the repository reflects the live post-validation configuration, including:

- `ipsla`
- `ethernet cfm`
- `ethernet sla`
- latest `telemetry model-driven` validation wiring
