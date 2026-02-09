# KPI Splitting Solutions for Cisco Mobility Core PM

## Problem Statement

Rakuten's Matrix schema has a **40 KPI per objectType limit**. The MME schema contains 127 unique KPIs, exceeding this limit and requiring automatic splitting into multiple objectTypes.

### Current Schema Analysis

| Schema | Unique KPIs | Status | Action |
|--------|-------------|--------|--------|
| `mme` | 127 | **Exceeds limit** | Split into 5 objectTypes |
| `egtpc` | 28 | ✅ Under limit | No action needed |
| `schema` | 25 | ✅ Under limit | No action needed |
| `hss` | 11 | ✅ Under limit | No action needed |
| `port` | 5 | ✅ Under limit | No action needed |

### MME Schema Split

The 127 MME KPIs are split into **5 objectTypes**, each under 40 KPIs:

| ObjectType | KPI Patterns | Count |
|------------|--------------|-------|
| `cisco-mobilitycore-pm-mme-inter` | `MME_Inter_*` | 23 |
| `cisco-mobilitycore-pm-mme-intra` | `MME_Intra_*` | 20 |
| `cisco-mobilitycore-pm-mme-emm` | `EMM_*`, `EPS_*`, `Intra_MME_*`, `UE_*`, `S1_Paging*` | 22 |
| `cisco-mobilitycore-pm-mme-pdn` | `MME_PDN_*`, `MME_DCNR_*` | 28 |
| `cisco-mobilitycore-pm-mme` | All remaining KPIs | 34 |
| **Total** | | **127** |

---

## Available Solutions

Two production-ready configurations are provided:

| File | Approach | Best For |
|------|----------|----------|
| `telegraf-starlark.conf` | Starlark processor | Extensibility, complex logic |
| `telegraf-routing.conf` | Template routing | Maximum throughput |

---

## Solution 1: Starlark Processor

**File:** `telegraf-starlark.conf`

### How It Works

A centralized Starlark processor applies KPI splitting rules from a dictionary:

```starlark
SUFFIX_RULES = {
    "mme": [
        ("MME_Inter_", "-inter"),    # 23 KPIs → mme-inter
        ("MME_Intra_", "-intra"),    # 20 KPIs → mme-intra
        ("EMM_", "-emm"),            # Part of 22 KPIs → mme-emm
        ("EPS_", "-emm"),
        ("Intra_MME_", "-emm"),
        ("UE_", "-emm"),
        ("S1_Paging", "-emm"),
        ("MME_PDN_", "-pdn"),        # Part of 28 KPIs → mme-pdn
        ("MME_DCNR_", "-pdn"),
        # Remaining 34 KPIs use base objectType (no suffix)
    ],
}
```

### Adding New Rules

To add splitting for another schema (e.g., when `egtpc` exceeds 40 KPIs):

```starlark
SUFFIX_RULES = {
    "mme": [...],
    "egtpc": [
        ("Create_Bearer_", "-bearer"),
        ("Create_Session_", "-session"),
    ],
}
```

**Effort:** 1-2 lines of code

### Advantages

- ✅ All rules in one location
- ✅ Easy to extend (1 line per rule)
- ✅ Supports complex logic (regex, counters)
- ✅ State management for safety checks

### Disadvantages

- ❌ Requires Starlark knowledge
- ❌ Script interpretation overhead
- ❌ Harder to debug than config

---

## Solution 2: Template Routing

**File:** `telegraf-routing.conf`

### How It Works

Uses Telegraf's native `namepass`/`namedrop` filters with template processors:

```toml
# Group 1: Inter metrics → mme-inter
[[processors.template]]
  order = 7
  namepass = ["MME_Inter_*"]
  tag = "objectType"
  template = 'cisco-mobilitycore-pm-{{ .Tag "schema" }}-inter'
  [processors.template.tagpass]
  schema = ["mme"]

# Repeat for sessionName, sessionId...
```

### Adding New Rules

To add splitting for another pattern, add **3 processor blocks**:

```toml
# objectType
[[processors.template]]
  namepass = ["NEW_PATTERN_*"]
  tag = "objectType"
  template = 'cisco-mobilitycore-pm-{{ .Tag "schema" }}-newsuffix'

# sessionName
[[processors.template]]
  namepass = ["NEW_PATTERN_*"]
  tag = "sessionName"
  template = '{{ .Tag "device" }}_{{ .Tag "index" }}_cisco-mobilitycore-pm-{{ .Tag "schema" }}-newsuffix'

# sessionId  
[[processors.template]]
  namepass = ["NEW_PATTERN_*"]
  tag = "sessionId"
  template = '{{ .Tag "node_id" }}_{{ .Tag "index" }}_cisco-mobilitycore-pm-{{ .Tag "schema" }}-newsuffix'
```

Then update default processors to exclude the new pattern in `namedrop`.

**Effort:** 6+ locations to modify

### Advantages

- ✅ Native Go performance (fastest)
- ✅ Declarative configuration
- ✅ Familiar TOML syntax
- ✅ IDE syntax highlighting

### Disadvantages

- ❌ Rules scattered across many blocks
- ❌ Adding rules requires 6+ changes
- ❌ Risk of missing updates

---

## Benchmark Results

Performance comparison using Kafka-based testing with 704 messages:

| Metric | Starlark | Routing | Winner |
|--------|----------|---------|--------|
| **Throughput** | 458 metrics/sec | 691 metrics/sec | **Routing (+50%)** |
| **Processing Time** | 15.37s | 15.28s | Routing |
| **Memory Usage** | 146 MB | 154 MB | **Starlark (-5%)** |

### Key Findings

1. **Routing is ~50% faster** - Native Go template processors execute faster than Starlark interpreter
2. **Starlark uses ~5% less memory** - Routing creates more processor instances
3. **Both produce identical output** - Verified by integration tests

### Recommendation

| Use Case | Recommended |
|----------|-------------|
| Performance critical | **Routing** |
| Frequent rule changes | **Starlark** |
| Memory constrained | **Starlark** |
| Simple maintenance | **Starlark** |

---

## 40 KPI Limit Safety Check

Both configurations include a **runtime safety check** that:

1. **Tracks unique KPIs** per objectType
2. **Generates alerts** when approaching limit (35+ KPIs)
3. **Drops metrics** that exceed 40 KPIs
4. **Throttles alerts** to 1 per minute per objectType
5. **Resets counters** every 5 minutes

### Alert Metrics

When limits are violated, `kpi_limit_alert` metrics are generated:

| Severity | Trigger | Fields |
|----------|---------|--------|
| `warning` | 35+ KPIs | `current_count`, `limit`, `message` |
| `critical` | 40+ KPIs (dropped) | `dropped_kpi`, `dropped_total`, `message` |
| `critical` | Period summary | `dropped_count`, `unique_kpis` |

### Alert Output Options

Three output options are provided (choose one):

```toml
# Option 1: WebSocket to telemetry-collector (uncomment to enable)
# [[outputs.websocket]]
#   namepass = ["kpi_limit_alert"]
#   url = "ws://127.0.0.1:3000/alerts"
#   data_format = "json"

# Option 2: HTTP POST to alerting API
# [[outputs.http]]
#   namepass = ["kpi_limit_alert"]
#   url = "http://127.0.0.1:3000/api/alerts"
#   method = "POST"
#   data_format = "json"

# Option 3: Log file (enabled by default)
[[outputs.file]]
  namepass = ["kpi_limit_alert"]
  files = ["/tmp/kpi_alerts.log"]
  data_format = "json"
  rotation_max_size = "10MB"
  rotation_max_archives = 5
```

### Example Alert (JSON)

```json
{
  "name": "kpi_limit_alert",
  "tags": {
    "objectType": "cisco-mobilitycore-pm-mme-extra",
    "schema": "mme",
    "severity": "critical",
    "alert_type": "limit_exceeded"
  },
  "fields": {
    "dropped_kpi": "New_Unexpected_KPI",
    "current_count": 40,
    "limit": 40,
    "dropped_total": 1,
    "message": "KPI limit exceeded for cisco-mobilitycore-pm-mme-extra: New_Unexpected_KPI dropped (total: 1)"
  }
}
```

### Monitoring Recommendations

1. **Monitor `/tmp/kpi_alerts.log`** for any alert entries
2. **Set up log aggregation** (Splunk/ELK) to parse alert JSON
3. **Create dashboards** for `kpi_limit_alert` metrics
4. **Alert on `severity=critical`** for immediate investigation

---

## Configuration Files

| File | Purpose | Lines |
|------|---------|-------|
| `telegraf.conf` | Original config (no splitting) | 159 |
| `telegraf-starlark.conf` | Starlark-based splitting | 400 |
| `telegraf-routing.conf` | Template routing splitting | 445 |

---

## Deployment

### Prerequisites

- Telegraf 1.34+ (Starlark processor support)
- Kafka broker accessible at `{{server_ip}}:9092`

### Steps

1. Choose configuration based on requirements:
   - Performance priority → `telegraf-routing.conf`
   - Extensibility priority → `telegraf-starlark.conf`

2. Replace `{{server_ip}}` placeholder:
   ```bash
   sed -i 's/{{server_ip}}/your-kafka-host/g' telegraf-starlark.conf
   ```

3. Configure alert output (edit outputs section)

4. Validate configuration:
   ```bash
   telegraf --config telegraf-starlark.conf --test
   ```

5. Deploy and monitor `/tmp/kpi_alerts.log`

---

## Future Extensions

When adding new schemas that exceed 40 KPIs:

### For Starlark

Add rules to `SUFFIX_RULES` dictionary:

```starlark
SUFFIX_RULES = {
    "mme": [...],
    "new_schema": [
        ("PREFIX_A_", "-group-a"),
        ("PREFIX_B_", "-group-b"),
    ],
}
```

### For Routing

1. Add 3 template processor blocks per group
2. Update default processors' `namedrop` lists
3. Update documentation

---

## Support

For issues or questions:
- Check `/tmp/kpi_alerts.log` for limit violations
- Enable Telegraf debug mode: `debug = true`
- Review this documentation for configuration guidance
