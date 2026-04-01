# Node-Exporter example

This set of files is meant to be used as an introduction into new data sources ingestion for PCA.

Using this example you can see the steps required to ingest a new data source in PCA with very little prerequisites. For the hands-on exercise, you can get started on your laptop if you have `docker` available.

This example includes all the completed configuration files, from docker-compose to telegraf.conf, and all the way to PCA ingestion dictionaries. The procedure still explains the steps as if you were building it from zero, but you can look at the files to better understand the process, or simply use them as-is.

I selected and mapped a few of the object types for the example. You could follow the same pattern to complete and cover missing metrics if you need more for your use case.

This example also serves as a good starting point to see what needs to be done for ingesting other Prometheus/Openmetrics data sources.

For complete documentation on deploying of Sensor Collector and Telemetry Collector refer to this article: https://docs.accedian.io/on-prem-solution/docs/pca-for-mobility-pm . You can use this documentation to assist if any of the steps below are lacking the detail you need.

## Step 1 - Build your telegraf configuration


Since the procedure to ingest a new data source starts with a generic telegraf instance to discover and model and map your data into the proper objects, I supplied a sample `docker-compose-standalone.yaml` file that can be used for a complete standalone test (no PCA components needed yet). Once you have perfected your `telegraf.conf` file in this standalone environment you are ready for step 2.

### Initial config

Baby steps. Start with a simple telegraf.conf config (as shown below) and build it up step by step while confirming your output in the `docker logs` along the way.

```
[agent]
  interval = "60s"
  flush_interval = "20s"
  debug = true
  omit_hostname = true

[[inputs.prometheus]]
  urls = ["http://node-exporter:9100/metrics"]
  interval = "60s"

[[outputs.file]]  
   data_format = "json"
   # Files to write to, "stdout" is a specially handled file.
   # files = ["stdout","/tmp/metrics.out"]
   files = ["stdout"]

```


### Required tags for PCA


When building your telegraf.conf there are several considerations you need to keep in mind.

There are four mandatory data fields that **MUST** to be added:
1.	objectType
2.	sessionName
3.	sessionId
4.	direction 

### Grouping metrics into object types

Start formulating a grouping strategy to regroup metrics into objects. 

When choosing your object types, keep in mind:
- object types are groups of related metrics. For example: all the metrics belonging to an ethernet interface would create a interface object.
- try to regroup your metrics with similar scope/parity/sameness. For example: if you select a very high level metric, say system overall health, don't mix it with a very low level metric, like number of memory page faults, or number of bytes in swap.
- up to 40 metrics can be enabled per object type in PCA at any time, try to use this a a guideline. Object types with fewer than 5 metrics might be to narrowly scoped, and those with more than 50 metrics should probably be broken into smaller objects. Also keep in mind, objects with more than 255 metrics **will not make it through to PCA**. 
- each metric can have up to 4 directions (none, source-destination, destination-source, round-trip)
- within an object type, a metric cannot be repeated.

### Naming considerations

- Choose the names of your new **objectTypes** that makes them intuitive, and shows both their uniqueness (and relatedness). Try to make sure to make them descriptive and not overly broad, or overly specific. For example, here are some existing objectType values in PCA:
    - `cisco-telemetry-xe-cpu`
    - `cisco-telemetry-xe-interface`
    - `cisco-telemetry-xr-ipsla-icmp-echo`
- The **sessionName** is what you will find in your PCA inventory screen. Come up with a suitable **sessionName** pattern for your objects that makes them globally unique and clearly understood by the end users. As a rule of thumb, sessionNames are generally built using the supplied tags available in your data stream concatenated and some constant value.
  - Identify the base object. For example: hostname, servicename, customer, basestation, etc.
  - Identify the specific sub-component (if applicable). For example: interface name, cpu number, network slice, tunnel id, vlan, qos marking, etc
  - The objectName **MUST** be globally unique within the system. If there is the slightest risk that the combination of tags you choose might be repeated for a different object type, add a trailer that appends the objectType name at the end of the objectName to make it unique. 
  - For example, for a node-exported-cpu object: `'host-{{ .Tag "host" }}_cpu-{{ .Tag "cpu" }}_node-exporter-cpu'` will give us: host-561a0f455637_cpu-0_**node-exporter-cpu**. 
- Every object needs a unique **sessionId** identifier in PCA. Based on the information available in your data stream, choose a suitable **sessionId** unique identifier for your objects. This can be the same as the object name, but it doesn't have to be.
- And finally set the **direction** to `-1` (None).

### Metric processing and customization

Looking at your metric name in your `docker logs`:
1. Find a pattern to assign metrics into those objectTypes you have identified 
2. Process the names of the metrics (if needed) to make them unique within a given object type.
3. Do a 1st pass of data processing to parse the metric names and remove repetitive and filler strings (if needed)
4. Identify early processing that should be done at this stage (i.e unit conversions) or to compute custom KPIs (if needed)
5. If applicable, you can choose to also drop/remove any unneeded metrics from the data stream at this point to avoid unneeded processing down the line. This can be done with starlark and/or `tag_limit` paired with `fieldinclude` to only select those metrics/fields you want to preserve.

Once you have iterated over your configuration and you are happy with the results in your `docker logs` you are ready to move on to Step 2.

## Step 2 - Convert your telegraf configuration to Telemetry Collector

Prerequisite: a working Sensor Collector configured for `"Type": "Gateway"` and `"Metric Configuration": "telemetry-collector"` (this is outside the scope of this document). Refer to the article mentioned at the top of this procedure for further reference.

1. In your `telegraf.conf` from step #1, comment out your `[agent]` and `[[output.*]]` configs since we will be supplying those from the PCA orchestration. Then convert your `telegraf.conf` file into a base64 blob.
```
base64 < telegraf.conf > telegraf.base64
```
2. Using the PCA UI, create a new Telemetry Collector instance and select the `IOS XR` Transform configuration. For greater detail, refer to the procedure documented here: 
- https://docs.accedian.io/on-prem-solution/docs/pca-for-mobility-pm#personal-access-token
- https://docs.accedian.io/on-prem-solution/docs/pca-for-mobility-pm#2-configure-customize-and-deploy-telemetry-collector

3. Install this new Telemetry Collector instance on a compute node that will have network access to the node-exporter you wish to scrape (if you need help with these steps refer to the documentation mentioned at the top of this procedure for further reference). A sample docker-compose is provided (`docker-compose-telemetry-collector.yaml`).
4. Using the Telemetry Collector ID (you can see it in the PCA UI) `GET` the `/api/orchestrate/v3/agents/configuration/{{agentId}}` API. 
5. Edit the JSON payload you got as a response to the API call in #4 to replace the `data -> attributes -> telemetry -> dataCollection` with the one you generated in step #1.  
6. (optional but recommended) Also update the `data -> attributes -> telemetry -> templateName` to something representative of your updated configuration such as "Node-Exporter"
7. Using the Telemetry Collector ID, `PUT` back the `/api/orchestrate/v3/agents/configuration/{{agentId}}` API with your updated configuration
8. In your Telemetry Collector `docker logs` you should see entries indicating a new configuration was applied. You can look for messages such as those: 
```
Reloading Telegraf config
applying new telegraf configuration 'DataCollection: Node-Exporter version 1'
```
Soon after you should see a number of log entries like these:
```
creating reporting stream for session , host-f67c0f157bb1_cpu-4_node-exporter-cpu (object type node-exporter-cpu), stream index 2
Setting data gateway ring buffer size to 500000
creating reporting stream for session , host-f67c0f157bb1_cpu-0_node-exporter-cpu (object type node-exporter-cpu), stream index 3
creating reporting stream for session , host-f67c0f157bb1_cpu-3_node-exporter-cpu (object type node-exporter-cpu), stream index 4
creating reporting stream for session , host-f67c0f157bb1_cpu-6_node-exporter-cpu (object type node-exporter-cpu), stream index 5
creating reporting stream for session , host-f67c0f157bb1_node-exporter-netstat (object type node-exporter-netstat), stream index 6
creating reporting stream for session , host-f67c0f157bb1_cpu-8_node-exporter-cpu (object type node-exporter-cpu), stream index 7
creating reporting stream for session , host-f67c0f157bb1_device-/dev/sdb1_node-exporter-filesystem (object type node-exporter-filesystem), stream index 8
creating reporting stream for session , host-f67c0f157bb1_cpu-2_node-exporter-cpu (object type node-exporter-cpu), stream index 9
creating reporting stream for session , host-f67c0f157bb1_cpu-7_node-exporter-cpu (object type node-exporter-cpu), stream index 10
creating reporting stream for session , host-f67c0f157bb1_cpu-11_node-exporter-cpu (object type node-exporter-cpu), stream index 11
creating reporting stream for session , host-f67c0f157bb1_device-eth0_node-exporter-network (object type node-exporter-network), stream index 12
creating reporting stream for session , host-f67c0f157bb1_device-lo_node-exporter-network (object type node-exporter-network), stream index 13
creating reporting stream for session , host-f67c0f157bb1_cpu-9_node-exporter-cpu (object type node-exporter-cpu), stream index 14
creating reporting stream for session , host-f67c0f157bb1_node-exporter-memory (object type node-exporter-memory), stream index 15
creating reporting stream for session , host-f67c0f157bb1_cpu-15_node-exporter-cpu (object type node-exporter-cpu), stream index 16
creating reporting stream for session , host-f67c0f157bb1_cpu-5_node-exporter-cpu (object type node-exporter-cpu), stream index 17
creating reporting stream for session , host-f67c0f157bb1_device-sda_node-exporter-disk (object type node-exporter-disk), stream index 18
creating reporting stream for session , host-f67c0f157bb1_cpu-10_node-exporter-cpu (object type node-exporter-cpu), stream index 19
creating reporting stream for session , host-f67c0f157bb1_cpu-13_node-exporter-cpu (object type node-exporter-cpu), stream index 20
creating reporting stream for session , host-f67c0f157bb1_cpu-14_node-exporter-cpu (object type node-exporter-cpu), stream index 21
creating reporting stream for session , host-f67c0f157bb1_cpu-1_node-exporter-cpu (object type node-exporter-cpu), stream index 22
creating reporting stream for session , host-f67c0f157bb1_cpu-12_node-exporter-cpu (object type node-exporter-cpu), stream index 23
creating reporting stream for session , host-f67c0f157bb1_device-sdb_node-exporter-disk (object type node-exporter-disk), stream index 24
Sending performance data
```

**Note** : It is also expected that you will see a number of warning messages that state: `missing objectType label - data will be discarded`. This is expected and nothing to worry about.

## Step 3 - Setup the PCA configuration

1. When the data flows into PCA, the system will automatically generate new data dictionarie for you. You **MUST** update those dictionaries **BEFORE** turning on metric ingestion in your ingestion profiles to avoid problems down the line.
1. Using your API client, GET the dictionaries (`{{baseUrl}}/v3/ingestion-dictionaries/:IngestionDictionaryId`)
1. Using the dictionaries provided in this git repository, update as needed. More details on this here: https://docs.accedian.io/on-prem-solution/docs/pca-for-mobility-pm#3-install-ingestion-dictionaries-for-mobility-pm-data
1. If the dictionaries are not supplied as part of this example, use the provided python script to parse and convert the resulting dictionary that was automatically created to one which conforms with the best practices. 

As a general rule, it is recommended to:
- make a strong effort to reuse existing `analyticsName` whenever a metric for the same concept already exists. If you need to convert units, do it using sqlExpr (writing sqlExpr is beyond the scope of this readme)
- If you have to create new metrics, try to standardize on camelCase for the `analyticsName`
- Remove repetitive prefix or suffix strings that do not help clarify the metric
- Avoid making the metric name vendor specific if you can avoid it (future reusability)
- When appropriate assign units to benefit from auto-summarization (bytes->KB->MB->GB...)

4. Once your dictionaries are in good shape, PATCH them back into the system with (`{{baseUrl}}/v3/ingestion-dictionaries/:IngestionDictionaryId`). If you are not allowed to PATCH because of authorization issues, send your updated dictionaries to CSM as JSON files using the Intercom chat feature built into PCA

## Step 4 - Enable your new metrics

In PCA, navigate to the Settings -> Ingestion to configure your ingestion profile for the new dictionaries you just updated and toggle on all the metrics you need for your use case.

## Step 5 - Future work and finishing touches

Once your data is coming in, your might notice that some metrics are behaving like counters (forever increasing), while others behave like gauges (moving up and down based on the instant value). 

If you want to convert a counter into a gauge, a Sensor Collector configuration will be required. But at this moment this is beyond the scope for this document.

Also, once you start to see the metrics come in, you might want to go back and tweak your dictionary units. This is easy and allowed, simply go back to the dictionary step and tweak and patch until you are satisfied with the result.

Finally, build dashboards. You can see more information on this process here: https://docs.accedian.io/on-prem-solution/docs/pca-for-mobility-pm#5-setup-dashboards
