
# Cisco Segment Routing (SR) Performance Monitoring (PM)

This directory contains the base configuration to support several different paths within the SR-PM YANG model. 

| objectType | telegraf metric name | Yang model and path |
| :-- | :-- | :-- |
| cisco_mdt_sr_pm_last_adv | sr_pm_last_adv | "Cisco-IOS-XR-perf-meas-oper:performance-measurement/nodes/node/interfaces/interface-delay/interface-last-advertisements/interface-last-advertisement" |
| cisco_mdt_sr_pm_last_agg | sr_pm_last_agg | "Cisco-IOS-XR-perf-meas-oper:performance-measurement/nodes/node/interfaces/interface-delay/interface-last-aggregations/interface-last-aggregation" |
| cisco_mdt_sr_policy_detail | sr_pm_policy | "Cisco-IOS-XR-perf-meas-oper:performance-measurement/nodes/node/sr-policies/sr-policy-details/sr-policy-detail" | 

| objectType | Relevance |
| :-- | :-- | 
| cisco_mdt_sr_pm_last_adv | Represents the values communicated to the network routing protocol (like IS-IS or OSPF) for traffic engineering. They only update when a significant change occurs (triggered by advertisement_reason) if you are debugging why traffic is shifting (Traffic Engineering/SR-TE paths) i.e.: What does the routing table think is happening?" | 
| cisco_mdt_sr_pm_last_agg |  Represents the raw performance data collected from probes over a specific interval. last-agg tells you about a single physical hop. This tells you what is actually happening on the link right now if you only care about monitoring link latency and loss for graphing/ dashboards i.e.: What is the network feeling? | 
| cisco_mdt_sr_policy_detail | This data appears only when you have defined SR Policies (traffic engineering tunnels) and have explicitly enabled performance measurement (liveness or delay) for those policies. Unlike interface probes (which run point-to-point between neighbors), these probes measure the end-to-end performance of a specific engineered path or color across the network. sr-policy-detail tells you the cumulative experience of a packet traveling the full length of the policy | 
