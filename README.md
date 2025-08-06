# Ingestion Configuration
This repository is intended to act as the source of truth for configuration artifacts associated with ingesting  telemetry into PCA.

We recommend that you start any new integration work by creating a new branch from the main branch's latest commit. your new branch should follow the same folder structure as the pre-existing ingestion pipelines. 

For any Cisco ingestion work, please work under the `cisco-community` folder. Our R&D team will be the ones responsible for vetting and moving `cisco-community` configurations into `cisco-pca-product` as submissions get approved and moved into the product. Pull requests made under `cisco-pca-product` will be refused.

Whenever adding a new configuration, please follow the existing tiered logic:
- First folder tier should be reserved for vendor names (or cross-vendor stardards). 
- Second folder tier is reserved for second-order grouping. We are not prescriptive on this categorization but please keep it logical and don't mix concepts if you can avoid it. Ie: don't mix single model, with model series, with Operating System families, with protocol names. Because this creates confusion on how to navigate the folder structure.
- Third folder tier is made of those folders, none are mandatory, but when you are creating the folder structure please re-use those names whenever appropriate to optimize ease of use
  - `pca-ingestion-dictionaries-configuration`
  - `sensor-collector-configuration`
  - `telemetry-collector-configuration`
  - `snmp-exporter-configuration` 
  - `README.md`
  - `supporting-files`


`pca-ingestion-dictionaries-configuration`
---
- Should contain all your customized ingestion dictionaries.
- Please follow the guidelines when creating dictionary
  - If a concept was previously created by an existing type in PCA, then efforts should be made to reuse the same analyticsName and convert (if needed) the metric value to the same base unit as the existing. i.e. if we already have a cpuUtilization expressed as % with values ranging from 0-100. And you have a cpuUsage with values 0-1. Rename you new metric to cpuUtilization and multiply its value by 100 to conform with the pre-established standard.
  - When creating *new* metrics, `analyticsName` should always be in **camelCase**
  - Directions should be used whenever appropriate
  - Basic sqlExpr can be used to cast, convert, or pick, between indicators
  - If needed, a second stage of data processing can be used to create custom metrics. Custom metrics have their own requirements: 
    - `analyticsName` should always be in **camelCase**
    - They can only be computed using the `analyticsName` of stage 1 metrics.
    - Stage 1 metrics used within a Stage 2 calculation need to be enabled for the calculation to work 

`sensor-collector-configuration`
---
- Optional
- Mainly needed for computing deltas and rates
  - Delta is an operation by which Sensor Collector turns a counter into a gauge. For example, instead of an ever increasing bytesIn counter, Delta will tell you how many bytesIn were counted since the last data point.
  - Rate is another operation by which Sensor Collector turns a counter into a gauge, only it also factors in the time. For example, instead of an ever increasing bytesIn counter, Rate will convert your bytesIn value into a bytes per second.
  - **IMPORTANT**: Deltas and Rates apply on the metrics in place and overrides its existing value with the result of the operation. If you want to preserve the original metric, you will have to duplicate it *before* applying the delta or rate. 

`telemetry-collector-configuration`
---
- Optional in theory, but we expect the majority of the data feeds will have one because most integrations use the telemetry collector. 
- To contain your Telegraf configuration file
- To be valid for PCA, your Telegraf configuration file should produce at a minimum:
  - A metric name that complies with OpenMetrics standards
  - A metric value
  - A corresponding timestamp
  - A set of labels that includes:
    - `objectId` : a system wide unique identifier
    - `objectName` : a system wide unique name
    - `objectType` : a category under which all related metrics will be regrouped
    - `direction` : a direction for the measurement ["-1","0","1","2"].
        - -1 : NON - Direction None - For when the concept of direction doesnt apply (i.e. CPU Utilization)
        - 0 : SD - Source to Destination 
        - 1 : DS - Destination to Source (on the return to the sender)
        - 2 : RT - Round Trip

`snmp-exporter-configuration` 
---
- Optional. For the high-data-rate SNMP configurations where the SNMP Exporter architecture was chosen instead of the telemetry collector.
- Used to store the snmp-exporter configuration.

`README.md`
---
- Try to add a simple README that describes what is provided. For which client it was built, what is supported, what are the known limitations and the next steps (if any) 

`supporting-files`
---
- If you have supporting files, like docker-compose configs, or Dockerfile or other relevant configurations. Feel free to add and use this folder.

