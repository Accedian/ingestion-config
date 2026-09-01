# IP RIB Summary-Proto Metric Review

This file reviews the `summary-proto` samples captured from:

- `Cisco-IOS-XR-ip-rib-ipv4-oper:rib/rib-table-ids/rib-table-id/summary-protos/summary-proto`
- `Cisco-IOS-XR-ip-rib-ipv6-oper:ipv6-rib/rib-table-ids/rib-table-id/summary-protos/summary-proto`

Goal:

- keep high-value performance and route-scale metrics
- stay comfortably below the `40` KPI soft limit
- avoid burning KPI budget on families that are currently always `0`

## Recommended object shape

Suggested object split:

- `cisco-telemetry-xr-ipv4-rib-summary`
- `cisco-telemetry-xr-ipv6-rib-summary`

Suggested dimensions:

- `source`
- `node_id_str`
- `tableid`
- `protoid`
- `name`
- `instance` when present

## Core metrics

These are the strongest keep candidates.

| Raw metric | Proposed metric name | Values seen in raw samples | Keep/Drop | Reason |
|---|---|---:|---|---|
| `proto-route-count.active-routes-count` | `activeRoutesCount` | `0, 1, 4, 17` | `Keep` | Primary route scale KPI. |
| `proto-route-count.num-active-paths` | `activePathsCount` | `0, 1, 4, 31` | `Keep` | High-value path scale KPI. |
| `proto-route-count.num-backup-routes` | `backupRoutesCount` | `0, 1` | `Keep` | Useful resilience signal with low metric cost. |
| `proto-route-count.num-backup-paths` | `backupPathsCount` | `0, 1` | `Keep` | Useful resilience signal with low metric cost. |

## Conditional metrics

These are structurally valid, but only worth keeping if you confirm they are non-zero in real target topologies. In the current samples, the BGP-specific families are all `0`.

| Raw metric | Proposed metric name | Values seen in raw samples | Keep/Drop | Reason |
|---|---|---:|---|---|
| `rtype-bgp-int.active-routes-count` | `bgpInternalRoutesCount` | `0` | `Drop for now` | Good semantic KPI, but zero in current samples. Promote later if a real BGP-populated router shows signal. |
| `rtype-bgp-int.num-active-paths` | `bgpInternalPathsCount` | `0` | `Drop for now` | Same reason. |
| `rtype-bgp-int.num-backup-routes` | `bgpInternalBackupRoutesCount` | `0` | `Drop for now` | Same reason. |
| `rtype-bgp-int.num-backup-paths` | `bgpInternalBackupPathsCount` | `0` | `Drop for now` | Same reason. |
| `rtype-bgp-ext.active-routes-count` | `bgpExternalRoutesCount` | `0` | `Drop for now` | Good future KPI, no current signal. |
| `rtype-bgp-ext.num-active-paths` | `bgpExternalPathsCount` | `0` | `Drop for now` | Same reason. |
| `rtype-bgp-ext.num-backup-routes` | `bgpExternalBackupRoutesCount` | `0` | `Drop for now` | Same reason. |
| `rtype-bgp-ext.num-backup-paths` | `bgpExternalBackupPathsCount` | `0` | `Drop for now` | Same reason. |
| `rtype-bgp-loc.active-routes-count` | `bgpLocalRoutesCount` | `0` | `Drop for now` | Good future KPI, no current signal. |
| `rtype-bgp-loc.num-active-paths` | `bgpLocalPathsCount` | `0` | `Drop for now` | Same reason. |
| `rtype-bgp-loc.num-backup-routes` | `bgpLocalBackupRoutesCount` | `0` | `Drop for now` | Same reason. |
| `rtype-bgp-loc.num-backup-paths` | `bgpLocalBackupPathsCount` | `0` | `Drop for now` | Same reason. |
| `rtype-isis-l2.active-routes-count` | `isisL2RoutesCount` | `0, 17` | `Keep` | This one has real signal in both samples and helps explain route distribution by protocol family. |
| `rtype-isis-l2.num-active-paths` | `isisL2PathsCount` | `0, 31` | `Keep` | Same reason. |
| `rtype-isis-l2.num-backup-routes` | `isisL2BackupRoutesCount` | `0` | `Drop for now` | No current signal. |
| `rtype-isis-l2.num-backup-paths` | `isisL2BackupPathsCount` | `0` | `Drop for now` | No current signal. |

## Low-value or zero-only metrics

These were all `0` across the current IPv4 and IPv6 samples and should not be included in the first-pass object.

| Raw metric family | Proposed action | Reason |
|---|---|---|
| `rtype-none.*` | `Drop` | No observed signal and low product value. |
| `rtype-other.*` | `Drop` | No observed signal and ambiguous meaning. |
| `rtype-ospf-intra.*` | `Drop` | No observed signal. |
| `rtype-ospf-inter.*` | `Drop` | No observed signal. |
| `rtype-ospf-extern1.*` | `Drop` | No observed signal. |
| `rtype-ospf-extern2.*` | `Drop` | No observed signal. |
| `rtype-ospf-nssa1.*` | `Drop` | No observed signal. |
| `rtype-ospf-nssa2.*` | `Drop` | No observed signal. |
| `rtype-isis-sum.*` | `Drop` | No observed signal. |
| `rtype-isis-l1.*` | `Drop` | No observed signal. |
| `rtype-isis-l1-ia.*` | `Drop` | No observed signal. |
| `rtype-igrp2-int.*` | `Drop` | No observed signal. |
| `rtype-igrp2-ext.*` | `Drop` | No observed signal. |

## First-pass KPI shortlist

If we want a compact and defensible first version, this is the recommended set:

| Proposed metric name | Keep/Drop |
|---|---|
| `activeRoutesCount` | `Keep` |
| `activePathsCount` | `Keep` |
| `backupRoutesCount` | `Keep` |
| `backupPathsCount` | `Keep` |
| `isisL2RoutesCount` | `Keep` |
| `isisL2PathsCount` | `Keep` |

This produces a very small object with clear meaning and room to grow later.

## Expanded but still safe shortlist

If you want a broader object while staying well under `50` KPIs, keep the compact set above and add the BGP-specific counters only after you capture a topology where they are non-zero:

| Proposed metric name | Keep/Drop now | Promotion rule |
|---|---|---|
| `bgpInternalRoutesCount` | `Drop for now` | Add once raw samples show non-zero values. |
| `bgpInternalPathsCount` | `Drop for now` | Add once raw samples show non-zero values. |
| `bgpExternalRoutesCount` | `Drop for now` | Add once raw samples show non-zero values. |
| `bgpExternalPathsCount` | `Drop for now` | Add once raw samples show non-zero values. |
| `bgpLocalRoutesCount` | `Drop for now` | Add once raw samples show non-zero values. |
| `bgpLocalPathsCount` | `Drop for now` | Add once raw samples show non-zero values. |

## Current recommendation

Preferred direction now:

- keep a single mixed `summary-proto` object per AF:
  - `cisco-telemetry-xr-ipv4-rib-summary`
  - `cisco-telemetry-xr-ipv6-rib-summary`
- keep the four `proto-route-count` totals
- append family/type names to family-specific metrics
- do not split into separate object types per routing family
- do not gate object creation by family-specific non-zero logic

Why this is preferred:

- it stays well under the KPI soft limit
- it matches the raw model more directly
- it avoids multiplying object types
- it avoids ambiguity about whether generic total counters are really family-scoped
- it keeps the analytics schema straightforward

## Considered But Not Preferred: One Object Per Routing Family

This was considered seriously, but it is no longer the preferred direction.

### Idea

Instead of one mixed `ipv4_rib_summary` or `ipv6_rib_summary` object carrying many family-specific fields, create separate object types per address family and routing family, for example:

- `cisco-telemetry-xr-ipv4-rib-summary-bgp`
- `cisco-telemetry-xr-ipv4-rib-summary-ospf`
- `cisco-telemetry-xr-ipv4-rib-summary-isis`
- `cisco-telemetry-xr-ipv6-rib-summary-bgp`
- `cisco-telemetry-xr-ipv6-rib-summary-ospf`
- `cisco-telemetry-xr-ipv6-rib-summary-isis`

and also encode the routing family as a tag:

- `routingFamily=bgp`
- `routingFamily=ospf`
- `routingFamily=isis`

### Why this is attractive

- metric names become generic and reusable
- database metric sprawl is reduced because field names do not need family prefixes
- each object is semantically tighter and easier to explain
- KPI budgets become easier to manage per family
- future expansion is cleaner because adding a new family does not pollute the field set of every other family

### Example metric naming change

Instead of:

- `bgpInternalRoutesCount`
- `bgpExternalRoutesCount`
- `isisL2RoutesCount`

use:

- `internalRoutesCount`
- `externalRoutesCount`
- `l2RoutesCount`

within an object whose type and tags already identify the routing family.

### Example shape

For a BGP family object:

- objectType: `cisco-telemetry-xr-ipv4-rib-summary-bgp`
- tags:
  - `source`
  - `node_id_str`
  - `tableid`
  - `af=ipv4`
  - `routingFamily=bgp`
  - `instance` when present

candidate metrics:

- `activeRoutesCount`
- `activePathsCount`
- `backupRoutesCount`
- `backupPathsCount`
- `internalRoutesCount`
- `internalPathsCount`
- `externalRoutesCount`
- `externalPathsCount`
- `localRoutesCount`
- `localPathsCount`

For an ISIS family object:

- objectType: `cisco-telemetry-xr-ipv4-rib-summary-isis`
- tags:
  - `source`
  - `node_id_str`
  - `tableid`
  - `af=ipv4`
  - `routingFamily=isis`

candidate metrics:

- `activeRoutesCount`
- `activePathsCount`
- `backupRoutesCount`
- `backupPathsCount`
- `l1RoutesCount`
- `l1PathsCount`
- `l2RoutesCount`
- `l2PathsCount`
- `l1IaRoutesCount`
- `l1IaPathsCount`
- `summaryRoutesCount`
- `summaryPathsCount`

For an OSPF family object:

- objectType: `cisco-telemetry-xr-ipv4-rib-summary-ospf`
- tags:
  - `source`
  - `node_id_str`
  - `tableid`
  - `af=ipv4`
  - `routingFamily=ospf`

candidate metrics:

- `activeRoutesCount`
- `activePathsCount`
- `backupRoutesCount`
- `backupPathsCount`
- `intraRoutesCount`
- `intraPathsCount`
- `interRoutesCount`
- `interPathsCount`
- `extern1RoutesCount`
- `extern1PathsCount`
- `extern2RoutesCount`
- `extern2PathsCount`
- `nssa1RoutesCount`
- `nssa1PathsCount`
- `nssa2RoutesCount`
- `nssa2PathsCount`

### Tradeoffs

Benefits:

- clearer object semantics
- cleaner field namespace
- easier downstream reuse of metric names
- less pressure to create long family-prefixed analytics names

Costs:

- more object types
- some duplication of the generic total counters across families
- requires Telegraf logic to split the mixed `summary-proto` feed into separate family-specific objects

### Why it is not preferred now

- the top-level `proto-route-count.*` counters are easier to reason about in a mixed summary object
- splitting by family creates more object types without enough payoff at the current KPI scale
- family-split gating logic in Telegraf adds complexity that is not necessary for the current design
- the mixed object still leaves enough room to keep valuable family-specific counters with explicit names

## Final direction for now

Use one summary object per address family and keep family names in the metrics that need them.

Example metric naming:

- `activeRoutesCount`
- `activePathsCount`
- `backupRoutesCount`
- `backupPathsCount`
- `bgpInternalRoutesCount`
- `bgpInternalPathsCount`
- `bgpExternalRoutesCount`
- `bgpExternalPathsCount`
- `bgpLocalRoutesCount`
- `bgpLocalPathsCount`
- `isisL1RoutesCount`
- `isisL1PathsCount`
- `isisL2RoutesCount`
- `isisL2PathsCount`
- `isisL1IaRoutesCount`
- `isisL1IaPathsCount`
- `ospfIntraRoutesCount`
- `ospfIntraPathsCount`
- `ospfInterRoutesCount`
- `ospfInterPathsCount`
- `ospfExtern1RoutesCount`
- `ospfExtern1PathsCount`
- `ospfExtern2RoutesCount`
- `ospfExtern2PathsCount`
- `ospfNssa1RoutesCount`
- `ospfNssa1PathsCount`
- `ospfNssa2RoutesCount`
- `ospfNssa2PathsCount`
