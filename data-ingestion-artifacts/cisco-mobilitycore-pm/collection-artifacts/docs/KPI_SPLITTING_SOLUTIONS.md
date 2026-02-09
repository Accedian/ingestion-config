# KPI Splitting Solutions for Cisco Mobility Core PM

## Problem Statement

Rakuten's Matrix schema has a **40 KPI per objectType limit**. The MME schema contains 150 unique KPIs, exceeding this limit and requiring automatic splitting into multiple objectTypes.

### Source of Truth

**KPI mappings are defined in the CSV catalog:** `Kpi_calc_Kpicatalog-updatedGrouping.csv`

A Go tool is available to generate splitting rules from this catalog:
```bash
# Generate summary of all schemas
go run generate_splitting_rules.go -format summary

# Generate Starlark rules for MME
go run generate_splitting_rules.go -format starlark -schema mme

# Generate routing rules
go run generate_splitting_rules.go -format routing -schema mme

# Validate all groups are under 40 KPIs
go run generate_splitting_rules.go -format validate
```

### Current Schema Analysis (from CSV Catalog)

| Schema | Unique KPIs | Status | Action |
|--------|-------------|--------|--------|
| `mme` | 150 | **Exceeds limit** | Split into 7 objectTypes |
| `egtpc` | 30 | ✅ Under limit | No action needed |
| `schema` | 30 | ✅ Under limit | No action needed |
| `mme-paging-profile` | 20 | ✅ Under limit | No action needed |
| `tai` | 14 | ✅ Under limit | No action needed |
| `card` | 11 | ✅ Under limit | No action needed |
| `hss` | 11 | ✅ Under limit | No action needed |
| `port` | 11 | ✅ Under limit | No action needed |
| `apn` | 8 | ✅ Under limit | No action needed |
| `sbc` | 4 | ✅ Under limit | No action needed |
| `sx` | 4 | ✅ Under limit | No action needed |

### MME Schema Split

The 150 MME KPIs are split into **7 objectTypes**, each under 40 KPIs:

| ObjectType | KPI Count | Description |
|------------|-----------|-------------|
| `cisco-mobilitycore-pm-mme` | 26 | Base KPIs (not in split groups) |
| `cisco-mobilitycore-pm-mme-failure` | 24 | EMM_Attach_Reject_*, MME_*_Failure*, etc. |
| `cisco-mobilitycore-pm-mme-inter` | 24 | MME_Inter_* KPIs |
| `cisco-mobilitycore-pm-mme-sr` | 22 | Success rate KPIs |
| `cisco-mobilitycore-pm-mme-intra` | 20 | MME_Intra_* KPIs |
| `cisco-mobilitycore-pm-mme-dcnr` | 19 | MME_DCNR_* KPIs |
| `cisco-mobilitycore-pm-mme-pdn` | 15 | MME_PDN_* KPIs |
| **Total** | **150** | |

---

## Available Solutions

Two production-ready configurations are provided, both using **exact KPI name matching**:

| File | Approach | Best For |
|------|----------|----------|
| `telegraf-starlark.conf` | Starlark processor | Extensibility, complex logic |
| `telegraf-routing.conf` | Template routing | Maximum throughput |

---

## Solution 1: Starlark Processor (Recommended for Production)

**File:** `telegraf-starlark.conf`

### How It Works

A centralized Starlark processor uses **EXACT_RULES** - a dictionary mapping each KPI name to its objectType suffix. This is the safest approach for production as it uses exact matches only.

```starlark
# Source of truth: Kpi_calc_Kpicatalog-updatedGrouping.csv
# Logic: Use column 3 (ObjectIdentifier) if defined, otherwise column 2 (Schema)

EXACT_RULES = {
    # mme-inter: 24 KPIs
    "MME_Inter_MME_TAU_Reject_eps_not_allowed": "-inter",
    "MME_Inter_MME_TAU_Reject_illegal_me": "-inter",
    ...
    
    # mme-failure: 24 KPIs
    "EMM_Attach_Reject_Decode_Failure": "-failure",
    "EMM_Attach_Reject_EPS_Not_Allowed": "-failure",
    ...
    
    # mme-sr: 22 KPIs
    "MME_Overall_Attach_Success_Rate": "-sr",
    ...
}

def apply(metric):
    kpi = metric.name
    suffix = EXACT_RULES.get(kpi, "")  # Simple dict lookup
    ...
```

### Adding New Rules

When the CSV catalog is updated:

```bash
# Regenerate the exact rules from CSV
go run generate_splitting_rules.go -format starlark -schema <name>

# Copy the output into the EXACT_RULES dictionary
```

**Effort:** Copy/paste from tool output

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

## Solution 2: Template Routing (EXACT RULES)

**File:** `telegraf-routing.conf`

### How It Works

Uses Telegraf's native `namepass`/`namedrop` filters with template processors. Now uses **exact KPI names** (no wildcards) for production safety:

```toml
# Group 1: Inter metrics (24 exact KPIs) → mme-inter
[[processors.template]]
  order = 7
  namepass = ["MME_Inter_MME_Relocations_S1_S10_HO_Incoming_Success_Rate", "MME_Inter_MME_TAU_Reject_eps_not_allowed", ...]
  tag = "objectType"
  template = 'cisco-mobilitycore-pm-{{ .Tag "schema" }}-inter'
  [processors.template.tagpass]
  schema = ["mme"]

# Repeat for sessionName, sessionId...
```

### Adding New Rules

Use the generate_splitting_rules.go tool to regenerate rules from CSV:

```bash
go run generate_splitting_rules.go -format routing -schema mme
```

Then copy the output into the config file.

**Effort:** Single command, copy/paste output

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
| `telegraf-starlark.conf` | Starlark-based splitting | ~430 |
| `telegraf-routing.conf` | Template routing splitting | ~450 |
| `Kpi_calc_Kpicatalog-updatedGrouping.csv` | KPI catalog (source of truth) | ~300 |
| `generate_splitting_rules.go` | Rule generation tool | ~250 |

---

## Generate Splitting Rules Tool

A Go tool is provided to generate Telegraf splitting rules from the KPI CSV catalog.

### Usage

```bash
# Show summary of all schemas and their KPI counts
go run generate_splitting_rules.go -format summary

# Generate Starlark rules for all schemas
go run generate_splitting_rules.go -format starlark

# Generate Starlark rules for a specific schema
go run generate_splitting_rules.go -format starlark -schema mme

# Generate routing rules
go run generate_splitting_rules.go -format routing -schema mme

# Export as JSON
go run generate_splitting_rules.go -format json

# Validate all objectTypes are under 40 KPIs
go run generate_splitting_rules.go -format validate
```

### Example Output (summary)

```
MME (150 KPIs) - ⚠️ SPLIT INTO 7 GROUPS
--------------------------------------------------
  ✓ (base)           26 KPIs → cisco-mobilitycore-pm-mme
  ✓ -failure         24 KPIs → cisco-mobilitycore-pm-mme-failure
  ✓ -inter           24 KPIs → cisco-mobilitycore-pm-mme-inter
  ✓ -sr              22 KPIs → cisco-mobilitycore-pm-mme-sr
  ✓ -intra           20 KPIs → cisco-mobilitycore-pm-mme-intra
  ✓ -dcnr            19 KPIs → cisco-mobilitycore-pm-mme-dcnr
  ✓ -pdn             15 KPIs → cisco-mobilitycore-pm-mme-pdn
```

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
