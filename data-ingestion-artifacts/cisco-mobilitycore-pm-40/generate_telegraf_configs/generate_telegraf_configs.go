// generate_telegraf_configs.go - Generate complete Telegraf KPI splitting configurations
//
// This tool reads the KPI catalog CSV and generates:
//   - telegraf-starlark.conf (Starlark processor implementation)
//   - telegraf-routing.conf (Template routing implementation)
//   - Ingestion dictionaries (optional, with -dictionaries flag)
//
// The tool validates all groups are under 40 KPIs before generating configs.
//
// Usage:
//
//	go run generate_telegraf_configs.go -csv Kpi_calc_Kpicatalog-updatedGrouping.csv -output ./generated
//
// With dictionary generation:
//
//	go run generate_telegraf_configs.go -csv Kpi_calc_Kpicatalog-updatedGrouping.csv -output ./generated -dictionaries ./dictionaries
package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

const (
	MaxKPIsPerObjectType = 40
	WarningThreshold     = 35
)

// KPIEntry represents a single KPI from the CSV catalog
type KPIEntry struct {
	KPIName          string
	Schema           string
	ObjectIdentifier string
}

// SchemaGroup represents a group of KPIs for a specific objectType
type SchemaGroup struct {
	Schema     string
	Suffix     string
	ObjectType string
	KPIs       []string
	Count      int
}

// SplittingRules contains all schema groupings
type SplittingRules struct {
	Schemas map[string][]SchemaGroup
}

// Config holds the configuration for the generator
type Config struct {
	CSVFile         string
	OutputDir       string
	DictionariesDir string
	KafkaBroker     string
	KafkaTopic      string
	AlertLogPath    string
}

func main() {
	cfg := Config{}
	flag.StringVar(&cfg.CSVFile, "csv", "Kpi_calc_Kpicatalog-updatedGrouping.csv", "Path to KPI catalog CSV")
	flag.StringVar(&cfg.OutputDir, "output", ".", "Output directory for generated configs")
	flag.StringVar(&cfg.DictionariesDir, "dictionaries", "", "Output directory for ingestion dictionaries (optional)")
	flag.StringVar(&cfg.KafkaBroker, "broker", "{{server_ip}}:9092", "Kafka broker address")
	flag.StringVar(&cfg.KafkaTopic, "topic", "pca_kpi_topic", "Kafka topic name")
	flag.StringVar(&cfg.AlertLogPath, "alert-log", "/tmp/kpi_alerts.log", "Path for KPI alert log file")
	flag.Parse()

	fmt.Println("=" + strings.Repeat("=", 70))
	fmt.Println("TELEGRAF CONFIG GENERATOR")
	fmt.Println("=" + strings.Repeat("=", 70))
	fmt.Printf("CSV: %s\n", cfg.CSVFile)
	fmt.Printf("Output: %s\n", cfg.OutputDir)
	fmt.Println()

	// Parse CSV
	entries, err := parseCSV(cfg.CSVFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing CSV: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Parsed %d KPI entries from CSV\n", len(entries))

	// Build rules
	rules := buildRules(entries)

	// Validate before generation
	fmt.Println("\n--- VALIDATION ---")
	if !validate(rules) {
		fmt.Fprintf(os.Stderr, "\n❌ Validation failed. Fix CSV groupings before generating configs.\n")
		os.Exit(1)
	}
	fmt.Println("✓ All objectTypes within 40 KPI limit")

	// Print summary
	printSummary(rules)

	// Create output directory
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Generate configs
	fmt.Println("\n--- GENERATING CONFIGS ---")

	starlarkPath := filepath.Join(cfg.OutputDir, "telegraf-starlark.conf")
	if err := generateStarlarkConfig(starlarkPath, rules, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating Starlark config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Generated %s\n", starlarkPath)

	routingPath := filepath.Join(cfg.OutputDir, "telegraf-routing.conf")
	if err := generateRoutingConfig(routingPath, rules, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating Routing config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Generated %s\n", routingPath)

	// Generate dictionaries if requested
	if cfg.DictionariesDir != "" {
		fmt.Println("\n--- GENERATING DICTIONARIES ---")
		if err := os.MkdirAll(cfg.DictionariesDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating dictionaries directory: %v\n", err)
			os.Exit(1)
		}
		count, err := generateDictionaries(cfg.DictionariesDir, rules)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating dictionaries: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated %d dictionary files in %s\n", count, cfg.DictionariesDir)
	}

	fmt.Println("\n✓ Done!")
}

func parseCSV(filename string) ([]KPIEntry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.Comma = ';'
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var entries []KPIEntry
	seen := make(map[string]int) // KPI -> line number where first seen
	var duplicates []string

	for i, record := range records {
		if i == 0 || len(record) < 2 {
			continue
		}
		kpi := strings.TrimSpace(record[0])
		schema := strings.TrimSpace(record[1])
		objID := ""
		if len(record) >= 3 {
			objID = strings.TrimSpace(record[2])
		}
		if kpi != "" && schema != "" {
			if firstLine, exists := seen[kpi]; exists {
				duplicates = append(duplicates, fmt.Sprintf("  Line %d: %s (first seen at line %d)", i+1, kpi, firstLine))
			} else {
				seen[kpi] = i + 1
				entries = append(entries, KPIEntry{kpi, strings.ToLower(schema), objID})
			}
		}
	}

	if len(duplicates) > 0 {
		fmt.Printf("\n⚠️  WARNING: Found %d duplicate KPI(s) in CSV (skipped):\n", len(duplicates))
		for _, d := range duplicates {
			fmt.Println(d)
		}
		fmt.Println()
	}

	return entries, nil
}

func buildRules(entries []KPIEntry) SplittingRules {
	schemaGroups := make(map[string]map[string][]string)

	for _, e := range entries {
		if schemaGroups[e.Schema] == nil {
			schemaGroups[e.Schema] = make(map[string][]string)
		}
		objID := e.ObjectIdentifier
		if objID == "" || strings.EqualFold(objID, e.Schema) {
			objID = ""
		}
		schemaGroups[e.Schema][objID] = append(schemaGroups[e.Schema][objID], e.KPIName)
	}

	rules := SplittingRules{Schemas: make(map[string][]SchemaGroup)}

	for schema, groups := range schemaGroups {
		var list []SchemaGroup
		for objID, kpis := range groups {
			suffix := ""
			objectType := "cisco-mobilitycore-pm-" + schema
			if objID != "" {
				suffix = strings.ToLower(objID)
				suffix = strings.TrimPrefix(suffix, strings.ToLower(schema)+"_")
				suffix = strings.TrimPrefix(suffix, strings.ToLower(schema))
				if suffix != "" && !strings.HasPrefix(suffix, "-") {
					suffix = "-" + suffix
				}
				objectType = "cisco-mobilitycore-pm-" + schema + suffix
			}
			sort.Strings(kpis)
			list = append(list, SchemaGroup{schema, suffix, objectType, kpis, len(kpis)})
		}
		sort.Slice(list, func(i, j int) bool { return list[i].Count > list[j].Count })
		rules.Schemas[schema] = list
	}
	return rules
}

func validate(rules SplittingRules) bool {
	valid := true
	for _, groups := range rules.Schemas {
		for _, g := range groups {
			if g.Count > MaxKPIsPerObjectType {
				fmt.Printf("❌ %s: %d KPIs (exceeds %d)\n", g.ObjectType, g.Count, MaxKPIsPerObjectType)
				valid = false
			}
		}
	}
	return valid
}

func printSummary(rules SplittingRules) {
	fmt.Println("\n--- SCHEMA SUMMARY ---")
	schemas := sortedSchemas(rules)
	for _, schema := range schemas {
		groups := rules.Schemas[schema]
		total := 0
		for _, g := range groups {
			total += g.Count
		}
		status := "✓ OK"
		if len(groups) > 1 {
			status = fmt.Sprintf("⚠️ SPLIT (%d groups)", len(groups))
		}
		fmt.Printf("  %-25s %3d KPIs  %s\n", strings.ToUpper(schema), total, status)
	}
}

func sortedSchemas(rules SplittingRules) []string {
	var s []string
	for k := range rules.Schemas {
		s = append(s, k)
	}
	sort.Strings(s)
	return s
}

// =============================================================================
// STARLARK CONFIG GENERATION
// =============================================================================

func generateStarlarkConfig(path string, rules SplittingRules, cfg Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	writeHeader(w, "STARLARK-BASED IMPLEMENTATION", cfg)
	writeKafkaInput(w, cfg)
	writeInternalHandler(w)
	writeIndexCleanup(w)
	writeStarlarkSplitting(w, rules)
	writeP2PHandler(w)
	writeP2PFilter(w)
	writeRenameProcessor(w)
	writeSafetyCheck(w)
	writeTagLimit(w)
	writeAlertOutput(w, cfg)

	return nil
}

func writeHeader(w *bufio.Writer, impl string, cfg Config) {
	fmt.Fprintf(w, `#######################
# Cisco PCA Fault and Mobility PM Ingestion Pipeline
# %s
#
# AUTO-GENERATED by generate_telegraf_configs.go
# Generated: %s
# Source CSV: %s
#
# DO NOT EDIT MANUALLY - regenerate from CSV catalog
#######################

`, impl, time.Now().Format("2006-01-02 15:04:05"), cfg.CSVFile)
}

func writeKafkaInput(w *bufio.Writer, cfg Config) {
	fmt.Fprintf(w, `[[inputs.kafka_consumer]]
  interval = "60s"
  brokers = ["%s"]
  topics = ["%s"]
  offset = "oldest"
  max_message_len = 1000000
  data_format = "json"
  json_query = "@this"
  json_name_key = "kpi"
  json_time_key = "timestamp"
  json_time_format = "unix"
  tag_keys = ["device", "kpi","index","node_id","schema", "source_ip", "node_ip"]

`, cfg.KafkaBroker, cfg.KafkaTopic)
}

func writeIndexCleanup(w *bufio.Writer) {
	fmt.Fprint(w, `# Processor: Clean "index" tag - remove square brackets
[[processors.regex]]
  order = 3
  namedrop = ["internal_*"]
  [[processors.regex.tags]]
    key = "index"
    pattern = '^\["(.*?)"\]$'
    replacement = '${1}'

# Processor: Clean "index" tag - replace commas with underscores
[[processors.strings]]
  order = 4
  [[processors.strings.replace]]
    tag = "index"
    old = ","
    new = "_"

`)
}

func writeStarlarkSplitting(w *bufio.Writer, rules SplittingRules) {
	fmt.Fprint(w, `#######################
# KPI SPLITTING PROCESSOR (EXACT RULES)
#
# Routes KPIs to objectTypes based on exact name matching from CSV catalog.
# KPIs not in EXACT_RULES use base objectType (schema only).
#######################
[[processors.starlark]]
  order = 7
  namedrop = ["internal_*"]
  
  source = '''
# =============================================================================
# EXACT_RULES: Maps each KPI name to its target objectType suffix
# Source of truth: CSV catalog (column 3 if defined, else column 2)
# =============================================================================

EXACT_RULES = {
`)

	// Generate EXACT_RULES dictionary
	schemas := sortedSchemas(rules)
	for _, schema := range schemas {
		groups := rules.Schemas[schema]
		if len(groups) <= 1 {
			continue // No splitting needed
		}

		total := 0
		for _, g := range groups {
			total += g.Count
		}

		fmt.Fprintf(w, "\n    # %s (%d KPIs → %d objectTypes)\n", strings.ToUpper(schema), total, len(groups))

		for _, g := range groups {
			if g.Suffix == "" {
				fmt.Fprintf(w, "    # Base %s: %d KPIs (no entry = falls through to base)\n", schema, g.Count)
				continue
			}
			fmt.Fprintf(w, "    # %s%s: %d KPIs\n", schema, g.Suffix, g.Count)
			for _, kpi := range g.KPIs {
				fmt.Fprintf(w, "    \"%s\": \"%s\",\n", kpi, g.Suffix)
			}
		}
	}

	fmt.Fprint(w, `}

def apply(metric):
    kpi = metric.name
    schema = metric.tags.get("schema", "")
    device = metric.tags.get("device", "")
    index = metric.tags.get("index", "")
    node_id = metric.tags.get("node_id", "")
    
    # Lookup suffix from EXACT_RULES (empty string if not found = base objectType)
    suffix = EXACT_RULES.get(kpi, "")
    
    # Build object type
    object_type_base = "cisco-mobilitycore-pm-" + schema
    object_type = object_type_base + suffix
    
    # Build session identifiers
    if schema == "p2p":
        session_name = device + "_" + object_type
        session_id = node_id + "_" + object_type
    else:
        session_name = device + "_" + index + "_" + object_type
        session_id = node_id + "_" + index + "_" + object_type
    
    metric.tags["sessionName"] = session_name
    metric.tags["sessionId"] = session_id
    metric.tags["objectType"] = object_type
    metric.tags["direction"] = "-1"
    
    return metric
'''
  [processors.starlark.tagdrop]
  schema = ["p2p"]

`)
}

func writeP2PHandler(w *bufio.Writer) {
	fmt.Fprint(w, `# Handle p2p schema separately (no index in session identifiers)
[[processors.starlark]]
  order = 7
  namedrop = ["internal_*"]
  
  source = '''
def apply(metric):
    schema = metric.tags.get("schema", "")
    device = metric.tags.get("device", "")
    node_id = metric.tags.get("node_id", "")
    
    object_type = "cisco-mobilitycore-pm-" + schema
    session_name = device + "_" + object_type
    session_id = node_id + "_" + object_type
    
    metric.tags["sessionName"] = session_name
    metric.tags["sessionId"] = session_id
    metric.tags["objectType"] = object_type
    metric.tags["direction"] = "-1"
    
    return metric
'''
  [processors.starlark.tagpass]
  schema = ["p2p"]

`)
}

func writeP2PFilter(w *bufio.Writer) {
	fmt.Fprint(w, `# Filter out p2p_protocol objects
[[processors.starlark]]
  order = 9
  namedrop = ["internal_*"]

  source = '''
def apply(metric):
    index = metric.tags.get("index", "")
    if "p2p_protocol#" in index:
        return []
    else:
        return metric
  '''
  [processors.starlark.tagpass]
  schema = ["p2p"]

`)
}

func writeRenameProcessor(w *bufio.Writer) {
	fmt.Fprint(w, `[[processors.rename]]
  order = 8
  [[processors.rename.replace]]
    tag = "node_ip"
    dest = "source_ip"

`)
}

func writeSafetyCheck(w *bufio.Writer) {
	fmt.Fprint(w, `#######################
# KPI LIMIT MONITOR (Monitor-Only Mode)
#
# With EXACT_RULES, KPI distribution is guaranteed by the CSV catalog.
# This processor MONITORS for unexpected KPIs (not in catalog) that
# could cause the base objectType to exceed the 40 KPI limit.
#
# MONITOR-ONLY: Metrics are NEVER dropped - only alerts are generated.
#
# Alert metrics (kpi_limit_alert) are created with severity levels:
#   - warning: approaching limit (35+ KPIs) - investigate CSV catalog
#   - critical: limit exceeded - update CSV catalog with new KPI mappings
#######################
[[processors.starlark]]
  order = 10
  namedrop = ["internal_*", "kpi_limit_alert"]
  
  source = '''
MAX_KPIS_PER_OBJECT = 40
WARNING_THRESHOLD = 35
RESET_WINDOW_SECONDS = 300

def apply(metric):
    object_type = metric.tags.get("objectType", "unknown")
    schema = metric.tags.get("schema", "")
    kpi_name = metric.name
    current_time = metric.time
    
    if object_type not in state:
        state[object_type] = {
            "kpis": {},
            "last_reset": current_time,
            "exceeded_count": 0,
            "last_alert_time": 0,
        }
    
    obj_state = state[object_type]
    
    # Reset window check
    time_diff = current_time - obj_state["last_reset"]
    if time_diff > RESET_WINDOW_SECONDS * 1000000000:
        if obj_state["exceeded_count"] > 0:
            alert = Metric("kpi_limit_alert")
            alert.tags["objectType"] = object_type
            alert.tags["schema"] = schema
            alert.tags["severity"] = "critical"
            alert.tags["alert_type"] = "period_summary"
            alert.fields["exceeded_count"] = obj_state["exceeded_count"]
            alert.fields["unique_kpis"] = len(obj_state["kpis"])
            alert.fields["message"] = "Period ended: %d KPIs exceeded limit for %s - update CSV catalog" % (
                obj_state["exceeded_count"], object_type)
            alert.time = current_time
            obj_state["kpis"] = {}
            obj_state["last_reset"] = current_time
            obj_state["exceeded_count"] = 0
            obj_state["last_alert_time"] = 0
            return [metric, alert]
        obj_state["kpis"] = {}
        obj_state["last_reset"] = current_time
        obj_state["exceeded_count"] = 0
        obj_state["last_alert_time"] = 0
    
    if kpi_name in obj_state["kpis"]:
        return metric
    
    current_count = len(obj_state["kpis"])
    
    if current_count >= MAX_KPIS_PER_OBJECT:
        # MONITOR-ONLY: Track but do NOT drop
        obj_state["exceeded_count"] = obj_state["exceeded_count"] + 1
        obj_state["kpis"][kpi_name] = True
        
        alert_interval = 60 * 1000000000
        if current_time - obj_state["last_alert_time"] > alert_interval:
            obj_state["last_alert_time"] = current_time
            alert = Metric("kpi_limit_alert")
            alert.tags["objectType"] = object_type
            alert.tags["schema"] = schema
            alert.tags["severity"] = "critical"
            alert.tags["alert_type"] = "limit_exceeded"
            alert.fields["unexpected_kpi"] = kpi_name
            alert.fields["current_count"] = current_count + 1
            alert.fields["limit"] = MAX_KPIS_PER_OBJECT
            alert.fields["exceeded_total"] = obj_state["exceeded_count"]
            alert.fields["message"] = "KPI limit exceeded for %s: %s not in CSV catalog (total: %d)" % (
                object_type, kpi_name, obj_state["exceeded_count"])
            alert.time = current_time
            return [metric, alert]
        
        return metric
    
    obj_state["kpis"][kpi_name] = True
    
    if current_count >= WARNING_THRESHOLD:
        alert_interval = 60 * 1000000000
        if current_time - obj_state.get("last_warning_time", 0) > alert_interval:
            obj_state["last_warning_time"] = current_time
            alert = Metric("kpi_limit_alert")
            alert.tags["objectType"] = object_type
            alert.tags["schema"] = schema
            alert.tags["severity"] = "warning"
            alert.tags["alert_type"] = "approaching_limit"
            alert.fields["current_count"] = current_count + 1
            alert.fields["limit"] = MAX_KPIS_PER_OBJECT
            alert.fields["message"] = "Approaching KPI limit for %s: %d/%d - verify CSV catalog" % (
                object_type, current_count + 1, MAX_KPIS_PER_OBJECT)
            alert.time = current_time
            return [metric, alert]
    
    return metric
'''

`)
}

func writeTagLimit(w *bufio.Writer) {
	fmt.Fprint(w, `[[processors.tag_limit]]
  order = 11
  namedrop = ["internal_*"]
  limit = 10
  keep = ["sessionId", "sessionName", "objectType", "direction", "source_ip"]
  [processors.tag_limit.tagdrop]
  schema = ["p2p",""]

`)
}

func writeAlertOutput(w *bufio.Writer, cfg Config) {
	fmt.Fprintf(w, `#######################
# ALERT OUTPUT
#
# Routes kpi_limit_alert metrics to log file for monitoring.
#######################
[[outputs.file]]
  namepass = ["kpi_limit_alert"]
  files = ["%s"]
  data_format = "json"
  rotation_max_size = "10MB"
  rotation_max_archives = 5
`, cfg.AlertLogPath)
}

func writeInternalHandler(w *bufio.Writer) {
	fmt.Fprint(w, `#######################
# INTERNAL METRICS HANDLER
#
# Adds required tags to Telegraf internal metrics (internal_*)
# These are used for agent health monitoring.
#######################
[[processors.starlark]]
  order = 2
  namepass = ["internal_*"]
  
  source = '''
def apply(metric):
    metric.tags["objectType"] = "internal"
    metric.tags["sessionId"] = "internal"
    metric.tags["sessionName"] = "internal"
    metric.tags["direction"] = "-1"
    return metric
'''

`)
}

// =============================================================================
// ROUTING CONFIG GENERATION
// =============================================================================

func generateRoutingConfig(path string, rules SplittingRules, cfg Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	writeHeader(w, "TEMPLATE ROUTING IMPLEMENTATION (EXACT MATCHES)", cfg)
	writeKafkaInput(w, cfg)
	writeInternalHandler(w)
	writeIndexCleanup(w)
	writeRoutingSplitting(w, rules)
	writeDefaultRouting(w, rules)
	writeP2PRoutingHandler(w)
	writeP2PFilter(w)
	writeRenameProcessor(w)
	writeSafetyCheck(w)
	writeTagLimit(w)
	writeAlertOutput(w, cfg)

	return nil
}

func writeRoutingSplitting(w *bufio.Writer, rules SplittingRules) {
	fmt.Fprint(w, `#######################
# KPI SPLITTING (TEMPLATE ROUTING WITH EXACT MATCHES)
#
# Each group has 3 template blocks: objectType, sessionName, sessionId
# Uses exact KPI names in namepass (no wildcards for production safety)
#######################
`)

	schemas := sortedSchemas(rules)
	groupNum := 0

	for _, schema := range schemas {
		groups := rules.Schemas[schema]
		if len(groups) <= 1 {
			continue // No splitting needed
		}

		for _, g := range groups {
			if g.Suffix == "" {
				continue // Base group handled separately
			}
			groupNum++

			// Build namepass list
			patterns := make([]string, len(g.KPIs))
			for i, kpi := range g.KPIs {
				patterns[i] = fmt.Sprintf("\"%s\"", kpi)
			}
			namepass := strings.Join(patterns, ", ")

			fmt.Fprintf(w, "\n#######################\n")
			fmt.Fprintf(w, "# GROUP %d: %s%s (%d KPIs)\n", groupNum, strings.ToUpper(schema), g.Suffix, g.Count)
			fmt.Fprintf(w, "#######################\n")

			// objectType template
			fmt.Fprintf(w, "\n[[processors.template]]\n")
			fmt.Fprintf(w, "  order = 7\n")
			fmt.Fprintf(w, "  namedrop = [\"internal_*\"]\n")
			fmt.Fprintf(w, "  namepass = [%s]\n", namepass)
			fmt.Fprintf(w, "  tag = \"objectType\"\n")
			fmt.Fprintf(w, "  template = 'cisco-mobilitycore-pm-{{ .Tag \"schema\" }}%s'\n", g.Suffix)
			fmt.Fprintf(w, "  [processors.template.tagpass]\n")
			fmt.Fprintf(w, "  schema = [\"%s\"]\n", schema)

			// sessionName template
			fmt.Fprintf(w, "\n[[processors.template]]\n")
			fmt.Fprintf(w, "  order = 7\n")
			fmt.Fprintf(w, "  namedrop = [\"internal_*\"]\n")
			fmt.Fprintf(w, "  namepass = [%s]\n", namepass)
			fmt.Fprintf(w, "  tag = \"sessionName\"\n")
			fmt.Fprintf(w, "  template = '{{ .Tag \"device\" }}_{{ .Tag \"index\" }}_cisco-mobilitycore-pm-{{ .Tag \"schema\" }}%s'\n", g.Suffix)
			fmt.Fprintf(w, "  [processors.template.tagpass]\n")
			fmt.Fprintf(w, "  schema = [\"%s\"]\n", schema)

			// sessionId template
			fmt.Fprintf(w, "\n[[processors.template]]\n")
			fmt.Fprintf(w, "  order = 7\n")
			fmt.Fprintf(w, "  namedrop = [\"internal_*\"]\n")
			fmt.Fprintf(w, "  namepass = [%s]\n", namepass)
			fmt.Fprintf(w, "  tag = \"sessionId\"\n")
			fmt.Fprintf(w, "  template = '{{ .Tag \"node_id\" }}_{{ .Tag \"index\" }}_cisco-mobilitycore-pm-{{ .Tag \"schema\" }}%s'\n", g.Suffix)
			fmt.Fprintf(w, "  [processors.template.tagpass]\n")
			fmt.Fprintf(w, "  schema = [\"%s\"]\n", schema)
		}
	}
	fmt.Fprintln(w)
}

func writeDefaultRouting(w *bufio.Writer, rules SplittingRules) {
	schemas := sortedSchemas(rules)

	for _, schema := range schemas {
		groups := rules.Schemas[schema]
		if len(groups) <= 1 {
			continue // No splitting = no special base handling needed
		}

		// Collect all non-base KPIs for namedrop
		var allNonBaseKPIs []string
		baseCount := 0
		for _, g := range groups {
			if g.Suffix == "" {
				baseCount = g.Count
			} else {
				allNonBaseKPIs = append(allNonBaseKPIs, g.KPIs...)
			}
		}

		if len(allNonBaseKPIs) == 0 {
			continue
		}

		// Build namedrop list
		namedrop := make([]string, len(allNonBaseKPIs)+1)
		namedrop[0] = "\"internal_*\""
		for i, kpi := range allNonBaseKPIs {
			namedrop[i+1] = fmt.Sprintf("\"%s\"", kpi)
		}
		namedropStr := strings.Join(namedrop, ", ")

		fmt.Fprintf(w, "#######################\n")
		fmt.Fprintf(w, "# DEFAULT: Base %s (%d KPIs) - excludes all split KPIs\n", strings.ToUpper(schema), baseCount)
		fmt.Fprintf(w, "#######################\n")

		// objectType template
		fmt.Fprintf(w, "\n[[processors.template]]\n")
		fmt.Fprintf(w, "  order = 8\n")
		fmt.Fprintf(w, "  namedrop = [%s]\n", namedropStr)
		fmt.Fprintf(w, "  tag = \"objectType\"\n")
		fmt.Fprintf(w, "  template = 'cisco-mobilitycore-pm-{{ .Tag \"schema\" }}'\n")
		fmt.Fprintf(w, "  [processors.template.tagpass]\n")
		fmt.Fprintf(w, "  schema = [\"%s\"]\n", schema)

		// sessionName template
		fmt.Fprintf(w, "\n[[processors.template]]\n")
		fmt.Fprintf(w, "  order = 8\n")
		fmt.Fprintf(w, "  namedrop = [%s]\n", namedropStr)
		fmt.Fprintf(w, "  tag = \"sessionName\"\n")
		fmt.Fprintf(w, "  template = '{{ .Tag \"device\" }}_{{ .Tag \"index\" }}_cisco-mobilitycore-pm-{{ .Tag \"schema\" }}'\n")
		fmt.Fprintf(w, "  [processors.template.tagpass]\n")
		fmt.Fprintf(w, "  schema = [\"%s\"]\n", schema)

		// sessionId template
		fmt.Fprintf(w, "\n[[processors.template]]\n")
		fmt.Fprintf(w, "  order = 8\n")
		fmt.Fprintf(w, "  namedrop = [%s]\n", namedropStr)
		fmt.Fprintf(w, "  tag = \"sessionId\"\n")
		fmt.Fprintf(w, "  template = '{{ .Tag \"node_id\" }}_{{ .Tag \"index\" }}_cisco-mobilitycore-pm-{{ .Tag \"schema\" }}'\n")
		fmt.Fprintf(w, "  [processors.template.tagpass]\n")
		fmt.Fprintf(w, "  schema = [\"%s\"]\n", schema)

		fmt.Fprintln(w)
	}
}

func writeP2PRoutingHandler(w *bufio.Writer) {
	fmt.Fprint(w, `#######################
# P2P SCHEMA HANDLER (no index in session identifiers)
#######################

[[processors.template]]
  order = 7
  namedrop = ["internal_*"]
  tag = "objectType"
  template = 'cisco-mobilitycore-pm-{{ .Tag "schema" }}'
  [processors.template.tagpass]
  schema = ["p2p"]

[[processors.template]]
  order = 7
  namedrop = ["internal_*"]
  tag = "sessionName"
  template = '{{ .Tag "device" }}_cisco-mobilitycore-pm-{{ .Tag "schema" }}'
  [processors.template.tagpass]
  schema = ["p2p"]

[[processors.template]]
  order = 7
  namedrop = ["internal_*"]
  tag = "sessionId"
  template = '{{ .Tag "node_id" }}_cisco-mobilitycore-pm-{{ .Tag "schema" }}'
  [processors.template.tagpass]
  schema = ["p2p"]

[[processors.template]]
  order = 7
  namedrop = ["internal_*"]
  tag = "direction"
  template = '-1'
  [processors.template.tagpass]
  schema = ["p2p"]

`)
}

// =============================================================================
// DICTIONARY GENERATION
// =============================================================================

// Dictionary represents an ADH Gather ingestion dictionary
type Dictionary struct {
	ID             string             `json:"_id"`
	CustomMetrics  interface{}        `json:"customMetrics"`
	DictionaryName string             `json:"dictionaryName"`
	DictionaryType string             `json:"dictionaryType"`
	Dimensions     []DictionaryDim    `json:"dimensions"`
	MetricType     string             `json:"metricType"`
	Metrics        []DictionaryMetric `json:"metrics"`
	ObjectType     string             `json:"objectType"`
	TenantID       string             `json:"tenantId"`
	Vendor         string             `json:"vendor"`
	ID2            string             `json:"id"`
	Type           string             `json:"type"`
}

// DictionaryDim represents a dimension in the dictionary
type DictionaryDim struct {
	AnalyticsName string `json:"analyticsName"`
	DataType      string `json:"dataType"`
	RawName       string `json:"rawName"`
}

// DictionaryMetric represents a metric in the dictionary
type DictionaryMetric struct {
	AnalyticsName string   `json:"analyticsName"`
	DataType      string   `json:"dataType"`
	Directions    []string `json:"directions"`
	RawName       string   `json:"rawName"`
	Unit          string   `json:"unit"`
}

// Standard dimensions used in all dictionaries
var standardDimensions = []DictionaryDim{
	{AnalyticsName: "monitoredObjectId", DataType: "string", RawName: "monitoredObjectId"},
	{AnalyticsName: "monitoredObjectName", DataType: "string", RawName: "monitoredObjectName"},
	{AnalyticsName: "timestamp", DataType: "long", RawName: "timestamp"},
	{AnalyticsName: "direction", DataType: "integer", RawName: "direction"},
	{AnalyticsName: "agentId", DataType: "string", RawName: "agentId"},
	{AnalyticsName: "agentName", DataType: "string", RawName: "agentName"},
	{AnalyticsName: "agentType", DataType: "string", RawName: "agentType"},
	{AnalyticsName: "host", DataType: "string", RawName: "host"},
	{AnalyticsName: "index", DataType: "string", RawName: "index"},
	{AnalyticsName: "node_id", DataType: "string", RawName: "node_id"},
	{AnalyticsName: "objectType", DataType: "string", RawName: "objectType"},
	{AnalyticsName: "schema", DataType: "string", RawName: "schema"},
	{AnalyticsName: "sessionId", DataType: "string", RawName: "sessionId"},
	{AnalyticsName: "sessionName", DataType: "string", RawName: "sessionName"},
	{AnalyticsName: "source_ip", DataType: "string", RawName: "source_ip"},
}

func generateDictionaries(outputDir string, rules SplittingRules) (int, error) {
	count := 0
	for _, groups := range rules.Schemas {
		for _, g := range groups {
			dict := createDictionary(g)
			filename := g.ObjectType + ".json"
			path := filepath.Join(outputDir, filename)

			data, err := json.MarshalIndent(dict, "", "    ")
			if err != nil {
				return count, fmt.Errorf("error marshaling dictionary for %s: %w", g.ObjectType, err)
			}

			if err := os.WriteFile(path, data, 0644); err != nil {
				return count, fmt.Errorf("error writing dictionary %s: %w", path, err)
			}
			count++
		}
	}
	return count, nil
}

func createDictionary(g SchemaGroup) Dictionary {
	// Create dictionary ID with "cisco-" prefix (matching existing pattern)
	dictID := "cisco-" + g.ObjectType

	// Create metrics from KPIs, detecting duplicates
	metrics := make([]DictionaryMetric, 0, len(g.KPIs))
	seenAnalyticsNames := make(map[string]string) // analyticsName -> original rawName

	for _, kpi := range g.KPIs {
		analyticsName := toAnalyticsName(kpi)

		// Check for duplicate analyticsName
		if existingRawName, exists := seenAnalyticsNames[analyticsName]; exists {
			fmt.Printf("  ⚠️  DUPLICATE analyticsName '%s' in %s:\n", analyticsName, g.ObjectType)
			fmt.Printf("      - Already added from rawName: %s\n", existingRawName)
			fmt.Printf("      - Skipping duplicate rawName: %s\n", kpi)
			continue // Skip duplicate
		}

		seenAnalyticsNames[analyticsName] = kpi
		metrics = append(metrics, DictionaryMetric{
			AnalyticsName: analyticsName,
			DataType:      "double",
			Directions:    []string{"-1"},
			RawName:       kpi,
			Unit:          "value",
		})
	}

	return Dictionary{
		ID:             dictID,
		CustomMetrics:  nil,
		DictionaryName: dictID,
		DictionaryType: "global",
		Dimensions:     standardDimensions,
		MetricType:     "timeseries",
		Metrics:        metrics,
		ObjectType:     g.ObjectType,
		TenantID:       "",
		Vendor:         "cisco",
		ID2:            dictID,
		Type:           "ingestionDictionaries",
	}
}

// toAnalyticsName converts a KPI name to camelCase analytics name
// Example: "MME_DCNR_Attach_Accept_Denied" -> "dcnrAttachAcceptDenied"
func toAnalyticsName(kpi string) string {
	// Split by underscore
	parts := strings.Split(kpi, "_")
	if len(parts) == 0 {
		return strings.ToLower(kpi)
	}

	// Remove common prefixes like MME, SGW, etc.
	prefixes := map[string]bool{
		"MME": true, "SGW": true, "PGW": true, "HSS": true,
		"SAEGW": true, "APN": true, "EGTPC": true, "SX": true,
		"TAI": true, "SBC": true, "SGS": true, "DCCA": true,
		"AMF": true, "SMF": true, "UPF": true,
	}

	// Skip prefix if it's a known schema prefix
	startIdx := 0
	if len(parts) > 0 && prefixes[strings.ToUpper(parts[0])] {
		startIdx = 1
	}

	if startIdx >= len(parts) {
		return strings.ToLower(kpi)
	}

	// Build camelCase
	var result strings.Builder
	for i := startIdx; i < len(parts); i++ {
		part := strings.ToLower(parts[i])
		if i == startIdx {
			result.WriteString(part)
		} else {
			result.WriteString(capitalize(part))
		}
	}

	return result.String()
}

// capitalize returns the string with the first letter capitalized
func capitalize(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
