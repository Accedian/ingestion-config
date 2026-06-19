# cisco-labs Optical / OTU Source Summary

This package starts from a current tenant-side `telemetry.dataCollection`
configuration captured during the telemetry collector custom configuration
census on 2026-06-17.

## Provenance

### Collector Configuration

- Tenant: `cisco-labs.analytics.accedian.io`
- Collector: `ofc2025-lab-telemetry-collector-10_89_205_253`
- Agent ID: `48ab384c-2832-4991-8983-281716d3d31f`
- Template name: `Cisco-Telemetry-IOS-XR-custom-optical`
- Template version: `2`
- Decoded field: `telemetry.dataCollection`
- Source evidence status: current tenant config

### Ingestion Dictionaries

- Tenant: `cisco-rcdn-lab.analytics.accedian.io`
- Fetch date: `2026-06-17`
- Endpoint scope: tenant ingestion dictionaries
- Source dictionaries:
  - `openmetrics-cisco_ios_xr_optics_info`
  - `openmetrics-cisco_ios_xr_otu_info`
- Sanitization: removed tenant ID, revision, and timestamp metadata before
  adding to this package.
- Bandwidth mapping note:
  - The collector config scopes `bandwidth` suffix trimming and float
    conversion to `cisco_ios_xr_otu_info`.
  - The referenced `openmetrics-cisco_ios_xr_optics_info` dictionary does not
    define a `cisco_ios_xr_optics_info_bandwidth` metric.
  - The referenced `openmetrics-cisco_ios_xr_otu_info` dictionary does define
    `cisco_ios_xr_otu_info_bandwidth` with unit `mbps` and a `sqlExpr` that
    multiplies the source value by `1000`.

### YANG Bandwidth Check

- Local Cisco XR YANG source checked under `/Users/etmartel/dev/yang/vendor/cisco/xr`.
- `Cisco-IOS-XR-controller-otu-oper.yang` maps `/otu/controllers/controller/info`
  to the `OTU-PER-PORT-INFO` grouping.
- `Cisco-IOS-XR-controller-otu-oper-sub1.yang` defines `leaf bandwidth` inside
  `OTU-PER-PORT-INFO`.
- `Cisco-IOS-XR-controller-optics-oper.yang` maps
  `/optics-oper/optics-ports/optics-port/optics-info` to the
  `OPTICS-EDM-INFO` grouping.
- No `leaf bandwidth` definition was found in local
  `Cisco-IOS-XR-controller-optics-oper*.yang` files.

## Collection Mode

The source config uses IOS XR MDT dial-out:

- Telegraf input: `inputs.cisco_telemetry_mdt`
- Transport: `tcp`
- Listener setting: `service_address`

The source config does not use gNMI dial-in and does not include device
username/password values, `servers`, or client-side TLS credential fields.

## Source Paths

- `Cisco-IOS-XR-controller-optics-oper:optics-oper/optics-ports/optics-port/optics-info`
- `Cisco-IOS-XR-controller-otu-oper:otu/controllers/controller/info`

## Object Types

- `cisco_ios_xr_optics_info`
- `cisco_ios_xr_otu_info`

## Secret Boundary

The raw tenant response and decoded source config remain in Etienne's private
AI workspace temporary census folder. This repository package contains only a
sanitized, parameterized derivative.
