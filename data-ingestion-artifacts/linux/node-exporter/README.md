Node-Exporter example
===

This set of files is meant to be used as an introduction into 3rd party (external) data sources ingestion for PCA.

Using this example you can see the various steps required to ingest a new data source in PCA with very little requirements. You can get started on your laptop if you have `docker` available.

I selected and mapped a few of the object types for the example and you could easily follow the same pattern to complete and cover missing metrics if you needed more for your use case.

This example also serves as a good starting point to see what needs to be done for ingesting other Prometheus/Openmetrics data sources.

Here is a link to the complete documented procedure use as a reference to this content https://docs.accedian.io/

Step 1
---

Since the procedure to ingest a new data source starts with a vanilla telegraf instance to discover and model and map your data into the proper objects, I supplied a sample `docker-compose.yaml` file that can be used for a complete standalone test (no PCA needed). Once you have perfected your `telegraf.conf` file in this standalone environment you are ready for step 2.

Step 2
---

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

Step 3
---

1. Using your API client, GET the dictionaries
1. Using the dictionaries privided in this git repositories, update as needed
1. If the dictionaries are not supplied as part of this example, use the provided python script to parse and convert the resulting dictionary that was automatically created to one which conforms with the best practices