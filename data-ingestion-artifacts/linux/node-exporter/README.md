# Node-Exporter example


This set of files is meant to be used as an introduction into 3rd party (external) data sources ingestion for PCA.

Using this example you can see the various steps required to ingest a new data source in PCA with very little requirements. You can get started on your laptop if you have `docker` available.

This example includes all the completed artifacts, from docker-compose to telegraf.conf, and all the way to ingestion dictionaries. The procedure still explains the steps as if you were building it from zero, but you can look at the files to better understand the process, of simply skip ahead.

I selected and mapped a few of the object types for the example and you could easily follow the same pattern to complete and cover missing metrics if you needed more for your use case.

This example also serves as a good starting point to see what needs to be done for ingesting other Prometheus/Openmetrics data sources.

Here is a link to the complete documented procedure use as a reference to this content https://docs.accedian.io/

## Step 1


Since the procedure to ingest a new data source starts with a vanilla telegraf instance to discover and model and map your data into the proper objects, I supplied a sample `docker-compose.yaml` file that can be used for a complete standalone test (no PCA needed). Once you have perfected your `telegraf.conf` file in this standalone environment you are ready for step 2.

### Initial config

Baby steps. Start with a bare bones telegraf.conf config and build it up step by step while confirming your output in the `docker logs` along the way.

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

There are four key data fields are usually required to be added:
1.	objectType
2.	sessionName
3.	sessionId
4.	direction 

### Grouping metrics into object types

Start formulating a grouping strategy to regroup metrics into objects when choosing you object types, keep in mind:
- object types are groups of related metrics, try to regroup your metrics with similar level of parity/sameness 
- up to 40 metrics can be enabled per object type in PCA at any time, try to use this a a guideline. Object types with fewer than 5 metrics might be to narrow scoped, and those with more than 50 metrics should probably be broken into smaller objects
- each metric can have up to 4 directions (none, source-destination, destination-source, round-trip)
- within a given instance of an object, a metric cannot be repeated.

### Naming considerations

- Choose the names of your new **objectTypes** that makes them intuitive, and shows both their uniqueness (and relatedness). Try to make sure to make them descriptive and not overly broad, or overly specific.  
- Come up with a suitable **sessionName** pattern for your objects that makes them globally unique and clearly understood by the end users. As a rule of thumb, sessionNames are generally built using the supplied tags available in your data stream and some constant value.
  - Identify the base object: hostname, servicename, customer, basestation, etc.
  - Identify the specific sub-component (if applicable): interface name, cpu number, network slice, tunnel id, vlan, qos marking, etc
  - If there is the slightest risk that the combination of tags you choose might be repeated for a different object type, add a trailer that repeats the objectType withing the objectName. That is because the objectName **MUST** be globally unique within the system.
  - ex: `'host-{{ .Tag "host" }}_cpu-{{ .Tag "cpu" }}_node-exporter-cpu'` will give us: `host-561a0f455637_cpu-0_node-exporter-cpu`. 
- Based on the information available in your data stream, choose a suitable **sessionId** unique identifier for your objects. This can be the same as the object name, but it doesnt have to be.
- And finally choose a **direction**, if you don’t know, pick `-1` (None)

### Metric processing and customization

Looking at your metric names, see if you can already identify patterns that will help you:
1. See if you can find a pattern to sort metrics into those objectTypes you have identified 
2. Process the names of the metrics (if needed) to make them unique within a given object
3. Do a 1st pass of data processing to parse the metric names and remove repetitive and filler strings (if needed)
4. Identify early processing that should be done at this stage (i.e unit conversions) or to compute custom KPIs (if needed)
5. If applicable, you can choose to also drop/remove any unneeded metric from the data stream at this point to avoid unneeded processing down the line. This can be done with starlark and/or `tag_limit` paired with `fieldinclude` to only select those metrics/fields you want to preserve.

## Step 2

Prerequisite: a working Sensor Collector configured for `"Type": "Gateway"` and `"Metric Configuration": "telemetry-collector"` (this is outside the scope of this document).

1. In your `telegraf.conf`, comment your `[agent]` and `[[output.*]]` configs. Then convert your `telegraf.conf` file into a base64 blob
2. Create a new Telemetry Collector instance in PCA and select IOS XR
3. Install this new Telemetry Collector instance on a compute node that will have network access to the node-exporter you wish to scrape (if you need help with these steps refer to the Telemtry collector documentation)
4. Using the Telemetry Collector ID `GET` the `/api/orchestrate/v3/agents/configuration/{{agentId}}` API
5. Edit the JSON payload you got as a response to the API call in #4 to replace the `data -> attributes -> telemetry -> dataCollection` with the one you generated in step #1.  
6. (optional but nice) Also update the `data -> attributes -> telemetry -> templateName` to something representative of your updated configuration such as "Node-Exporter"
7. Using the Telemetry Collector ID `PUT` back the `/api/orchestrate/v3/agents/configuration/{{agentId}}` API with your updated configuration
8. In your Telemetry Collector logs you should see entries indicating a new configuration was applied: 
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

## Step 3

1. Using your API client, GET the dictionaries
1. Using the dictionaries privided in this git repositories, update as needed
1. If the dictionaries are not supplied as part of this example, use the provided python script to parse and convert the resulting dictionary that was automatically created to one which conforms with the best practices