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

## Key Transformations

| Input (Raw) | Output (Normalized) | Purpose |
|-------------|---------------------|---------|
| `kpi` (metric name) | Used to lookup `objectType` suffix | Routes to correct **Ingestion Dictionary** |
| `schema` + `suffix` | `objectType` | Links to the dictionary (e.g., `cisco-mobilitycore-pm-mme-failure`) |
| `device` + `clean_index(index)` | `sessionName` | Human-readable **MonitoredObjectName** (e.g., `mme-node-1 / MME-SVC`) |
| `node_id` + `clean_index(index)` | `sessionId` | Unique **MonitoredObjectId** |

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

## The `clean_index()` Helper Function

The raw `index` values from Cisco Mobility Core come in JSON array format with prefixes. This function cleans them for human-readable sessionNames:

```python
def clean_index(raw_index):
    """
    Clean the index value from JSON array notation and common prefixes.
    Examples:
      '["servname#MME-SVC"]' -> 'MME-SVC'
      '["card#1,port#2"]' -> '1/2'
      '["NOINDEX"]' -> ''
      '["apn#internet,qci#5"]' -> 'internet/5'
    """
    if not raw_index:
        return ""
    
    # Strip JSON array brackets: ["..."] -> ...
    clean = raw_index
    if clean.startswith('["') and clean.endswith('"]'):
        clean = clean[2:-2]
    elif clean.startswith('[') and clean.endswith(']'):
        clean = clean[1:-1]
    
    # Handle NOINDEX
    if clean == "NOINDEX" or clean == "":
        return ""
    
    # Split by comma for multi-key indexes
    parts = clean.split(",")
    values = []
    
    for part in parts:
        # Remove prefix before # (e.g., "servname#MME-SVC" -> "MME-SVC")
        if "#" in part:
            value = part.split("#", 1)[1]
        else:
            value = part
        
        value = value.strip().strip('"').strip("'")
        if value:
            values.append(value)
    
    return "/".join(values)
```

### Index Cleaning Examples

| Raw Index | Cleaned Value | Notes |
|-----------|--------------|-------|
| `["servname#MME-SVC"]` | `MME-SVC` | Removes JSON brackets and `servname#` prefix |
| `["card#1,port#2"]` | `1/2` | Handles multi-key, joins with `/` |
| `["NOINDEX"]` | *(empty)* | Special case - no index needed |
| `["apn#internet,qci#5"]` | `internet/5` | Multiple values extracted |

## The `apply()` Function - Line by Line

```python
def apply(metric):
    kpi = metric.name                          # The KPI name (e.g., "MME_Overall_Attach_Success_Rate")
    schema = metric.tags.get("schema", "")     # DIMENSION: "mme", "sgw", "pgw", "p2p"
    device = metric.tags.get("device", "")     # DIMENSION: device hostname
    index = metric.tags.get("index", "")       # DIMENSION: interface/cell index (raw JSON format)
    node_id = metric.tags.get("node_id", "")   # DIMENSION: unique node identifier
    
    # Step 1: Determine objectType suffix from KPI name
    suffix = EXACT_RULES.get(kpi, "")          # "-failure", "-sr", "" (base), etc.
    
    # Step 2: Build objectType = "cisco-mobilitycore-pm-" + schema + suffix
    object_type = "cisco-mobilitycore-pm-" + schema + suffix
    # Example: "cisco-mobilitycore-pm-mme-failure"
    
    # Step 3: Clean the index value
    clean_idx = clean_index(index)             # "[servname#MME-SVC]" → "MME-SVC"
    
    # Step 4: Build MonitoredObject identifiers
    if schema == "p2p" or clean_idx == "":
        session_name = device                  # Just device name
        session_id = node_id                   # Just node ID
    else:
        session_name = device + " / " + clean_idx   # Human-readable: "mme-node-1 / MME-SVC"
        session_id = node_id + "_" + clean_idx      # Unique ID: "node123_MME-SVC"
    
    # Step 5: Add required tags for PCA pipeline
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
  "index": "[\"servname#MME-SVC\"]",
  "timestamp": 1707744000,
  "device": "mme-node-1",
  "node_id": "node123",
  "schema": "mme"
}
```

### After Starlark Transformation

```json
{
  "kpi": "MME_Overall_Attach_Success_Rate",    // METRIC: the KPI value
  "value": 99.5,                               // METRIC: numeric value
  "timestamp": 1707744000,
  
  // DIMENSIONS (for filtering/grouping):
  "device": "mme-node-1",
  "node_id": "node123", 
  "index": "[\"servname#MME-SVC\"]",           // Raw index (kept for reference)
  "schema": "mme",
  
  // ADDED BY STARLARK (for PCA pipeline):
  "objectType": "cisco-mobilitycore-pm-mme-sr",   // → Ingestion Dictionary lookup
  "sessionId": "node123_MME-SVC",                 // → MonitoredObjectId (cleaned)
  "sessionName": "mme-node-1 / MME-SVC",          // → MonitoredObjectName (human-readable!)
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

## Special Handling for p2p Schema and Empty Index

When `schema == "p2p"` OR when the cleaned index is empty (NOINDEX), only the device/node_id is used:

```python
clean_idx = clean_index(index)  # May return "" for NOINDEX or missing

if schema == "p2p" or clean_idx == "":
    session_name = device           # Just "mme-node-1"
    session_id = node_id            # Just "node123"
else:
    session_name = device + " / " + clean_idx   # "mme-node-1 / MME-SVC"
    session_id = node_id + "_" + clean_idx      # "node123_MME-SVC"
```

---

## Related Documentation

- Ingestion Dictionary files: `adh-gather/files/v3IngestionDictionaries/`

---

*Generated: 2026-02-12*
