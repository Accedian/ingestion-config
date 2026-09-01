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
  - `cisco-telemetry-xr-policy`
- partial:
  - `cisco-telemetry-xr-dmm`

## Artifact Discipline

For this XR pack, artifact boundaries are intentional and should not be mixed:

- raw source-near captures live under `supporting-files/golden-samples/.../raw/`
- real post-telemetry-collector captures live under `supporting-files/golden-samples/.../transformed/`
- sensor-collector and dictionary validation may be documented from those artifacts, but must not be confused with telemetry-collector output

The current `l56-2026-03-16` bundle contains:

- raw source captures for interface, environment, IPSLA, Ethernet SLA, and policy
- real runtime telemetry-collector output for:
  - interface
  - environment
  - ipsla-icmp-echo
  - ipsla-udp-jitter
  - slm
  - dmm partial
  - policy

## Current Lab Constraint

`dmm` remains partial even after the post-NTP retest. The refreshed CLI, raw MDT, and live telemetry-collector output all show the same pattern:

- round-trip delay and round-trip jitter are sane
- one-way delay remains wrong in both directions
- `metric-one-way-jitter-ds` is present
- `metric-one-way-jitter-sd` is still absent at the raw MDT source

That means the remaining gap is source-side XR behavior for this probe/platform/release, not a telemetry-collector parsing issue.

## Policy Model Follow-Up

The current XR policy ingestion model folds `policy_in_stats` and `policy_out_stats` into one dictionary object type and relies on `direction` to separate them later.

That contract is now considered questionable and likely misaligned with real device behavior:

- input and output service-policies are usually independent attachments
- they often differ in policy name, class structure, counters, and operational meaning
- they should not be treated as two directional views of the same underlying policy object by default

Follow-up intent:

- revisit the XR policy ingestion design
- evaluate splitting the current single `cisco-telemetry-xr-policy` object into distinct input and output policy object types, or another model that preserves their independence
- align the telemetry-collector, sensor-collector, and dictionary layers with the real source-side policy model instead of forcing fold-together semantics

This is a design rework item for future pipeline cleanup. It is not part of the current validated state and should be handled deliberately rather than patched incrementally.

## Next Capture

Before further XR product work, refresh the saved `show run` state from both routers so the repository reflects the live post-validation configuration, including:

- `ipsla`
- `ethernet cfm`
- `ethernet sla`
- latest `telemetry model-driven` validation wiring

## Supplemental XR Models Promoted from PR #16

The supplemental IOS XR models introduced by PR #16 are now part of the product XR artifact set. Their dictionary definitions are in `pca-ingestion-dictionaries-configuration/`, and their raw captures and transformed validation samples are preserved under `supporting-files/`.

Promoted models:

- IPv4 BGP process information
- FIB VRF summary
- FIB drops for IPv4 and IPv6
- FIB exceptions for IPv4, IPv6, and MPLS
- IPv4 and IPv6 RIB summaries
- IPM receiver metrics
- power-management producer nodes
- aggregate interface statistics

The source PR describes these additions as exploratory and partially validated. They are therefore included with their evidence and collector wiring for product review; they are not represented as fully live-validated flows until the corresponding device and PCA checks are completed.
