# XR Source State Intake

Drop the current-state command outputs for `l56` and `l57` into the matching files in this directory.

Required files:
- `show-run.txt`
- `show-run-telemetry-model-driven.txt`
- `show-run-interface.txt`
- `show-ipv4-interface-brief.txt`
- `show-ipv6-interface-brief.txt`
- `show-run-ethernet-cfm.txt`
- `show-run-ipsla.txt`
- `show-l2vpn-bridge-domain-brief.txt`

Helpful if present:
- `show-run-service-policy.txt`
- `show-run-router-static.txt`
- `show-run-router-isis.txt`
- `show-run-router-ospf.txt`

## Post-Validation Refresh

After live validation changes are made on the routers, also preserve the refreshed running config separately so the repository captures the final applied state rather than only the pre-validation intake.

Preferred files:
- `l56/show-run-post-validation.txt`
- `l56/show-run-telemetry-model-driven-post-validation.txt`
- `l57/show-run-post-validation.txt`
- `l57/show-run-telemetry-model-driven-post-validation.txt`
