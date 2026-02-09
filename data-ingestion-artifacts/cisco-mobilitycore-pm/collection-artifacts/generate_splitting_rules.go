// generate_splitting_rules.go - Tool to generate Telegraf KPI splitting rules from CSV catalog
package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

const MaxKPIsPerObjectType = 40

type KPIEntry struct {
	KPIName          string `json:"kpi_name"`
	Schema           string `json:"schema"`
	ObjectIdentifier string `json:"object_identifier"`
}

type SchemaGroup struct {
	Schema     string   `json:"schema"`
	Suffix     string   `json:"suffix"`
	ObjectType string   `json:"object_type"`
	KPIs       []string `json:"kpis"`
	Count      int      `json:"count"`
}

type SplittingRules struct {
	Schemas map[string][]SchemaGroup `json:"schemas"`
}

func main() {
	csvFile := flag.String("csv", "Kpi_calc_Kpicatalog-updatedGrouping.csv", "Path to KPI catalog CSV")
	format := flag.String("format", "summary", "Output format: summary, starlark, routing, json, validate")
	schema := flag.String("schema", "", "Filter by specific schema")
	flag.Parse()

	entries, err := parseCSV(*csvFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	rules := buildRules(entries)

	switch *format {
	case "summary":
		printSummary(rules, *schema)
	case "starlark":
		printStarlark(rules, *schema)
	case "routing":
		printRouting(rules, *schema)
	case "json":
		printJSON(rules, *schema)
	case "validate":
		validate(rules)
	}
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
			entries = append(entries, KPIEntry{kpi, strings.ToLower(schema), objID})
		}
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

func printSummary(rules SplittingRules, filterSchema string) {
	fmt.Println("=" + strings.Repeat("=", 70))
	fmt.Println("KPI SPLITTING RULES SUMMARY (from CSV catalog)")
	fmt.Println("=" + strings.Repeat("=", 70))

	schemas := sortedSchemas(rules, filterSchema)
	for _, schema := range schemas {
		groups := rules.Schemas[schema]
		total := 0
		for _, g := range groups {
			total += g.Count
		}
		status := "✓ OK"
		if len(groups) > 1 {
			status = fmt.Sprintf("⚠️ SPLIT INTO %d GROUPS", len(groups))
		}
		fmt.Printf("\n%s (%d KPIs) - %s\n", strings.ToUpper(schema), total, status)
		fmt.Println(strings.Repeat("-", 50))
		for _, g := range groups {
			ind := "✓"
			if g.Count > 40 {
				ind = "❌"
			} else if g.Count > 35 {
				ind = "⚡"
			}
			suf := "(base)"
			if g.Suffix != "" {
				suf = g.Suffix
			}
			fmt.Printf("  %s %-15s %3d KPIs → %s\n", ind, suf, g.Count, g.ObjectType)
		}
	}
}

func printStarlark(rules SplittingRules, filterSchema string) {
	fmt.Println("# AUTO-GENERATED FROM CSV CATALOG")
	fmt.Println("# Source of truth: Kpi_calc_Kpicatalog-updatedGrouping.csv")
	fmt.Println("# Logic: Use column 3 (ObjectIdentifier) if defined, otherwise column 2 (Schema)")
	fmt.Println("#")
	fmt.Println("# EXACT_RULES maps each KPI name to its target objectType suffix.")
	fmt.Println("# KPIs not in this dictionary use the base objectType (schema only).")
	fmt.Println("")
	fmt.Println("EXACT_RULES = {")

	schemas := sortedSchemas(rules, filterSchema)
	for _, schema := range schemas {
		groups := rules.Schemas[schema]
		if len(groups) <= 1 {
			continue
		}

		total := 0
		for _, g := range groups {
			total += g.Count
		}

		fmt.Printf("\n    # %s (%d KPIs → %d objectTypes)\n", strings.ToUpper(schema), total, len(groups))

		for _, g := range groups {
			if g.Suffix == "" {
				fmt.Printf("    # Base %s: %d KPIs (not in dict = falls through to base)\n", schema, g.Count)
				continue
			}
			fmt.Printf("    # %s%s: %d KPIs\n", schema, g.Suffix, g.Count)
			for _, kpi := range g.KPIs {
				fmt.Printf("    \"%s\": \"%s\",\n", kpi, g.Suffix)
			}
		}
	}
	fmt.Println("}")
}

func printRouting(rules SplittingRules, filterSchema string) {
	fmt.Println("# AUTO-GENERATED ROUTING RULES FROM CSV CATALOG (EXACT MATCHES)")
	fmt.Println("# Source of truth: Kpi_calc_Kpicatalog-updatedGrouping.csv")
	fmt.Println("# Logic: Use column 3 (ObjectIdentifier) if defined, otherwise column 2 (Schema)")
	fmt.Println("#")
	fmt.Println("# Each group has 3 template blocks: objectType, sessionName, sessionId")
	fmt.Println("# Uses exact KPI names in namepass (no wildcards)")

	schemas := sortedSchemas(rules, filterSchema)
	for _, schema := range schemas {
		groups := rules.Schemas[schema]
		if len(groups) <= 1 {
			continue
		}

		// Collect all non-base KPIs for namedrop in base section
		var allNonBaseKPIs []string
		for _, g := range groups {
			if g.Suffix != "" {
				allNonBaseKPIs = append(allNonBaseKPIs, g.KPIs...)
			}
		}

		groupNum := 0
		for _, g := range groups {
			if g.Suffix == "" {
				continue
			}
			groupNum++

			patterns := make([]string, len(g.KPIs))
			for i, kpi := range g.KPIs {
				patterns[i] = fmt.Sprintf("\"%s\"", kpi)
			}
			namepass := strings.Join(patterns, ", ")

			fmt.Printf("\n#######################\n")
			fmt.Printf("# GROUP %d: %s%s (%d KPIs)\n", groupNum, strings.ToUpper(schema), g.Suffix, g.Count)
			fmt.Printf("#######################\n")

			// objectType template
			fmt.Printf("\n[[processors.template]]\n")
			fmt.Printf("  order = 7\n")
			fmt.Printf("  namedrop = [\"internal_*\"]\n")
			fmt.Printf("  namepass = [%s]\n", namepass)
			fmt.Printf("  tag = \"objectType\"\n")
			fmt.Printf("  template = 'cisco-mobilitycore-pm-{{ .Tag \"schema\" }}%s'\n", g.Suffix)
			fmt.Printf("  [processors.template.tagpass]\n")
			fmt.Printf("  schema = [\"%s\"]\n", schema)

			// sessionName template
			fmt.Printf("\n[[processors.template]]\n")
			fmt.Printf("  order = 7\n")
			fmt.Printf("  namedrop = [\"internal_*\"]\n")
			fmt.Printf("  namepass = [%s]\n", namepass)
			fmt.Printf("  tag = \"sessionName\"\n")
			fmt.Printf("  template = '{{ .Tag \"device\" }}_{{ .Tag \"index\" }}_cisco-mobilitycore-pm-{{ .Tag \"schema\" }}%s'\n", g.Suffix)
			fmt.Printf("  [processors.template.tagpass]\n")
			fmt.Printf("  schema = [\"%s\"]\n", schema)

			// sessionId template
			fmt.Printf("\n[[processors.template]]\n")
			fmt.Printf("  order = 7\n")
			fmt.Printf("  namedrop = [\"internal_*\"]\n")
			fmt.Printf("  namepass = [%s]\n", namepass)
			fmt.Printf("  tag = \"sessionId\"\n")
			fmt.Printf("  template = '{{ .Tag \"node_id\" }}_{{ .Tag \"index\" }}_cisco-mobilitycore-pm-{{ .Tag \"schema\" }}%s'\n", g.Suffix)
			fmt.Printf("  [processors.template.tagpass]\n")
			fmt.Printf("  schema = [\"%s\"]\n", schema)
		}

		// Generate namedrop list for base section
		if len(allNonBaseKPIs) > 0 {
			namedrop := make([]string, len(allNonBaseKPIs)+1)
			namedrop[0] = "\"internal_*\""
			for i, kpi := range allNonBaseKPIs {
				namedrop[i+1] = fmt.Sprintf("\"%s\"", kpi)
			}

			// Find base group count
			baseCount := 0
			for _, g := range groups {
				if g.Suffix == "" {
					baseCount = g.Count
					break
				}
			}

			fmt.Printf("\n#######################\n")
			fmt.Printf("# DEFAULT: Base %s (%d KPIs) - excludes all split KPIs\n", strings.ToUpper(schema), baseCount)
			fmt.Printf("#######################\n")

			fmt.Printf("\n[[processors.template]]\n")
			fmt.Printf("  order = 8\n")
			fmt.Printf("  namedrop = [%s]\n", strings.Join(namedrop, ", "))
			fmt.Printf("  tag = \"objectType\"\n")
			fmt.Printf("  template = 'cisco-mobilitycore-pm-{{ .Tag \"schema\" }}'\n")
			fmt.Printf("  [processors.template.tagpass]\n")
			fmt.Printf("  schema = [\"%s\"]\n", schema)

			fmt.Printf("\n[[processors.template]]\n")
			fmt.Printf("  order = 8\n")
			fmt.Printf("  namedrop = [%s]\n", strings.Join(namedrop, ", "))
			fmt.Printf("  tag = \"sessionName\"\n")
			fmt.Printf("  template = '{{ .Tag \"device\" }}_{{ .Tag \"index\" }}_cisco-mobilitycore-pm-{{ .Tag \"schema\" }}'\n")
			fmt.Printf("  [processors.template.tagpass]\n")
			fmt.Printf("  schema = [\"%s\"]\n", schema)

			fmt.Printf("\n[[processors.template]]\n")
			fmt.Printf("  order = 8\n")
			fmt.Printf("  namedrop = [%s]\n", strings.Join(namedrop, ", "))
			fmt.Printf("  tag = \"sessionId\"\n")
			fmt.Printf("  template = '{{ .Tag \"node_id\" }}_{{ .Tag \"index\" }}_cisco-mobilitycore-pm-{{ .Tag \"schema\" }}'\n")
			fmt.Printf("  [processors.template.tagpass]\n")
			fmt.Printf("  schema = [\"%s\"]\n", schema)
		}
	}
}

func printJSON(rules SplittingRules, filterSchema string) {
	output := rules
	if filterSchema != "" {
		output = SplittingRules{Schemas: map[string][]SchemaGroup{filterSchema: rules.Schemas[filterSchema]}}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(output)
}

func validate(rules SplittingRules) {
	fmt.Println("VALIDATION REPORT")
	fmt.Println("=================")
	hasErr := false
	for _, groups := range rules.Schemas {
		for _, g := range groups {
			if g.Count > MaxKPIsPerObjectType {
				fmt.Printf("❌ %s: %d KPIs (exceeds %d)\n", g.ObjectType, g.Count, MaxKPIsPerObjectType)
				hasErr = true
			}
		}
	}
	if !hasErr {
		fmt.Println("✓ All objectTypes within 40 KPI limit")
	}
}

func sortedSchemas(rules SplittingRules, filter string) []string {
	var s []string
	for k := range rules.Schemas {
		if filter == "" || strings.EqualFold(k, filter) {
			s = append(s, k)
		}
	}
	sort.Strings(s)
	return s
}
