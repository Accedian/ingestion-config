# Telegraf Starlark Configuration Explained

This document explains the `telegraf-starlark.conf` configuration in the context of the PCA (Performance & Connectivity Analytics) data architecture.

## Overview

This configuration is the **transformation layer** in the Telemetry Collector that converts raw Cisco Mobility Core PM telemetry into the format required by the PCA ingestion pipeline.

## Data Flow Context

```
┌─────────────────┐    ┌─────────────────────────────────┐    ┌──────────────┐
│  Cisco Mobility │    │      THIS TELEGRAF CONFIG       │    │  Roadrunner  │
│  Core (Kafka)   │───▶│  (Starlark Transformation)      │───▶│    → Fedex   │
│                 │    │                                 │    │              │
│  Raw KPI data   │    │  Adds: objectType, sessionId,   │    │  Normalized  │
│                 │    │        sessionName, direction   │    │  metrics     │
└─────────────────┘    └─────────────────────────────────┘    └──────────────┘
```

## Processing Pipeline

The configuration processes metrics through several ordered stages:

| Order | Processor | Purpose |
|-------|-----------|---------|
| 2 | Internal handler | Adds required tags to `internal_*` metrics |
| 3 | Regex | Cleans index tag - removes `["` and `"]` brackets |
| 4 | Strings | Cleans index tag - replaces commas with underscores |
| 7 | KPI Splitting | Routes KPIs to objectTypes, builds session identifiers |
| 8 | Rename | Renames `node_ip` to `source_ip` |
| 9 | P2P Filter | Filters out `p2p_protocol#` objects |
| 10 | KPI Limit Monitor | Monitors for unexpected KPIs exceeding 40 limit |
| 11 | Tag Limit | Limits tags to required set |

## Key Transformations

| Input (Raw) | Output (Normalized) | Purpose |
|-------------|---------------------|---------|
| `kpi` (metric name) | Used to lookup `objectType` suffix | Routes to correct **Ingestion Dictionary** |
| `schema` + `suffix` | `objectType` | Links to the dictionary (e.g., `cisco-mobilitycore-pm-mme-failure`) |
| `device` + `index` + `objectType` | `sessionName` | **MonitoredObjectName** (e.g., `mme-node-1_servname#MME-SVC_cisco-mobilitycore-pm-mme-sr`) |
| `node_id` + `index` + `objectType` | `sessionId` | Unique **MonitoredObjectId** |

## The EXACT_RULES Dictionary

This maps each KPI to its **objectType suffix**, which determines which **Ingestion Dictionary** handles it:

```python
EXACT_RULES = {
    # These KPIs go to cisco-mobilitycore-pm-mme-failure
    "EMM_Attach_Reject_Decode_Failure": "-failure",
    "EMM_Attach_Reject_Network_Failure": "-failure",
    
    # These KPIs go to cisco-mobilitycore-pm-mme-sr (success rate)
    "MME_Overall_Attach_Success_Rate": "-sr",
    "MME_S1_Handover_Success_Rate": "-sr",
    
    # These KPIs go to cisco-mobilitycore-pm-mme-inter
    "MME_Inter_MME_TAU_Rejects_Total": "-inter",
    
    # KPIs NOT in this list → base objectType (cisco-mobilitycore-pm-mme)
}
```

### ObjectType Suffixes

| Suffix | ObjectType Example | Description |
|--------|-------------------|-------------|
| (none) | `cisco-mobilitycore-pm-mme` | Base MME metrics |
| `-failure` | `cisco-mobilitycore-pm-mme-failure` | Failure/rejection metrics |
| `-sr` | `cisco-mobilitycore-pm-mme-sr` | Success rate metrics |
| `-inter` | `cisco-mobilitycore-pm-mme-inter` | Inter-MME handover metrics |
| `-intra` | `cisco-mobilitycore-pm-mme-intra` | Intra-MME handover metrics |
| `-dcnr` | `cisco-mobilitycore-pm-mme-dcnr` | Dual Connectivity NR metrics |
| `-pdn` | `cisco-mobilitycore-pm-mme-pdn` | PDN connectivity metrics |

## Index Cleaning (Pre-Starlark)

Before the Starlark processor runs, the `index` tag is cleaned by two processors:

### 1. Regex Processor (order 3)
Removes JSON array brackets from the index:
```toml
[[processors.regex]]
  order = 3
  [[processors.regex.tags]]
    key = "index"
    pattern = '^\["(.*?)"\]$'
    replacement = '${1}'
```

### 2. Strings Processor (order 4)
Replaces commas with underscores:
```toml
[[processors.strings]]
  order = 4
  [[processors.strings.replace]]
    tag = "index"
    old = ","
    new = "_"
```

### Index Cleaning Examples

| Raw Index | After Regex | After Strings |
|-----------|-------------|---------------|
| `["servname#MME-SVC"]` | `servname#MME-SVC` | `servname#MME-SVC` |
| `["card#1,port#2"]` | `card#1,port#2` | `card#1_port#2` |
| `["apn#internet,qci#5"]` | `apn#internet,qci#5` | `apn#internet_qci#5` |

## The `apply()` Function - Line by Line

```python
def apply(metric):
    kpi = metric.name                          # The KPI name (e.g., "MME_Overall_Attach_Success_Rate")
    schema = metric.tags.get("schema", "")     # DIMENSION: "mme", "sgw", "pgw", "p2p"
    device = metric.tags.get("device", "")     # DIMENSION: device hostname
    index = metric.tags.get("index", "")       # DIMENSION: cleaned index value
    node_id = metric.tags.get("node_id", "")   # DIMENSION: unique node identifier
    
    # Step 1: Determine objectType suffix from KPI name
    suffix = EXACT_RULES.get(kpi, "")          # "-failure", "-sr", "" (base), etc.
    
    # Step 2: Build objectType = "cisco-mobilitycore-pm-" + schema + suffix
    object_type_base = "cisco-mobilitycore-pm-" + schema
    object_type = object_type_base + suffix
    # Example: "cisco-mobilitycore-pm-mme-sr"
    
    # Step 3: Build session identifiers (includes objectType for uniqueness)
    session_name = device + "_" + index + "_" + object_type
    session_id = node_id + "_" + index + "_" + object_type
    
    # Step 4: Add required tags for PCA pipeline
    metric.tags["sessionName"] = session_name   # → MonitoredObjectName
    metric.tags["sessionId"] = session_id       # → MonitoredObjectId
    metric.tags["objectType"] = object_type     # → Links to Ingestion Dictionary
    metric.tags["direction"] = "-1"             # No direction for this data
    
    return metric
```

## Real Example

### Input from Kafka (Raw)

```json
{
  "kpi": "MME_Overall_Attach_Success_Rate",
  "value": 99.5,
  "index": "[\"servname#MME-SVC\"]",
  "timestamp": 1707744000,
  "device": "mme-node-1",
  "node_id": "node123",
  "schema": "mme"
}
```

### After Index Cleaning (order 3-4)

```json
{
  "kpi": "MME_Overall_Attach_Success_Rate",
  "value": 99.5,
  "index": "servname#MME-SVC",
  "timestamp": 1707744000,
  "device": "mme-node-1",
  "node_id": "node123",
  "schema": "mme"
}
```

### After Starlark Transformation (order 7)

```json
{
  "kpi": "MME_Overall_Attach_Success_Rate",
  "value": 99.5,
  "timestamp": 1707744000,
  
  "device": "mme-node-1",
  "node_id": "node123", 
  "index": "servname#MME-SVC",
  "schema": "mme",
  
  "objectType": "cisco-mobilitycore-pm-mme-sr",
  "sessionId": "node123_servname#MME-SVC_cisco-mobilitycore-pm-mme-sr",
  "sessionName": "mme-node-1_servname#MME-SVC_cisco-mobilitycore-pm-mme-sr",
  "direction": "-1"
}
```

## Why This Matters (PCA Context)

1. **ObjectType** → Routes data to the correct **Ingestion Dictionary** (e.g., `cisco-mobilitycore-pm-mme-sr.json`)
2. **SessionId** → Becomes the **MonitoredObjectId** - uniquely identifies this data stream in the system
3. **SessionName** → Becomes the **MonitoredObjectName** - displayed in UI dashboards
4. The dictionary defines which fields are **METRICS** (success rates, counts - values you can do math on) vs **DIMENSIONS** (device, index - labels for filtering/grouping)

## Understanding Dimensions vs Metrics

| Aspect | **Dimensions** | **Metrics** |
|--------|---------------|-------------|
| **What they are** | Labels/identifiers that describe *what* is being measured | Numeric values that represent *how much* |
| **Purpose** | Used for filtering, grouping, slicing data | Used for calculations, aggregations |
| **Examples in this config** | `device`, `node_id`, `index`, `schema` | `value` (the KPI measurement) |
| **SQL analogy** | `GROUP BY` columns | `SUM()`, `AVG()`, `COUNT()` columns |

### Simple Rule of Thumb

> **Can you do math on it?**
> - ✅ YES → It's a **METRIC** (success rate, packet count, latency)
> - ❌ NO → It's a **DIMENSION** (device name, node ID, schema type)

## Schema Types

The `schema` tag determines the base objectType:

| Schema | Base ObjectType | Description |
|--------|-----------------|-------------|
| `mme` | `cisco-mobilitycore-pm-mme` | Mobility Management Entity |
| `sgw` | `cisco-mobilitycore-pm-sgw` | Serving Gateway |
| `pgw` | `cisco-mobilitycore-pm-pgw` | PDN Gateway |
| `p2p` | `cisco-mobilitycore-pm-p2p` | Point-to-Point (no index in session ID) |

## Special Handling for p2p Schema

The p2p schema is handled by a separate Starlark processor that excludes the index from session identifiers:

```python
def apply(metric):
    schema = metric.tags.get("schema", "")
    device = metric.tags.get("device", "")
    node_id = metric.tags.get("node_id", "")
    
    object_type = "cisco-mobilitycore-pm-" + schema
    session_name = device + "_" + object_type      # No index
    session_id = node_id + "_" + object_type       # No index
    
    metric.tags["sessionName"] = session_name
    metric.tags["sessionId"] = session_id
    metric.tags["objectType"] = object_type
    metric.tags["direction"] = "-1"
    
    return metric
```

Additionally, p2p metrics with `p2p_protocol#` in the index are filtered out entirely.

## KPI Limit Monitor

The configuration includes a monitor (order 10) that tracks unique KPIs per objectType:

- **Warning** at 35 KPIs: Approaching the limit
- **Critical** at 40 KPIs: Limit exceeded

This is **monitor-only** - metrics are never dropped. Alerts are written to `/tmp/kpi_alerts.log`.

---

## Related Documentation

- Ingestion Dictionary files: `adh-gather/files/v3IngestionDictionaries/`

---

*Generated: 2026-02-13*
