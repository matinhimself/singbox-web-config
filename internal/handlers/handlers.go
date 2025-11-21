package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/matinhimself/singbox-web-config/internal/types"
)

// PageData represents common data for all pages
type PageData struct {
	Title string
	Data  interface{}
}

// handleIndex handles the home page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := PageData{
		Title: "Sing-Box Config Manager",
		Data: map[string]interface{}{
			"Metadata": types.Metadata,
		},
	}

	if err := s.renderTemplate(w, "index.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleRulesPage handles the rules management page
func (s *Server) handleRulesPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Route Rules",
		Data: map[string]interface{}{
			"RuleTypes": types.AvailableRuleTypes(),
		},
	}

	if err := s.renderTemplate(w, "rules.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleServicePage handles the service management page
func (s *Server) handleServicePage(w http.ResponseWriter, r *http.Request) {
	status, err := s.serviceManager.GetStatus()
	if err != nil {
		log.Printf("Error getting service status: %v", err)
	}

	data := PageData{
		Title: "Service Management",
		Data: map[string]interface{}{
			"Status": status,
		},
	}

	if err := s.renderTemplate(w, "service.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleRulesList handles the HTMX endpoint for rules list
func (s *Server) handleRulesList(w http.ResponseWriter, r *http.Request) {
	rules, err := s.configManager.GetRules()
	if err != nil {
		log.Printf("Error getting rules: %v", err)
		http.Error(w, "Failed to load rules", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Rules": rules,
	}

	if err := s.renderTemplate(w, "rule-list.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleRuleForm handles the HTMX endpoint for rule forms
func (s *Server) handleRuleForm(w http.ResponseWriter, r *http.Request) {
	ruleType := r.URL.Query().Get("type")
	indexStr := r.URL.Query().Get("index")
	editMode := indexStr != ""

	var ruleData map[string]interface{}
	var ruleIndex int

	if editMode {
		// Get existing rule for editing
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}
		ruleIndex = index

		rules, err := s.configManager.GetRules()
		if err != nil {
			log.Printf("Error getting rules: %v", err)
			http.Error(w, "Failed to get rules", http.StatusInternalServerError)
			return
		}

		if index < 0 || index >= len(rules) {
			http.Error(w, "Index out of range", http.StatusBadRequest)
			return
		}

		// Convert rule to map
		if rule, ok := rules[index].(map[string]interface{}); ok {
			ruleData = rule
			// Determine rule type from the rule data if not specified
			if ruleType == "" {
				ruleType = s.determineRuleType(rule)
			}
		} else {
			http.Error(w, "Invalid rule format", http.StatusInternalServerError)
			return
		}
	}

	if ruleType == "" {
		ruleType = "RawDefaultRule" // Default type
	}

	formDef, err := s.formBuilder.BuildForm(ruleType)
	if err != nil {
		log.Printf("Error building form: %v", err)
		http.Error(w, "Failed to build form", http.StatusInternalServerError)
		return
	}

	// Populate form with existing values if editing
	if editMode && ruleData != nil {
		s.formBuilder.PopulateFormValues(formDef, ruleData)
	}

	// Get outbounds for dropdown
	outbounds, err := s.getOutboundTags()
	if err != nil {
		log.Printf("Warning: failed to get outbounds: %v", err)
	}

	// Update Outbound field to be a select with outbound options
	for i := range formDef.Fields {
		if formDef.Fields[i].JSONTag == "outbound" {
			// If it's an array field (for DNS rules), keep it as array but still show options
			if formDef.Fields[i].Type != "array" {
				formDef.Fields[i].Type = "select"
			}
			formDef.Fields[i].Options = outbounds
			break
		}
	}

	data := map[string]interface{}{
		"Form":      formDef,
		"RuleTypes": s.formBuilder.GetAvailableRuleTypes(),
		"EditMode":  editMode,
		"RuleIndex": ruleIndex,
	}

	if err := s.renderTemplate(w, "rule-form.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// determineRuleType tries to determine the rule type from rule data
func (s *Server) determineRuleType(rule map[string]interface{}) string {
	// Check for logical rule
	if _, hasMode := rule["mode"]; hasMode {
		if _, hasRules := rule["rules"]; hasRules {
			// Check if it's DNS or routing
			if _, hasServer := rule["server"]; hasServer {
				return "RawLogicalDNSRule"
			}
			return "RawLogicalRule"
		}
	}

	// Check for rule set
	if _, hasType := rule["type"]; hasType {
		if ruleType, ok := rule["type"].(string); ok {
			if ruleType == "local" {
				return "LocalRuleSet"
			}
			if ruleType == "remote" {
				return "RemoteRuleSet"
			}
		}
	}

	// Check if it's a DNS rule
	if _, hasServer := rule["server"]; hasServer {
		return "RawDefaultDNSRule"
	}

	// Default to regular rule
	return "RawDefaultRule"
}

// getOutboundTags retrieves all outbound tags from the config
func (s *Server) getOutboundTags() ([]string, error) {
	config, err := s.configManager.LoadConfig()
	if err != nil {
		return nil, err
	}

	var tags []string
	for _, outbound := range config.Outbounds {
		if outboundMap, ok := outbound.(map[string]interface{}); ok {
			if tag, ok := outboundMap["tag"].(string); ok {
				tags = append(tags, tag)
			}
		}
	}

	// Add some common default outbounds
	if len(tags) == 0 {
		tags = []string{"direct", "block", "dns-out"}
	}

	return tags, nil
}

// handleRuleCreate handles creating a new rule
func (s *Server) handleRuleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Build rule from form data
	rule := s.buildRuleFromForm(r)

	// Get current rules
	rules, err := s.configManager.GetRules()
	if err != nil {
		log.Printf("Error getting rules: %v", err)
		http.Error(w, "Failed to get rules", http.StatusInternalServerError)
		return
	}

	// Add new rule
	rules = append(rules, rule)

	// Update config
	if err := s.configManager.UpdateRules(rules); err != nil {
		log.Printf("Error updating rules: %v", err)
		http.Error(w, "Failed to save rules", http.StatusInternalServerError)
		return
	}

	// Reload service to apply changes
	if err := s.serviceManager.Reload(); err != nil {
		log.Printf("Warning: failed to reload service: %v", err)
	}

	// Return updated rules list
	s.handleRulesList(w, r)
}

// handleRuleDelete handles deleting a rule
func (s *Server) handleRuleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	indexStr := r.URL.Query().Get("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	// Get current rules
	rules, err := s.configManager.GetRules()
	if err != nil {
		log.Printf("Error getting rules: %v", err)
		http.Error(w, "Failed to get rules", http.StatusInternalServerError)
		return
	}

	// Check bounds
	if index < 0 || index >= len(rules) {
		http.Error(w, "Index out of range", http.StatusBadRequest)
		return
	}

	// Remove rule
	rules = append(rules[:index], rules[index+1:]...)

	// Update config
	if err := s.configManager.UpdateRules(rules); err != nil {
		log.Printf("Error updating rules: %v", err)
		http.Error(w, "Failed to save rules", http.StatusInternalServerError)
		return
	}

	// Reload service
	if err := s.serviceManager.Reload(); err != nil {
		log.Printf("Warning: failed to reload service: %v", err)
	}

	// Return updated rules list
	s.handleRulesList(w, r)
}

// handleRuleUpdate handles updating a rule
func (s *Server) handleRuleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	indexStr := r.FormValue("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	// Build rule from form data
	rule := s.buildRuleFromForm(r)

	// Get current rules
	rules, err := s.configManager.GetRules()
	if err != nil {
		log.Printf("Error getting rules: %v", err)
		http.Error(w, "Failed to get rules", http.StatusInternalServerError)
		return
	}

	// Check bounds
	if index < 0 || index >= len(rules) {
		http.Error(w, "Index out of range", http.StatusBadRequest)
		return
	}

	// Update rule
	rules[index] = rule

	// Update config
	if err := s.configManager.UpdateRules(rules); err != nil {
		log.Printf("Error updating rules: %v", err)
		http.Error(w, "Failed to save rules", http.StatusInternalServerError)
		return
	}

	// Reload service
	if err := s.serviceManager.Reload(); err != nil {
		log.Printf("Warning: failed to reload service: %v", err)
	}

	// Return updated rules list
	s.handleRulesList(w, r)
}

// handleRuleReorder handles reordering rules via drag and drop
func (s *Server) handleRuleReorder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	fromStr := r.FormValue("from")
	toStr := r.FormValue("to")

	fromIndex, err := strconv.Atoi(fromStr)
	if err != nil {
		http.Error(w, "Invalid from index", http.StatusBadRequest)
		return
	}

	toIndex, err := strconv.Atoi(toStr)
	if err != nil {
		http.Error(w, "Invalid to index", http.StatusBadRequest)
		return
	}

	// Get current rules
	rules, err := s.configManager.GetRules()
	if err != nil {
		log.Printf("Error getting rules: %v", err)
		http.Error(w, "Failed to get rules", http.StatusInternalServerError)
		return
	}

	// Check bounds
	if fromIndex < 0 || fromIndex >= len(rules) || toIndex < 0 || toIndex >= len(rules) {
		http.Error(w, "Index out of range", http.StatusBadRequest)
		return
	}

	// Reorder rules
	rule := rules[fromIndex]
	rules = append(rules[:fromIndex], rules[fromIndex+1:]...)

	// Insert at new position
	if toIndex > fromIndex {
		toIndex--
	}

	newRules := make([]interface{}, 0, len(rules)+1)
	newRules = append(newRules, rules[:toIndex]...)
	newRules = append(newRules, rule)
	newRules = append(newRules, rules[toIndex:]...)

	// Update config
	if err := s.configManager.UpdateRules(newRules); err != nil {
		log.Printf("Error updating rules: %v", err)
		http.Error(w, "Failed to save rules", http.StatusInternalServerError)
		return
	}

	// Reload service
	if err := s.serviceManager.Reload(); err != nil {
		log.Printf("Warning: failed to reload service: %v", err)
	}

	// Return updated rules list
	s.handleRulesList(w, r)
}

// buildRuleFromForm builds a rule map from form data
func (s *Server) buildRuleFromForm(r *http.Request) map[string]interface{} {
	rule := make(map[string]interface{})

	for key, values := range r.Form {
		if key == "index" || key == "rule_type" {
			continue
		}

		// Skip empty values
		if len(values) == 0 || (len(values) == 1 && values[0] == "") {
			continue
		}

		// Handle array fields
		// Check if the field looks like it should be an array
		if strings.HasSuffix(key, "[]") || len(values) > 1 {
			// Remove [] suffix if present
			fieldName := strings.TrimSuffix(key, "[]")

			// Parse comma-separated values
			var allValues []string
			for _, v := range values {
				if v != "" {
					// Split by comma for multiple values in one field
					parts := strings.Split(v, ",")
					for _, part := range parts {
						trimmed := strings.TrimSpace(part)
						if trimmed != "" {
							allValues = append(allValues, trimmed)
						}
					}
				}
			}

			if len(allValues) > 0 {
				rule[fieldName] = allValues
			}
		} else {
			// Single value field
			value := values[0]
			if value != "" {
				rule[key] = value
			}
		}
	}

	return rule
}

// Service management handlers

func (s *Server) handleServiceStatus(w http.ResponseWriter, r *http.Request) {
	status, err := s.serviceManager.GetStatus()
	if err != nil {
		log.Printf("Error getting service status: %v", err)
		http.Error(w, "Failed to get service status", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Status": status,
	}

	if err := s.renderTemplate(w, "service-status.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) handleServiceStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.serviceManager.Start(); err != nil {
		log.Printf("Error starting service: %v", err)
		http.Error(w, fmt.Sprintf("Failed to start service: %v", err), http.StatusInternalServerError)
		return
	}

	s.handleServiceStatus(w, r)
}

func (s *Server) handleServiceStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.serviceManager.Stop(); err != nil {
		log.Printf("Error stopping service: %v", err)
		http.Error(w, fmt.Sprintf("Failed to stop service: %v", err), http.StatusInternalServerError)
		return
	}

	s.handleServiceStatus(w, r)
}

func (s *Server) handleServiceRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.serviceManager.Restart(); err != nil {
		log.Printf("Error restarting service: %v", err)
		http.Error(w, fmt.Sprintf("Failed to restart service: %v", err), http.StatusInternalServerError)
		return
	}

	s.handleServiceStatus(w, r)
}

func (s *Server) handleServiceLogs(w http.ResponseWriter, r *http.Request) {
	lines := 100
	if linesStr := r.URL.Query().Get("lines"); linesStr != "" {
		if l, err := strconv.Atoi(linesStr); err == nil {
			lines = l
		}
	}

	logs, err := s.serviceManager.GetLogs(lines)
	if err != nil {
		log.Printf("Error getting service logs: %v", err)
		http.Error(w, "Failed to get service logs", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Logs": logs,
	}

	if err := s.renderTemplate(w, "service-logs.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Config management handlers

func (s *Server) handleConfigExport(w http.ResponseWriter, r *http.Request) {
	config, err := s.configManager.LoadConfig()
	if err != nil {
		log.Printf("Error loading config: %v", err)
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=sing-box-config.json")

	if err := json.NewEncoder(w).Encode(config); err != nil {
		log.Printf("Error encoding config: %v", err)
		http.Error(w, "Failed to export config", http.StatusInternalServerError)
	}
}

func (s *Server) handleConfigBackups(w http.ResponseWriter, r *http.Request) {
	backups, err := s.configManager.ListBackups()
	if err != nil {
		log.Printf("Error listing backups: %v", err)
		http.Error(w, "Failed to list backups", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Backups": backups,
	}

	if err := s.renderTemplate(w, "config-backups.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) handleConfigRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	backupName := r.FormValue("backup")
	if backupName == "" {
		http.Error(w, "No backup specified", http.StatusBadRequest)
		return
	}

	if err := s.configManager.RestoreBackup(backupName); err != nil {
		log.Printf("Error restoring backup: %v", err)
		http.Error(w, fmt.Sprintf("Failed to restore backup: %v", err), http.StatusInternalServerError)
		return
	}

	// Reload service
	if err := s.serviceManager.Reload(); err != nil {
		log.Printf("Warning: failed to reload service: %v", err)
	}

	w.Header().Set("HX-Redirect", "/rules")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleConfigCreateBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		name = fmt.Sprintf("Manual backup %s", time.Now().Format("2006-01-02 15:04:05"))
	}

	description := r.FormValue("description")
	if description == "" {
		description = "Manual backup created by user"
	}

	if err := s.configManager.CreateBackupWithName(name, description); err != nil {
		log.Printf("Error creating backup: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create backup: %v", err), http.StatusInternalServerError)
		return
	}

	// Return updated backup list
	w.Header().Set("HX-Trigger", "backupCreated")
	s.handleConfigBackups(w, r)
}
