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
- Current maturity: strong Telegraf experimentation, incomplete PCA artifact set
- Main gap: PCA dictionaries, repeatable validation, and final productization

### Model progress

| Model / Object | Requested by | Sample data | YANG available locally | Telegraf alias | Object identity in Telegraf | PCA dictionary | Sensor Collector logic | Validation maturity | Status | Notes |
|---|---|---|---|---|---|---|---|---|---|---|
| `ipv4_bgp` | Caroline | yes | yes | yes | yes | no | not identified yet | partial | `partial` | rawName exploration exists in `supporting-files/Cisco-IOS-XR-ipv4-bgp-oper/rawName.txt` |
| `fib_vrf_summary` | Caroline | yes | yes | yes | yes | no | not identified yet | partial | `partial` | sample exists in `supporting-files/Cisco-IOS-XR-fib-common-oper/` |
| `ip_rib_ipv4` | Caroline | no, query returned empty payload on L56 | yes | yes | not obvious yet | no | not identified yet | source-side gap confirmed on current router | `blocked` | telemetry subscription exists and `mdt_exec` against `Cisco-IOS-XR-ip-rib-ipv4-oper:rib/vrfs/vrf/afs/af/safs/saf/ip-rib-route-table-names/ip-rib-route-table-name/protocol/bgp/as/information` returned `len 0`; `show route vrf all ipv4 bgp wide` also returned no matching routes |
| `ip_rib_ipv6` | Caroline | no, query returned empty payload on L56 | yes | yes | not obvious yet | no | not identified yet | source-side gap confirmed on current router | `blocked` | telemetry subscription exists and `mdt_exec` against `Cisco-IOS-XR-ip-rib-ipv6-oper:ipv6-rib/vrfs/vrf/afs/af/safs/saf/ip-rib-route-table-names/ip-rib-route-table-name/protocol/bgp/as/information` returned `len 0`; `show route vrf all ipv6 bgp wide` also returned no matching routes |
| `ipm` | Amish | not obvious in this folder | yes | yes | yes | no | likely needed | blocked | `blocked` | histogram semantics unresolved |
| `power_management_producer_node` | Amish | not obvious in this folder | yes | yes | yes | no | likely not needed | partial | `in progress` | Telegraf includes embedded tags and unit conversion logic |

### Artifact status

| Artifact area | Status | Notes |
|---|---|---|
| `telemetry-collector-configuration/telegraf.conf` | `partial` | substantial working draft exists |
| `pca-ingestion-dictionaries-configuration/` | `not started` | directory exists but is empty |
| `sensor-collector-configuration/` | `not started` | not present yet |
| `supporting-files/` | `partial` | useful samples exist for some models |
| `working-directory/` | `not started` | currently empty |

### Next recommended steps

1. Find a router or topology where the IPv4 and IPv6 BGP RIB YANG paths actually return populated data.
2. Confirm supporting sample files for `ipm` and `power_management_producer_node`.
3. Derive per-model candidate dictionaries from the existing Telegraf output and YANG models.
4. Decide whether any model needs Sensor Collector delta/rate logic.
5. Add a repeatable validation workflow with golden samples.
6. Populate `pca-ingestion-dictionaries-configuration/`.
