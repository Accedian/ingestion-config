// Dictionary Uploader Tool
//
// Uploads ingestion dictionaries prepared by generate_telegraf_configs.go to cloud.
//
// Environment Variables:
//
//	BASE_URL             - Base URL of the PCA API (default: https://pca.kajar.npav.accedian.net)
//	AUTHORIZATION_HEADER - Authorization header value (default: Bearer XXXXXXX)
//	DICTIONARIES_PATH    - Path to dictionaries directory (default: ./generated_dictionaries)
//	INSECURE_SKIP_VERIFY - Skip TLS certificate verification (default: false)
//
// Usage:
//
//	export BASE_URL="https://pca.kajar.npav.accedian.net"
//	export AUTHORIZATION_HEADER="Bearer your-token-here"
//	export DICTIONARIES_PATH="./generated_dictionaries"
//	export INSECURE_SKIP_VERIFY="true"  # for self-signed certs
//	go run main.go
package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
)

// Config holds configuration for the uploader
type Config struct {
	BaseURL        string
	AuthHeader     string
	DictPath       string
	InsecureVerify bool
	Debug          bool
}

// Dictionary represents an ingestion dictionary
type Dictionary struct {
	ID             string      `json:"_id"`
	Rev            string      `json:"_rev,omitempty"`
	CustomMetrics  interface{} `json:"customMetrics"`
	DictionaryName string      `json:"dictionaryName"`
	DictionaryType string      `json:"dictionaryType"`
	Dimensions     []Dimension `json:"dimensions"`
	MetricType     string      `json:"metricType"`
	Metrics        []Metric    `json:"metrics"`
	ObjectType     string      `json:"objectType"`
	TenantID       string      `json:"tenantId"`
	Vendor         string      `json:"vendor"`
	ID2            string      `json:"id,omitempty"`
	Type           string      `json:"type"`
}

// Dimension represents a dimension in the dictionary
type Dimension struct {
	AnalyticsName string `json:"analyticsName"`
	DataType      string `json:"dataType"`
	RawName       string `json:"rawName"`
}

// Metric represents a metric in the dictionary
type Metric struct {
	AnalyticsName string   `json:"analyticsName"`
	DataType      string   `json:"dataType"`
	Directions    []string `json:"directions"`
	RawName       string   `json:"rawName"`
	Unit          string   `json:"unit"`
}

// UploadResult tracks the result of each dictionary upload
type UploadResult struct {
	DictionaryID string
	Action       string // "created", "updated", "unchanged", "error"
	Message      string
}

func main() {
	cfg := loadConfig()

	fmt.Println("=" + strings.Repeat("=", 70))
	fmt.Println("DICTIONARY UPLOADER")
	fmt.Println("=" + strings.Repeat("=", 70))
	fmt.Printf("Base URL: %s\n", cfg.BaseURL)
	fmt.Printf("Dictionaries Path: %s\n", cfg.DictPath)
	fmt.Println()

	// Validate authorization header
	if cfg.AuthHeader == "Bearer XXXXXXX" || cfg.AuthHeader == "" {
		fmt.Println("WARNING: Using default/empty authorization header")
		fmt.Println("   Set AUTHORIZATION_HEADER environment variable with valid token")
		fmt.Println()
	}

	// Load local dictionaries
	localDicts, err := loadLocalDictionaries(cfg.DictPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading local dictionaries: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Loaded %d local dictionaries\n", len(localDicts))

	// Create HTTP client with timeout
	transport := &http.Transport{}
	if cfg.InsecureVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		fmt.Println("WARNING: TLS certificate verification disabled")
		fmt.Println()
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	// Process each dictionary
	fmt.Println("\n--- PROCESSING DICTIONARIES ---")
	results := processDictionaries(client, cfg, localDicts)

	// Print summary
	printSummary(results)
}

func loadConfig() Config {
	insecure := getEnvOrDefault("INSECURE_SKIP_VERIFY", "false")
	debug := getEnvOrDefault("DEBUG", "false")
	return Config{
		BaseURL:        getEnvOrDefault("BASE_URL", "https://pca.kajar.npav.accedian.net"),
		AuthHeader:     getEnvOrDefault("AUTHORIZATION_HEADER", "Bearer XXXXXXX"),
		DictPath:       getEnvOrDefault("DICTIONARIES_PATH", "./generated_dictionaries"),
		InsecureVerify: insecure == "true" || insecure == "1",
		Debug:          debug == "true" || debug == "1",
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func loadLocalDictionaries(path string) (map[string]Dictionary, error) {
	dicts := make(map[string]Dictionary)

	files, err := filepath.Glob(filepath.Join(path, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list dictionary files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no dictionary files found in %s", path)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", file, err)
		}

		var dict Dictionary
		if err := json.Unmarshal(data, &dict); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", file, err)
		}

		if dict.ID == "" {
			return nil, fmt.Errorf("dictionary %s has no _id field", file)
		}

		dicts[dict.ID] = dict
	}

	return dicts, nil
}

func processDictionaries(client *http.Client, cfg Config, localDicts map[string]Dictionary) []UploadResult {
	var results []UploadResult

	// Sort dictionary IDs for consistent output
	var ids []string
	for id := range localDicts {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		localDict := localDicts[id]
		result := processOneDictionary(client, cfg, localDict)
		results = append(results, result)

		switch result.Action {
		case "created":
			fmt.Printf("  [CREATED]   %s\n", id)
		case "updated":
			fmt.Printf("  [UPDATED]   %s\n", id)
		case "unchanged":
			fmt.Printf("  [UNCHANGED] %s\n", id)
		case "error":
			fmt.Printf("  [ERROR]     %s: %s\n", id, result.Message)
		}
	}

	return results
}

func processOneDictionary(client *http.Client, cfg Config, localDict Dictionary) UploadResult {
	result := UploadResult{DictionaryID: localDict.ID}

	// Fetch existing dictionary from server
	remoteDict, exists, err := fetchRemoteDictionary(client, cfg, localDict.ID)
	if err != nil {
		result.Action = "error"
		result.Message = fmt.Sprintf("failed to fetch: %v", err)
		return result
	}

	if !exists {
		// Dictionary doesn't exist - create it
		if err := createDictionary(client, cfg, localDict); err != nil {
			result.Action = "error"
			result.Message = fmt.Sprintf("failed to create: %v", err)
			return result
		}
		result.Action = "created"
		result.Message = "dictionary created"
		return result
	}

	// Compare dictionaries (ignoring CouchDB fields)
	if dictionariesEqual(localDict, remoteDict) {
		result.Action = "unchanged"
		result.Message = "no changes detected"
		return result
	}

	// Dictionary differs - update it with the _rev from remote
	localDict.Rev = remoteDict.Rev
	if err := updateDictionary(client, cfg, localDict); err != nil {
		result.Action = "error"
		result.Message = fmt.Sprintf("failed to update: %v", err)
		return result
	}
	result.Action = "updated"
	result.Message = "dictionary updated"
	return result
}

func fetchRemoteDictionary(client *http.Client, cfg Config, id string) (Dictionary, bool, error) {
	url := fmt.Sprintf("%s/api/v3/ingestion-dictionaries/%s", cfg.BaseURL, id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Dictionary{}, false, err
	}
	req.Header.Set("Authorization", cfg.AuthHeader)
	req.Header.Set("Accept", "application/vnd.api+json")

	resp, err := client.Do(req)
	if err != nil {
		return Dictionary{}, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Dictionary{}, false, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return Dictionary{}, false, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON:API response format
	var jsonAPIResp JSONAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&jsonAPIResp); err != nil {
		return Dictionary{}, false, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert from JSON:API format to our Dictionary struct
	dict := Dictionary{
		ID:             jsonAPIResp.Data.ID,
		Rev:            jsonAPIResp.Data.Attributes.Rev,
		DictionaryName: jsonAPIResp.Data.Attributes.DictionaryName,
		DictionaryType: jsonAPIResp.Data.Attributes.DictionaryType,
		CustomMetrics:  jsonAPIResp.Data.Attributes.CustomMetrics,
		Dimensions:     jsonAPIResp.Data.Attributes.Dimensions,
		MetricType:     jsonAPIResp.Data.Attributes.MetricType,
		Metrics:        jsonAPIResp.Data.Attributes.Metrics,
		ObjectType:     jsonAPIResp.Data.Attributes.ObjectType,
		TenantID:       jsonAPIResp.Data.Attributes.TenantID,
		Vendor:         jsonAPIResp.Data.Attributes.Vendor,
		Type:           jsonAPIResp.Data.Type,
	}

	return dict, true, nil
}

// JSONAPIRequest represents the JSON:API format required by the PCA API
type JSONAPIRequest struct {
	Data JSONAPIData `json:"data"`
}

// JSONAPIResponse represents the JSON:API format returned by GET requests
type JSONAPIResponse struct {
	Data JSONAPIData `json:"data"`
}

// JSONAPIListResponse represents the JSON:API format returned by list requests
type JSONAPIListResponse struct {
	Data []JSONAPIData `json:"data"`
}

// JSONAPIData represents the data object in JSON:API format
type JSONAPIData struct {
	ID         string               `json:"id"`
	Type       string               `json:"type"`
	Attributes DictionaryAttributes `json:"attributes"`
}

// DictionaryAttributes contains the actual dictionary content for JSON:API
type DictionaryAttributes struct {
	Rev            string      `json:"_rev,omitempty"`
	DictionaryName string      `json:"dictionaryName"`
	DictionaryType string      `json:"dictionaryType"`
	CustomMetrics  interface{} `json:"customMetrics"`
	Dimensions     []Dimension `json:"dimensions"`
	MetricType     string      `json:"metricType"`
	Metrics        []Metric    `json:"metrics"`
	ObjectType     string      `json:"objectType"`
	TenantID       string      `json:"tenantId"`
	Vendor         string      `json:"vendor"`
}

// toJSONAPIRequest converts a Dictionary to JSON:API format for creation
func (d Dictionary) toJSONAPIRequest() JSONAPIRequest {
	return JSONAPIRequest{
		Data: JSONAPIData{
			ID:   d.ID,
			Type: "ingestionDictionaries",
			Attributes: DictionaryAttributes{
				DictionaryName: d.DictionaryName,
				DictionaryType: d.DictionaryType,
				CustomMetrics:  d.CustomMetrics,
				Dimensions:     d.Dimensions,
				MetricType:     d.MetricType,
				Metrics:        d.Metrics,
				ObjectType:     d.ObjectType,
				TenantID:       d.TenantID,
				Vendor:         d.Vendor,
			},
		},
	}
}

// toJSONAPIUpdateRequest converts a Dictionary to JSON:API format for updates (includes _rev)
func (d Dictionary) toJSONAPIUpdateRequest() JSONAPIRequest {
	return JSONAPIRequest{
		Data: JSONAPIData{
			ID:   d.ID,
			Type: "ingestionDictionaries",
			Attributes: DictionaryAttributes{
				Rev:            d.Rev,
				DictionaryName: d.DictionaryName,
				DictionaryType: d.DictionaryType,
				CustomMetrics:  d.CustomMetrics,
				Dimensions:     d.Dimensions,
				MetricType:     d.MetricType,
				Metrics:        d.Metrics,
				ObjectType:     d.ObjectType,
				TenantID:       d.TenantID,
				Vendor:         d.Vendor,
			},
		},
	}
}

func createDictionary(client *http.Client, cfg Config, dict Dictionary) error {
	url := fmt.Sprintf("%s/api/v3/ingestion-dictionaries", cfg.BaseURL)

	// Convert to JSON:API format
	jsonAPIReq := dict.toJSONAPIRequest()
	body, err := json.Marshal(jsonAPIReq)
	if err != nil {
		return fmt.Errorf("failed to marshal dictionary: %w", err)
	}

	if cfg.Debug {
		fmt.Printf("DEBUG CREATE %s:\n%s\n", dict.ID, string(body))
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", cfg.AuthHeader)
	req.Header.Set("Content-Type", "application/vnd.api+json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		respStr := string(respBody)

		// Check for duplicate objectType error and find conflicting dictionary
		if strings.Contains(respStr, "dictionaryDuplicateObjectType") || strings.Contains(respStr, "object type already in use") {
			conflictingID, found := findDictionaryByObjectType(client, cfg, dict.ObjectType)
			if found {
				return fmt.Errorf("HTTP %d: objectType '%s' already in use by dictionary '%s'. Consider updating that dictionary instead or use a different objectType",
					resp.StatusCode, dict.ObjectType, conflictingID)
			}
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, respStr)
	}

	return nil
}

// findDictionaryByObjectType searches all dictionaries to find one with the given objectType
func findDictionaryByObjectType(client *http.Client, cfg Config, objectType string) (string, bool) {
	// Try global dictionaries first
	url := fmt.Sprintf("%s/api/v3/ingestion-dictionaries/global", cfg.BaseURL)
	if id, found := searchDictionariesForObjectType(client, cfg, url, objectType); found {
		return id, true
	}

	// Try tenant dictionaries
	url = fmt.Sprintf("%s/api/v3/ingestion-dictionaries", cfg.BaseURL)
	if id, found := searchDictionariesForObjectType(client, cfg, url, objectType); found {
		return id, true
	}

	return "", false
}

// searchDictionariesForObjectType fetches dictionaries from the given URL and searches for objectType
func searchDictionariesForObjectType(client *http.Client, cfg Config, url, objectType string) (string, bool) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", false
	}
	req.Header.Set("Authorization", cfg.AuthHeader)
	req.Header.Set("Accept", "application/vnd.api+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false
	}

	var listResp JSONAPIListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return "", false
	}

	for _, d := range listResp.Data {
		if d.Attributes.ObjectType == objectType {
			return d.ID, true
		}
	}

	return "", false
}

func updateDictionary(client *http.Client, cfg Config, dict Dictionary) error {
	url := fmt.Sprintf("%s/api/v3/ingestion-dictionaries/%s", cfg.BaseURL, dict.ID)

	// Convert to JSON:API format for update (includes _rev)
	jsonAPIReq := dict.toJSONAPIUpdateRequest()
	body, err := json.Marshal(jsonAPIReq)
	if err != nil {
		return fmt.Errorf("failed to marshal dictionary: %w", err)
	}

	if cfg.Debug {
		fmt.Printf("DEBUG UPDATE %s:\n%s\n", dict.ID, string(body))
	}

	// Use PATCH as per OpenAPI spec
	req, err := http.NewRequest("PATCH", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", cfg.AuthHeader)
	req.Header.Set("Content-Type", "application/vnd.api+json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// dictionariesEqual compares two dictionaries, ignoring CouchDB metadata fields
func dictionariesEqual(local, remote Dictionary) bool {
	if local.ID != remote.ID {
		return false
	}
	if local.DictionaryName != remote.DictionaryName {
		return false
	}
	if local.DictionaryType != remote.DictionaryType {
		return false
	}
	if local.MetricType != remote.MetricType {
		return false
	}
	if local.ObjectType != remote.ObjectType {
		return false
	}
	if local.Vendor != remote.Vendor {
		return false
	}
	if local.Type != remote.Type {
		return false
	}

	if !dimensionsEqual(local.Dimensions, remote.Dimensions) {
		return false
	}

	if !metricsEqual(local.Metrics, remote.Metrics) {
		return false
	}

	return true
}

func dimensionsEqual(local, remote []Dimension) bool {
	if len(local) != len(remote) {
		return false
	}

	localMap := make(map[string]Dimension)
	for _, d := range local {
		localMap[d.RawName] = d
	}

	for _, rd := range remote {
		ld, exists := localMap[rd.RawName]
		if !exists {
			return false
		}
		if !reflect.DeepEqual(ld, rd) {
			return false
		}
	}

	return true
}

func metricsEqual(local, remote []Metric) bool {
	if len(local) != len(remote) {
		return false
	}

	localMap := make(map[string]Metric)
	for _, m := range local {
		localMap[m.RawName] = m
	}

	for _, rm := range remote {
		lm, exists := localMap[rm.RawName]
		if !exists {
			return false
		}
		if !reflect.DeepEqual(lm, rm) {
			return false
		}
	}

	return true
}

func printSummary(results []UploadResult) {
	fmt.Println("\n--- SUMMARY ---")

	created := 0
	updated := 0
	unchanged := 0
	errors := 0

	for _, r := range results {
		switch r.Action {
		case "created":
			created++
		case "updated":
			updated++
		case "unchanged":
			unchanged++
		case "error":
			errors++
		}
	}

	fmt.Printf("  Created:   %d\n", created)
	fmt.Printf("  Updated:   %d\n", updated)
	fmt.Printf("  Unchanged: %d\n", unchanged)
	fmt.Printf("  Errors:    %d\n", errors)
	fmt.Println()

	if errors > 0 {
		fmt.Println("WARNING: Some dictionaries failed to upload. Check the errors above.")
		os.Exit(1)
	} else if created+updated > 0 {
		fmt.Println("All dictionaries synchronized successfully!")
	} else {
		fmt.Println("All dictionaries already up to date!")
	}
}
