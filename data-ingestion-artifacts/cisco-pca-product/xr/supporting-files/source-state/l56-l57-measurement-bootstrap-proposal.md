# l56/l57 Measurement Bootstrap Proposal

This proposal is derived from the current source-state captures for `l56` and `l57` plus the local IOS XR YANG models:

- `Cisco-IOS-XR-um-ipsla-cfg.yang`
- `Cisco-IOS-XR-man-ipsla-cfg.yang`
- `Cisco-IOS-XR-um-ethernet-cfm-cfg.yang`

## What the current state supports

- `l56` and `l57` both have stable management IPv4 addresses:
  - `l56`: `10.255.18.156`
  - `l57`: `10.255.18.157`
- `l56` and `l57` both have stable loopback IPv6 addresses:
  - `l56`: `fcbb:bb00:56::1/128`
  - `l57`: `fcbb:bb00:57::1/128`
- Neither router currently has:
  - `ipsla`
  - `ethernet cfm`
  - an existing bridge-domain or xconnect service for Ethernet SLA

## Important constraint

`IPSLA icmp-echo` and `IPSLA udp-jitter` can be paired directly between `l56` and `l57` using the management IPv4 addresses.

`DMM` and `SLM` require a real Ethernet CFM service with peer MEPs on the same MA/MEG.

So the recommendation is:

- use `l56 <-> l57` directly for `ipsla-icmp-echo`
- use `l56 <-> l57` directly for `ipsla-udp-jitter`
- use the DMM/SLM templates below only after picking a confirmed directly connected Ethernet segment and replacing the interface names accordingly

## Proposed IPSLA bootstrap

These two operations are designed to create telemetry for:

- `cisco-telemetry-xr-ipsla-icmp-echo`
- `cisco-telemetry-xr-ipsla-udp-jitter`

### l56

```text
ipsla
 operation 561
  type icmp echo
   source address 10.255.18.156
   destination address 10.255.18.157
   timeout 1000
   frequency 1
  !
 !
 operation 562
  type udp jitter
   source address 10.255.18.156
   destination address 10.255.18.157
   packet count 10
   packet interval 20
   timeout 1000
   source port 37372
   datasize request 32
   destination port 14000
   frequency 1
   verify-data
  !
 !
 schedule operation 561
  start-time now
  life forever
 !
 schedule operation 562
  start-time now
  life forever
 !
 responder
  type udp ipv4 address 10.255.18.156 port 14001
 !
!
```

### l57

```text
ipsla
 operation 571
  type icmp echo
   source address 10.255.18.157
   destination address 10.255.18.156
   timeout 1000
   frequency 1
  !
 !
 operation 572
  type udp jitter
   source address 10.255.18.157
   destination address 10.255.18.156
   packet count 10
   packet interval 20
   timeout 1000
   source port 37373
   datasize request 32
   destination port 14001
   frequency 1
   verify-data
  !
 !
 schedule operation 571
  start-time now
  life forever
 !
 schedule operation 572
  start-time now
  life forever
 !
 responder
  type udp ipv4 address 10.255.18.157 port 14000
 !
!
```

## Proposed DMM/SLM bootstrap

This is now grounded in a proven shared Layer 2 path between:

- `l56 TenGigE0/0/0/4.1103`
- `l57 TenGigE0/0/0/1`

Evidence:

- `l56` base MAC on `TenGigE0/0/0/4` is `ec19.2eea.b914`
- `l57 show ipv6 neighbors` sees `ec19.2eea.b914` on `TenGigE0/0/0/1`
- `l57` base MAC on `TenGigE0/0/0/1` is `9088.554c.1c08`
- `l56 show ipv6 neighbors` sees `9088.554c.1c08` on `TenGigE0/0/0/4.1103`
- `l56 TenGigE0/0/0/4.1103` carries only:
  - `encapsulation dot1q 1103`
  - `ipv6 enable`
- `l57 TenGigE0/0/0/1` carries only:
  - `ipv6 enable`

Operational interpretation:

- VLAN `1103` from `l56` is being bridged by `jakubtor40`
- that same service is handed to `l57 TenGigE0/0/0/1`, likely as an untagged/native service

That makes this the preferred candidate for Ethernet CFM validation.

The product collector expects these probe types:

- `cfm-delay-measurement` -> `cisco-telemetry-xr-dmm`
- `cfm-synthetic-loss-measurement` -> `cisco-telemetry-xr-slm`

### Important caution

- this appears to be an existing live service, not a disposable validation-only segment
- use isolated `PCA-VALIDATION-*` names
- validate the exact CLI interactively with `?` before commit on box
- verify CCM adjacency first, then DMM, then SLM

### Candidate config for l56

```text
ethernet sla
 profile PCA-DMM type cfm-delay-measurement
  probe
   send packet every 1 seconds
  !
  schedule
   every 1 minutes for 1 minutes
  !
 !
 profile PCA-SLM type cfm-synthetic-loss-measurement
  probe
   send packet every 1 seconds
  !
  schedule
   every 1 minutes for 1 minutes
  !
 !
!
ethernet cfm
 domain PCA-VALIDATION-DOMAIN level 3
  service PCA-VALIDATION-SERVICE down-meps
   continuity-check interval 1s
  !
 !
!
interface TenGigE0/0/0/4.1103
 ethernet cfm
  mep domain PCA-VALIDATION-DOMAIN service PCA-VALIDATION-SERVICE mep-id 101
   loss-measurement counters aggregate
   sla operation profile PCA-DMM target mep-id 201
   sla operation profile PCA-SLM target mep-id 201
  !
 !
!
```

### Candidate config for l57

```text
ethernet sla
 profile PCA-DMM type cfm-delay-measurement
  probe
   send packet every 1 seconds
  !
  schedule
   every 1 minutes for 1 minutes
  !
 !
 profile PCA-SLM type cfm-synthetic-loss-measurement
  probe
   send packet every 1 seconds
  !
  schedule
   every 1 minutes for 1 minutes
  !
 !
!
ethernet cfm
 domain PCA-VALIDATION-DOMAIN level 3
  service PCA-VALIDATION-SERVICE down-meps
   continuity-check interval 1s
  !
 !
!
interface TenGigE0/0/0/1
 ethernet cfm
  mep domain PCA-VALIDATION-DOMAIN service PCA-VALIDATION-SERVICE mep-id 201
   loss-measurement counters aggregate
   sla operation profile PCA-DMM target mep-id 101
   sla operation profile PCA-SLM target mep-id 101
  !
 !
!
```

### Notes on this candidate

- both routers already have `performance-measurement ... delay-measurement` enabled on these interfaces, but that is not the source path used by the product XR DMM/SLM collector
- the product collector is built around:
  - `Cisco-IOS-XR-infra-sla-oper:sla/protocols/Cisco-IOS-XR-ethernet-cfm-oper:ethernet/statistics-historicals/statistics-historical`
- so Ethernet CFM/Y.1731 state is still the required validation target
- the `l57` block above is now based on validated on-box syntax
- the `l56` block is the direct mirror and should be treated as the candidate apply block until it is confirmed live

## What still needs confirmation

- The proposal now reflects the working `l57` CLI token order for:
  - `type udp jitter`
  - `datasize request`
  - `schedule operation`
  - `responder type udp ipv4 address ... port ...`
- The `l56` half is adapted to match that proven `l57` syntax, but still needs live apply confirmation.
- Which exact Ethernet segment should host the DMM/SLM peer MEPs.
- Whether management IPv4 reachability between `10.255.18.156` and `10.255.18.157` is open in the lab for the IPSLA pair.

## Recommended validation order

1. Apply only the IPSLA responder blocks first.
2. Apply the two ICMP echo operations and schedules.
3. Verify `show ipsla statistics` and `run mdt_exec` for the IPSLA path.
4. Apply the two UDP jitter operations and schedules.
5. Verify `specific_stats/op_type=udp-jitter` appears in the collector output.
6. Apply only the Ethernet CFM domain and service definitions.
7. Apply the two interface `ethernet cfm` MEP attachments on `l56 Te0/0/0/4.1103` and `l57 Te0/0/0/1`.
8. Verify CCM/peer MEP adjacency before checking telemetry.
9. Validate `run mdt_exec` on the Ethernet SLA subtree and confirm both:
   - `probe_type=cfm-delay-measurement`
   - `probe_type=cfm-synthetic-loss-measurement`

## Validation

RP/0/RP0/CPU0:l56#show ipsla statistics
Mon Mar 16 19:52:08.523 UTC
Entry number: 561 
    Modification time: 19:50:39.342 UTC Mon Mar 16 2026
    Start time       : 19:50:39.345 UTC Mon Mar 16 2026
    Number of operations attempted: 90
    Number of operations skipped  : 0
    Current seconds left in Life  : Forever
    Operational state of entry    : Active
    Operational frequency(seconds): 1
    Connection loss occurred      : FALSE
    Timeout occurred              : FALSE
    Latest RTT (milliseconds)     : 3
    Latest operation start time   : 19:52:08.348 UTC Mon Mar 16 2026
    Next operation start time     : 19:52:09.348 UTC Mon Mar 16 2026
    Latest operation return code  : OK
    RTT Values:
      RTTAvg  : 3          RTTMin: 3          RTTMax : 3         
      NumOfRTT: 1          RTTSum: 3          RTTSum2: 9

Entry number: 562 
    Modification time: 19:50:39.341 UTC Mon Mar 16 2026
    Start time       : 19:50:39.348 UTC Mon Mar 16 2026
    Number of operations attempted: 45
    Number of operations skipped  : 45
    Current seconds left in Life  : Forever
    Operational state of entry    : Active
    Operational frequency(seconds): 1
    Connection loss occurred      : FALSE
    Timeout occurred              : FALSE
    Latest RTT (milliseconds)     : 4
    Latest operation start time   : 19:52:07.349 UTC Mon Mar 16 2026
    Next operation start time     : 19:52:08.349 UTC Mon Mar 16 2026
    Latest operation return code  : OK
    RTT Values:
      RTTAvg  : 4          RTTMin: 3          RTTMax : 7         
      NumOfRTT: 10         RTTSum: 46         RTTSum2: 222
    Packet Loss Values:
      PacketLossSD       : 0          PacketLossDS : 0         
      PacketOutOfSequence: 0          PacketMIA    : 10        
      PacketLateArrival  : 0          PacketSkipped: 0
      Errors             : 0          Busies       : 0         
      InvalidTimestamp   : 0         
    Jitter Values :
      MinOfPositivesSD: 0          MaxOfPositivesSD: 0         
      NumOfPositivesSD: 0          SumOfPositivesSD: 0         
      Sum2PositivesSD : 0
      MinOfNegativesSD: 0          MaxOfNegativesSD: 0         
      NumOfNegativesSD: 0          SumOfNegativesSD: 0         
      Sum2NegativesSD : 0
      MinOfPositivesDS: 0          MaxOfPositivesDS: 0         
      NumOfPositivesDS: 0          SumOfPositivesDS: 0         
      Sum2PositivesDS : 0
      MinOfNegativesDS: 0          MaxOfNegativesDS: 0         
      NumOfNegativesDS: 0          SumOfNegativesDS: 0         
      Sum2NegativesDS : 0
      JitterAve: 0         JitterSDAve: 0      JitterDSAve: 0         
      Interarrival jitterout: 0              Interarrival jitterin: 0         
    One Way Values :
      NumOfOW: 0
      OWMinSD : 0          OWMaxSD: 0          OWSumSD: 0         
      OWSum2SD: 0          OWAveSD: 0         
      OWMinDS : 0          OWMaxDS: 0          OWSumDS: 0         
      OWSum2DS: 0          OWAveDS: 0         
RP/0/RP0/CPU0:l56#
RP/0/RP0/CPU0:l56#ssh fcbb:bb00:57::1 username lab source-interface Loopback0
Password: 
RP/0/RP0/CPU0:l56#show ipsla statistics
Mon Mar 16 19:53:45.246 UTC
Entry number: 561 
    Modification time: 19:50:39.342 UTC Mon Mar 16 2026
    Start time       : 19:50:39.345 UTC Mon Mar 16 2026
    Number of operations attempted: 187
    Number of operations skipped  : 0
    Current seconds left in Life  : Forever
    Operational state of entry    : Active
    Operational frequency(seconds): 1
    Connection loss occurred      : FALSE
    Timeout occurred              : FALSE
    Latest RTT (milliseconds)     : 5
    Latest operation start time   : 19:53:45.348 UTC Mon Mar 16 2026
    Next operation start time     : 19:53:46.348 UTC Mon Mar 16 2026
    Latest operation return code  : OK
    RTT Values:
      RTTAvg  : 5          RTTMin: 5          RTTMax : 5         
      NumOfRTT: 1          RTTSum: 5          RTTSum2: 25

Entry number: 562 
    Modification time: 19:50:39.340 UTC Mon Mar 16 2026
    Start time       : 19:50:39.347 UTC Mon Mar 16 2026
    Number of operations attempted: 94
    Number of operations skipped  : 93
    Current seconds left in Life  : Forever
    Operational state of entry    : Active
    Operational frequency(seconds): 1
    Connection loss occurred      : FALSE
    Timeout occurred              : FALSE
    Latest RTT (milliseconds)     : 4
    Latest operation start time   : 19:53:45.346 UTC Mon Mar 16 2026
    Next operation start time     : 19:53:46.346 UTC Mon Mar 16 2026
    Latest operation return code  : OK
    RTT Values:
      RTTAvg  : 4          RTTMin: 3          RTTMax : 5         
      NumOfRTT: 10         RTTSum: 40         RTTSum2: 166
    Packet Loss Values:
      PacketLossSD       : 0          PacketLossDS : 0         
      PacketOutOfSequence: 0          PacketMIA    : 10        
      PacketLateArrival  : 0          PacketSkipped: 0
      Errors             : 0          Busies       : 0         
      InvalidTimestamp   : 0         
    Jitter Values :
      MinOfPositivesSD: 0          MaxOfPositivesSD: 0         
      NumOfPositivesSD: 0          SumOfPositivesSD: 0         
      Sum2PositivesSD : 0
      MinOfNegativesSD: 0          MaxOfNegativesSD: 0         
      NumOfNegativesSD: 0          SumOfNegativesSD: 0         
      Sum2NegativesSD : 0
      MinOfPositivesDS: 0          MaxOfPositivesDS: 0         
      NumOfPositivesDS: 0          SumOfPositivesDS: 0         
      Sum2PositivesDS : 0
      MinOfNegativesDS: 0          MaxOfNegativesDS: 0         
      NumOfNegativesDS: 0          SumOfNegativesDS: 0         
      Sum2NegativesDS : 0
      JitterAve: 0         JitterSDAve: 0      JitterDSAve: 0         
      Interarrival jitterout: 0              Interarrival jitterin: 0         
    One Way Values :
      NumOfOW: 0
      OWMinSD : 0          OWMaxSD: 0          OWSumSD: 0         
      OWSum2SD: 0          OWAveSD: 0         
      OWMinDS : 0          OWMaxDS: 0          OWSumDS: 0         
      OWSum2DS: 0          OWAveDS: 0         
RP/0/RP0/CPU0:l56#

## Y1731 validation

RP/0/RP0/CPU0:l56#show ethernet cfm local maintenance-points 
Mon Mar 16 21:19:15.489 UTC
Domain/Level         Service             Interface         Type   ID   MAC
-------------------- ------------------- ----------------- ------ ---- --------
PCA-VALIDATION-DOMAIN/3 PCA-VALIDATION-SERVICE Te0/0/0/4.1103    Dn MEP  101 ea:b9:14
RP/0/RP0/CPU0:l56#show ethernet cfm peer meps ?              
  cross-check  Display only peer MEPs with cross-check errors
  detail       Display detailed information
  domain       Display for a given domain
  errors       Display only peer MEPs which have errors
  interface    Display for a given interface
  |            Output Modifiers
  <cr>         
RP/0/RP0/CPU0:l56#show ethernet cfm peer meps detail 
Mon Mar 16 21:19:47.750 UTC
Domain PCA-VALIDATION-DOMAIN (level 3), Service PCA-VALIDATION-SERVICE
Down MEP on TenGigE0/0/0/4.1103 MEP-ID 101
================================================================================
Peer MEP-ID 201, MAC 9088.554c.1c08
   CFM state: Ok, for 00:03:53
   Port state: Up
   CCMs received: 234
     Out-of-sequence:             0
     Remote Defect received:      0
     Wrong level:                 0
     Cross-connect (wrong MAID):  0
     Wrong interval:              0
     Loop (our MAC received):     0
     Config (our ID received):    0
   Last CCM received 00:00:00 ago:
     Level: 3, Version: 0, Interval: 1s
     Sequence number: 535, MEP-ID: 201
     MAID: String: PCA-VALIDATION-DOMAIN, String: PCA-VALIDATION-SERVICE
     Chassis ID: Local: l57; Management address: 'Not specified'
     Port status: Up, Interface status: Up

RP/0/RP0/CPU0:l57#show ethernet cfm local maintenance-points 
Mon Mar 16 22:14:38.282 UTC
Domain/Level         Service             Interface         Type   ID   MAC
-------------------- ------------------- ----------------- ------ ---- --------
PCA-VALIDATION-DOMAIN/3 PCA-VALIDATION-SERVICE Te0/0/0/1         Dn MEP  201 4c:1c:08
RP/0/RP0/CPU0:l57#show ethernet cfm peer meps detail 
Mon Mar 16 22:14:44.737 UTC
Domain PCA-VALIDATION-DOMAIN (level 3), Service PCA-VALIDATION-SERVICE
Down MEP on TenGigE0/0/0/1 MEP-ID 201
================================================================================
Peer MEP-ID 101, MAC ec19.2eea.b914
   CFM state: Ok, for 00:06:11
   Port state: Up
   CCMs received: 374
     Out-of-sequence:             0
     Remote Defect received:      0
     Wrong level:                 0
     Cross-connect (wrong MAID):  0
     Wrong interval:              0
     Loop (our MAC received):     0
     Config (our ID received):    0
   Last CCM received 00:00:00 ago:
     Level: 3, Version: 0, Interval: 1s
     Sequence number: 374, MEP-ID: 101
     MAID: String: PCA-VALIDATION-DOMAIN, String: PCA-VALIDATION-SERVICE
     Chassis ID: Local: l56; Management address: 'Not specified'
     Port status: Up, Interface status: Up

RP/0/RP0/CPU0:l57#show ethernet sla statistics
Mon Mar 16 22:14:59.775 UTC
Source: Interface TenGigE0/0/0/1, Domain PCA-VALIDATION-DOMAIN
Destination: Target MEP-ID 101
================================================================================
Profile 'PCA-DMM', packet type 'cfm-delay-measurement'
Scheduled to run every 1min first at 00:00:49 UTC for 1min

! Operation has no statistics - check profile configuration

Source: Interface TenGigE0/0/0/1, Domain PCA-VALIDATION-DOMAIN
Destination: Target MEP-ID 101
================================================================================
Profile 'PCA-SLM', packet type 'cfm-synthetic-loss-measurement'
Scheduled to run every 1min first at 00:00:49 UTC for 1min
Frame Loss Ratio calculated every 1min

! Operation has no statistics - check profile configuration




  ethernet sla 
   profile PCA-DMM type cfm-delay-measurement
  statistics measure one-way-delay-ds 
  statistics measure one-way-delay-sd
  statistics measure one-way-jitter-ds
  statistics measure round-trip-delay
  statistics measure round-trip-jitter
   profile PCA-SLM type cfm-synthetic-loss-measurement
  statistics measure one-way-loss-ds
  statistics measure one-way-loss-sd
  ethernet sla
   profile PCA-DMM type cfm-delay-measurement
    probe
   send burst every 1 seconds packet count 10
   send burst every 1 seconds packet count 10 interval 100 milliseconds 
  commit


RP/0/RP0/CPU0:l57#show ethernet cfm local maintenance-points 
Mon Mar 16 22:27:48.111 UTC
Domain/Level         Service             Interface         Type   ID   MAC
-------------------- ------------------- ----------------- ------ ---- --------
PCA-VALIDATION-DOMAIN/3 PCA-VALIDATION-SERVICE Te0/0/0/1         Dn MEP  201 4c:1c:08
RP/0/RP0/CPU0:l57#show ethernet cfm peer meps detail 
Mon Mar 16 22:27:56.956 UTC
Domain PCA-VALIDATION-DOMAIN (level 3), Service PCA-VALIDATION-SERVICE
Down MEP on TenGigE0/0/0/1 MEP-ID 201
================================================================================
Peer MEP-ID 101, MAC ec19.2eea.b914
   CFM state: Ok, for 00:19:23
   Port state: Up
   CCMs received: 1166
     Out-of-sequence:             0
     Remote Defect received:      0
     Wrong level:                 0
     Cross-connect (wrong MAID):  0
     Wrong interval:              0
     Loop (our MAC received):     0
     Config (our ID received):    0
   Last CCM received 00:00:00 ago:
     Level: 3, Version: 0, Interval: 1s
     Sequence number: 1166, MEP-ID: 101
     MAID: String: PCA-VALIDATION-DOMAIN, String: PCA-VALIDATION-SERVICE
     Chassis ID: Local: l56; Management address: 'Not specified'
     Port status: Up, Interface status: Up

RP/0/RP0/CPU0:l57#show ethernet sla statistics
Mon Mar 16 22:28:22.272 UTC
Source: Interface TenGigE0/0/0/1, Domain PCA-VALIDATION-DOMAIN
Destination: Target MEP-ID 101
================================================================================
Profile 'PCA-DMM', packet type 'cfm-delay-measurement'
Scheduled to run every 1min first at 00:00:49 UTC for 1min

Round Trip Delay
~~~~~~~~~~~~~~~~
1 probes per bucket

No stateful thresholds.

Bucket started at 22:25:22 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 270; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 1 (0.4%); Duplicates: 0 (0.0%)
    Result count: 270
    Min: 0.029ms; Max: 0.047ms; Mean: 0.031ms; StdDev: 0.002ms

    Results suspect due to a probe starting mid-way through a bucket

Bucket started at 22:25:49 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 600; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 600
    Min: 0.029ms; Max: 0.045ms; Mean: 0.031ms; StdDev: 0.002ms

Bucket started at 22:26:49 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 600; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 600
    Min: 0.028ms; Max: 0.043ms; Mean: 0.030ms; StdDev: 0.002ms

Bucket started at 22:27:49 UTC Mon 16 March 2026 lasting 1min
Bucket in progress
    Pkts sent: 300; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 300
    Min: 0.029ms; Max: 0.044ms; Mean: 0.031ms; StdDev: 0.002ms


One-way Delay (Source->Dest)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~
1 probes per bucket

No stateful thresholds.

Bucket started at 22:25:22 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 270; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 1 (0.4%); Duplicates: 0 (0.0%)
    Result count: 270
    Min: 1156950.231ms; Max: 1156950.330ms; Mean: 1156950.276ms; StdDev: -2147483.648ms

    Results suspect due to a probe starting mid-way through a bucket

Bucket started at 22:25:49 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 600; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 600
    Min: 1156950.033ms; Max: 1156950.241ms; Mean: 1156950.132ms; StdDev: -2147483.648ms

Bucket started at 22:26:49 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 600; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 600
    Min: 1156949.834ms; Max: 1156950.036ms; Mean: 1156949.934ms; StdDev: -2147483.648ms

Bucket started at 22:27:49 UTC Mon 16 March 2026 lasting 1min
Bucket in progress
    Pkts sent: 300; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 300
    Min: 1156949.736ms; Max: 1156949.842ms; Mean: 1156949.786ms; StdDev: -2147483.648ms


One-way Delay (Dest->Source)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~
1 probes per bucket

No stateful thresholds.

Bucket started at 22:25:22 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 270; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 1 (0.4%); Duplicates: 0 (0.0%)
    Result count: 270
    Min: -1156950.289ms; Max: -1156950.200ms; Mean: -1156950.244ms; StdDev: -2147483.648ms

    Results suspect due to a probe starting mid-way through a bucket

Bucket started at 22:25:49 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 600; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 600
    Min: -1156950.200ms; Max: -1156950.003ms; Mean: -1156950.101ms; StdDev: -2147483.648ms

Bucket started at 22:26:49 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 600; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 600
    Min: -1156950.003ms; Max: -1156949.805ms; Mean: -1156949.903ms; StdDev: -2147483.648ms

Bucket started at 22:27:49 UTC Mon 16 March 2026 lasting 1min
Bucket in progress
    Pkts sent: 300; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 300
    Min: -1156949.804ms; Max: -1156949.706ms; Mean: -1156949.754ms; StdDev: -2147483.648ms


Round Trip Jitter
~~~~~~~~~~~~~~~~~
1 probes per bucket

No stateful thresholds.

Bucket started at 22:25:22 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 270; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 1 (0.4%); Duplicates: 0 (0.0%)
    Result count: 269
    Min: 0.000ms; Max: 0.015ms; Mean: 0.002ms; StdDev: 0.003ms

    Results suspect due to a probe starting mid-way through a bucket

Bucket started at 22:25:49 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 600; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 600
    Min: 0.000ms; Max: 0.015ms; Mean: 0.002ms; StdDev: 0.002ms

Bucket started at 22:26:49 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 600; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 600
    Min: 0.000ms; Max: 0.013ms; Mean: 0.001ms; StdDev: 0.002ms

Bucket started at 22:27:49 UTC Mon 16 March 2026 lasting 1min
Bucket in progress
    Pkts sent: 300; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 300
    Min: 0.000ms; Max: 0.013ms; Mean: 0.002ms; StdDev: 0.002ms


One-way Jitter (Dest->Source)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
1 probes per bucket

No stateful thresholds.

Bucket started at 22:25:22 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 270; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 1 (0.4%); Duplicates: 0 (0.0%)
    Result count: 269
    Min: 0.000ms; Max: 0.006ms; Mean: 0.000ms; StdDev: 0.000ms

    Results suspect due to a probe starting mid-way through a bucket

Bucket started at 22:25:49 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 600; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 600
    Min: 0.000ms; Max: 0.004ms; Mean: 0.000ms; StdDev: 0.000ms

Bucket started at 22:26:49 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 600; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 600
    Min: 0.000ms; Max: 0.003ms; Mean: 0.000ms; StdDev: 0.000ms

Bucket started at 22:27:49 UTC Mon 16 March 2026 lasting 1min
Bucket in progress
    Pkts sent: 300; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 300
    Min: 0.000ms; Max: 0.003ms; Mean: 0.000ms; StdDev: 0.000ms


Source: Interface TenGigE0/0/0/1, Domain PCA-VALIDATION-DOMAIN
Destination: Target MEP-ID 101
================================================================================
Profile 'PCA-SLM', packet type 'cfm-synthetic-loss-measurement'
Scheduled to run every 1min first at 00:00:49 UTC for 1min
Frame Loss Ratio calculated every 1min

One-way Frame Loss (Source->Dest)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
1 probes per bucket

No stateful thresholds.

Bucket started at 22:24:49 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 27; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 1 (3.7%); Duplicates: 0 (0.0%)
    Result count: 1
    Min: 0.000%; Max: 0.000%; Mean; 0.000%; StdDev: 0.000%; Overall: 0.000%

    Results suspect as probe restarted mid-way through the bucket
    Results suspect as FLR calculations are based on a low packet count

Bucket started at 22:25:49 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 60; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 1
    Min: 0.000%; Max: 0.000%; Mean; 0.000%; StdDev: 0.000%; Overall: 0.000%

Bucket started at 22:26:49 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 60; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 1
    Min: 0.000%; Max: 0.000%; Mean; 0.000%; StdDev: 0.000%; Overall: 0.000%

Bucket started at 22:27:49 UTC Mon 16 March 2026 lasting 1min
Bucket in progress
    Pkts sent: 32; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 0
    Min: 0.000%; Max: 0.000%; Mean; 0.000%; StdDev: 0.000%;


One-way Frame Loss (Dest->Source)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
1 probes per bucket

No stateful thresholds.

Bucket started at 22:24:49 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 27; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 1 (3.7%); Duplicates: 0 (0.0%)
    Result count: 1
    Min: 0.000%; Max: 0.000%; Mean; 0.000%; StdDev: 0.000%; Overall: 0.000%

    Results suspect as probe restarted mid-way through the bucket
    Results suspect as FLR calculations are based on a low packet count

Bucket started at 22:25:49 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 60; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 1
    Min: 0.000%; Max: 0.000%; Mean; 0.000%; StdDev: 0.000%; Overall: 0.000%

Bucket started at 22:26:49 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 60; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 1
    Min: 0.000%; Max: 0.000%; Mean; 0.000%; StdDev: 0.000%; Overall: 0.000%

Bucket started at 22:27:49 UTC Mon 16 March 2026 lasting 1min
Bucket in progress
    Pkts sent: 32; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 0
    Min: 0.000%; Max: 0.000%; Mean; 0.000%; StdDev: 0.000%;

RP/0/RP0/CPU0:l56#show ethernet sla statistics
Mon Mar 16 21:36:23.254 UTC
Source: Interface TenGigE0/0/0/4.1103, Domain PCA-VALIDATION-DOMAIN
Destination: Target MEP-ID 201
================================================================================
Profile 'PCA-DMM', packet type 'cfm-delay-measurement'
Scheduled to run every 1min first at 00:00:01 UTC for 1min

Round Trip Delay
~~~~~~~~~~~~~~~~
1 probes per bucket

No stateful thresholds.

Bucket started at 21:34:19 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 417; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 1 (0.2%); Duplicates: 0 (0.0%)
    Result count: 417
    Min: 0.028ms; Max: 0.049ms; Mean: 0.031ms; StdDev: 0.003ms

    Results suspect due to a probe starting mid-way through a bucket

Bucket started at 21:35:01 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 600; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 600
    Min: 0.029ms; Max: 0.042ms; Mean: 0.030ms; StdDev: 0.001ms

Bucket started at 21:36:01 UTC Mon 16 March 2026 lasting 1min
Bucket in progress
    Pkts sent: 200; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 200
    Min: 0.029ms; Max: 0.042ms; Mean: 0.031ms; StdDev: 0.001ms


One-way Delay (Source->Dest)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~
1 probes per bucket

No stateful thresholds.

Bucket started at 21:34:19 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 417; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 1 (0.2%); Duplicates: 0 (0.0%)
    Result count: 417
    Min: -1156949.969ms; Max: -1156949.831ms; Mean: -1156949.899ms; StdDev: 2147483.647ms

    Results suspect due to a probe starting mid-way through a bucket

Bucket started at 21:35:01 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 600; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 600
    Min: -1156949.832ms; Max: -1156949.633ms; Mean: -1156949.732ms; StdDev: 2147483.647ms

Bucket started at 21:36:01 UTC Mon 16 March 2026 lasting 1min
Bucket in progress
    Pkts sent: 200; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 200
    Min: -1156949.634ms; Max: -1156949.568ms; Mean: -1156949.600ms; StdDev: 2147483.647ms


One-way Delay (Dest->Source)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~
1 probes per bucket

No stateful thresholds.

Bucket started at 21:34:19 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 417; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 1 (0.2%); Duplicates: 0 (0.0%)
    Result count: 417
    Min: 1156949.862ms; Max: 1156950.000ms; Mean: 1156949.931ms; StdDev: 2147483.647ms

    Results suspect due to a probe starting mid-way through a bucket

Bucket started at 21:35:01 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 600; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 600
    Min: 1156949.664ms; Max: 1156949.862ms; Mean: 1156949.763ms; StdDev: 2147483.647ms

Bucket started at 21:36:01 UTC Mon 16 March 2026 lasting 1min
Bucket in progress
    Pkts sent: 200; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 200
    Min: 1156949.598ms; Max: 1156949.664ms; Mean: 1156949.631ms; StdDev: 2147483.647ms


Round Trip Jitter
~~~~~~~~~~~~~~~~~
1 probes per bucket

No stateful thresholds.

Bucket started at 21:34:19 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 417; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 1 (0.2%); Duplicates: 0 (0.0%)
    Result count: 416
    Min: 0.000ms; Max: 0.019ms; Mean: 0.002ms; StdDev: 0.004ms

    Results suspect due to a probe starting mid-way through a bucket

Bucket started at 21:35:01 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 600; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 600
    Min: 0.000ms; Max: 0.012ms; Mean: 0.001ms; StdDev: 0.001ms

Bucket started at 21:36:01 UTC Mon 16 March 2026 lasting 1min
Bucket in progress
    Pkts sent: 200; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 200
    Min: 0.000ms; Max: 0.011ms; Mean: 0.001ms; StdDev: 0.001ms


One-way Jitter (Dest->Source)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
1 probes per bucket

No stateful thresholds.

Bucket started at 21:34:19 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 417; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 1 (0.2%); Duplicates: 0 (0.0%)
    Result count: 416
    Min: 0.000ms; Max: 0.012ms; Mean: 0.001ms; StdDev: 0.003ms

    Results suspect due to a probe starting mid-way through a bucket

Bucket started at 21:35:01 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 600; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 600
    Min: 0.000ms; Max: 0.003ms; Mean: 0.000ms; StdDev: 0.000ms

Bucket started at 21:36:01 UTC Mon 16 March 2026 lasting 1min
Bucket in progress
    Pkts sent: 200; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 200
    Min: 0.000ms; Max: 0.002ms; Mean: 0.000ms; StdDev: 0.000ms


Source: Interface TenGigE0/0/0/4.1103, Domain PCA-VALIDATION-DOMAIN
Destination: Target MEP-ID 201
================================================================================
Profile 'PCA-SLM', packet type 'cfm-synthetic-loss-measurement'
Scheduled to run every 1min first at 00:00:01 UTC for 1min
Frame Loss Ratio calculated every 1min
          
One-way Frame Loss (Source->Dest)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
1 probes per bucket

No stateful thresholds.

Bucket started at 21:34:01 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 41; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 1 (2.4%); Duplicates: 0 (0.0%)
    Result count: 1
    Min: 0.000%; Max: 0.000%; Mean; 0.000%; StdDev: 0.000%; Overall: 0.000%

    Results suspect as probe restarted mid-way through the bucket
    Results suspect as FLR calculations are based on a low packet count

Bucket started at 21:35:01 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 60; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 1
    Min: 0.000%; Max: 0.000%; Mean; 0.000%; StdDev: 0.000%; Overall: 0.000%

Bucket started at 21:36:01 UTC Mon 16 March 2026 lasting 1min
Bucket in progress
    Pkts sent: 0
    Min: 0.000%; Max: 0.000%; Mean; 0.000%; StdDev: 0.000%;


One-way Frame Loss (Dest->Source)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
1 probes per bucket

No stateful thresholds.

Bucket started at 21:34:01 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 41; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 1 (2.4%); Duplicates: 0 (0.0%)
    Result count: 1
    Min: 0.000%; Max: 0.000%; Mean; 0.000%; StdDev: 0.000%; Overall: 0.000%

    Results suspect as probe restarted mid-way through the bucket
    Results suspect as FLR calculations are based on a low packet count

Bucket started at 21:35:01 UTC Mon 16 March 2026 lasting 1min
    Pkts sent: 60; Lost: 0 (0.0%); Corrupt: 0 (0.0%);
                   Misordered: 0 (0.0%); Duplicates: 0 (0.0%)
    Result count: 1
    Min: 0.000%; Max: 0.000%; Mean; 0.000%; StdDev: 0.000%; Overall: 0.000%

Bucket started at 21:36:01 UTC Mon 16 March 2026 lasting 1min
Bucket in progress
    Pkts sent: 0
    Min: 0.000%; Max: 0.000%; Mean; 0.000%; StdDev: 0.000%;

