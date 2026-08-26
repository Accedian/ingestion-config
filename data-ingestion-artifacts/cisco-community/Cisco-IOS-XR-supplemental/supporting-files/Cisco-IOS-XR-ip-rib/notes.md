# IP RIB Sample Attempts

These notes capture source-side query attempts on the L56 router used for XR supplemental telemetry testing.

## Router context

- Host: `lab@10.11.9.1`
- Telemetry subscription on the router includes both IPv4 and IPv6 BGP RIB paths.

## IPv4 BGP RIB path

Command:

```text
mdt_exec -c30000 -s Cisco-IOS-XR-ip-rib-ipv4-oper:rib/vrfs/vrf/afs/af/safs/saf/ip-rib-route-table-names/ip-rib-route-table-name/protocol/bgp/as/information
```

Observed result:

```text
Sub_id 200000001, flag 0, len 0
Sub_id 200000001, flag 8, len 0
```

Interpretation:

- The subscription path was accepted.
- No payload was returned from this router for this path at the time of capture.

## IPv6 BGP RIB path

Command:

```text
mdt_exec -c30000 -s Cisco-IOS-XR-ip-rib-ipv6-oper:ipv6-rib/vrfs/vrf/afs/af/safs/saf/ip-rib-route-table-names/ip-rib-route-table-name/protocol/bgp/as/information
```

Observed result:

```text
Sub_id 200000001, flag 0, len 0
Sub_id 200000001, flag 8, len 0
```

Interpretation:

- The subscription path was accepted.
- No payload was returned from this router for this path at the time of capture.

## Current conclusion

- The live collector config includes `ip_rib_ipv4` and `ip_rib_ipv6`, but the current router did not provide sample data for those paths.
- The next source-side step is to identify a router or topology with populated BGP RIB data for these YANG paths.

## CLI confirmation on L56

Closest operational CLI checks:

```text
show route vrf all ipv4 bgp wide
show route vrf all ipv6 bgp wide
show rib vrf all ipv4 protocols
show rib vrf all ipv6 protocols
```

Observed results:

- `show route vrf all ipv4 bgp wide` returned no matching routes
- `show route vrf all ipv6 bgp wide` returned no matching routes
- `show rib vrf all ipv4 protocols` showed only `local` and `connected`
- current operational data on this router does not support a populated BGP RIB sample for the target YANG paths
