# IP RIB Summary-Proto Object Model Note

This note captures the current intended object model for the XR `summary-proto` path so it can be translated into `telegraf.conf` next.

## Status

Design intent only. Not implemented yet.

## Input paths

- IPv4:
  - `Cisco-IOS-XR-ip-rib-ipv4-oper:rib/rib-table-ids/rib-table-id/summary-protos/summary-proto`
- IPv6:
  - `Cisco-IOS-XR-ip-rib-ipv6-oper:ipv6-rib/rib-table-ids/rib-table-id/summary-protos/summary-proto`

## Selected object strategy

Use one mixed summary object per address family:

- `cisco-telemetry-xr-ipv4-rib-summary`
- `cisco-telemetry-xr-ipv6-rib-summary`

Do not split into separate object types per routing family.

Do keep routing-family detail in the field names where needed.

## Object instance rule

Create one object instance per:

- `node_id_str`
- `tableid`
- `name`
- optional `instance`
- address family suffix

Canonical construction:

```text
<node_id_str>_<tableid>_<name>[_<instance>]_ip-rib-ipvN_summary-proto
```

Examples:

- `l57_e0800000_vxlan_ip-rib-ipv4_summary-proto`
- `l57_e0800000_local-srv6_sidmgr_ip-rib-ipv4_summary-proto`

## Object creation gate

Create the object only when the four `proto-route-count` counters for that specific record are not all zero.

The four gating counters are:

- `proto-route-count.active-routes-count`
- `proto-route-count.num-active-paths`
- `proto-route-count.num-backup-routes`
- `proto-route-count.num-backup-paths`

Gate logic:

- if all four are `0`, drop the record and do not create an object
- if any one of the four is non-zero, create the object

Examples:

- `l57_e0800000_vxlan_ip-rib-ipv4_summary-proto`
  - all four `proto-route-count.*` values are `0`
  - result: no object created

- `l57_e0800000_local-srv6_sidmgr_ip-rib-ipv4_summary-proto`
  - `active-routes-count = 1`
  - result: object created

## Why this gate is preferred

- it operates on the actual record instance, not on a routing-family subview
- it suppresses empty objects cleanly
- it preserves the mixed-summary object strategy
- it avoids the complexity of family-specific object creation logic

## Session tags

Proposed tag construction:

- `sessionName = <node_id_str>_<tableid>_<name>[_<instance>]_ip-rib-ipvN_summary-proto`
- `sessionId = <node_id_str>_<tableid>_<name>[_<instance>]_ip-rib-ipvN_summary-proto`
- `objectType = cisco-telemetry-xr-ipvN-rib-summary`
- `direction = -1`

`sessionName` and `sessionId` can be identical for this object family unless a stronger stable machine identifier becomes available.

## Tag set

Keep these tags when present:

- `source`
- `node_id_str`
- `tableid`
- `protoid`
- `name`
- `instance`

`instance` should only be included when present and non-empty.

## Normalization rules

Before using `name` or `instance` in session tags:

- replace `/` with `_`
- replace spaces with `_`
- replace `:` with `_`
- keep only characters compatible with the workspace ingestion rules: `[a-zA-Z0-9_-]`

The goal is to keep `sessionName`, `sessionId`, and `objectType` valid without losing the operator-readable meaning.

## KPI schema direction

Keep the four generic totals:

- `activeRoutesCount`
- `activePathsCount`
- `backupRoutesCount`
- `backupPathsCount`

Keep family-specific counters with the family encoded in the field name.

Examples:

- BGP:
  - `bgpInternalRoutesCount`
  - `bgpInternalPathsCount`
  - `bgpInternalBackupRoutesCount`
  - `bgpInternalBackupPathsCount`
  - `bgpExternalRoutesCount`
  - `bgpExternalPathsCount`
  - `bgpExternalBackupRoutesCount`
  - `bgpExternalBackupPathsCount`
  - `bgpLocalRoutesCount`
  - `bgpLocalPathsCount`
  - `bgpLocalBackupRoutesCount`
  - `bgpLocalBackupPathsCount`

- ISIS:
  - `isisL1RoutesCount`
  - `isisL1PathsCount`
  - `isisL1BackupRoutesCount`
  - `isisL1BackupPathsCount`
  - `isisL2RoutesCount`
  - `isisL2PathsCount`
  - `isisL2BackupRoutesCount`
  - `isisL2BackupPathsCount`
  - `isisL1IaRoutesCount`
  - `isisL1IaPathsCount`
  - `isisL1IaBackupRoutesCount`
  - `isisL1IaBackupPathsCount`
  - `isisSummaryRoutesCount`
  - `isisSummaryPathsCount`
  - `isisSummaryBackupRoutesCount`
  - `isisSummaryBackupPathsCount`

- OSPF:
  - `ospfIntraRoutesCount`
  - `ospfIntraPathsCount`
  - `ospfIntraBackupRoutesCount`
  - `ospfIntraBackupPathsCount`
  - `ospfInterRoutesCount`
  - `ospfInterPathsCount`
  - `ospfInterBackupRoutesCount`
  - `ospfInterBackupPathsCount`
  - `ospfExtern1RoutesCount`
  - `ospfExtern1PathsCount`
  - `ospfExtern1BackupRoutesCount`
  - `ospfExtern1BackupPathsCount`
  - `ospfExtern2RoutesCount`
  - `ospfExtern2PathsCount`
  - `ospfExtern2BackupRoutesCount`
  - `ospfExtern2BackupPathsCount`
  - `ospfNssa1RoutesCount`
  - `ospfNssa1PathsCount`
  - `ospfNssa1BackupRoutesCount`
  - `ospfNssa1BackupPathsCount`
  - `ospfNssa2RoutesCount`
  - `ospfNssa2PathsCount`
  - `ospfNssa2BackupRoutesCount`
  - `ospfNssa2BackupPathsCount`

## Telegraf implementation intent

When this is implemented in `telegraf.conf`, the processing sequence should be:

1. ingest `summary-proto`
2. normalize tags such as `name` and `instance`
3. evaluate the object-creation gate using the four `proto-route-count.*` counters
4. drop records whose four gating counters are all `0`
5. create `sessionName`, `sessionId`, `objectType`, and `direction`
6. keep the approved field set with `fieldinclude`
7. optionally rename raw fields to the final analytics-facing names

## Open questions for implementation

- whether to keep every family-specific backup counter in the first pass or stage them in later
- whether `protoid` should remain a tag once `name` is present
- whether any families with consistent zero-only behavior should be omitted from the first dictionary even if the field names are reserved here
