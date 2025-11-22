package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/matinhimself/singbox-web-config/internal/config"
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

// Rule Actions Management

type RuleActionData struct {
	Type                      string
	Outbound                  string
	Sniffer                   []string
	Timeout                   uint32
	Server                    string
	Strategy                  string
	DisableCache              bool
	RewriteTTL                *uint32
	ClientSubnet              *string
	Method                    string
	NoDrop                    bool
	OverrideAddress           string
	OverridePort              uint16
	NetworkStrategy           *string
	FallbackDelay             uint32
	UDPDisableDomainUnmapping bool
	UDPConnect                bool
	UDPTimeout                uint32
	TLSFragment               bool
	TLSFragmentFallbackDelay  uint32
	TLSRecordFragment         bool
	Config                    map[string]interface{}
}

// handleRuleActionsPage handles the rule actions management page
func (s *Server) handleRuleActionsPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Rule Actions",
		Data:  map[string]interface{}{},
	}

	if err := s.renderTemplate(w, "rule-actions.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleRuleActionsList returns the list of configured rule actions
func (s *Server) handleRuleActionsList(w http.ResponseWriter, r *http.Request) {
	config, err := s.configManager.LoadConfig()
	if err != nil {
		log.Printf("Error getting config: %v", err)
		http.Error(w, "Failed to get config", http.StatusInternalServerError)
		return
	}

	// Extract rule actions from config
	ruleActions := []RuleActionData{}

	// Get from route.rule_action if it exists
	if config.Route != nil && config.Route.RuleAction != nil {
		for _, action := range config.Route.RuleAction {
			if actionMap, ok := action.(map[string]interface{}); ok {
				ruleAction := s.parseRuleAction(actionMap)
				ruleActions = append(ruleActions, ruleAction)
			}
		}
	}

	data := map[string]interface{}{
		"RuleActions": ruleActions,
	}

	if err := s.renderTemplate(w, "rule-action-list.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// parseRuleAction parses a rule action from config
func (s *Server) parseRuleAction(actionMap map[string]interface{}) RuleActionData {
	action := RuleActionData{
		Config: actionMap,
	}

	// Get action type
	if actionType, ok := actionMap["action"].(string); ok {
		action.Type = actionType
	}

	// Parse based on type
	if outbound, ok := actionMap["outbound"].(string); ok {
		action.Outbound = outbound
	}
	if sniffer, ok := actionMap["sniffer"].([]interface{}); ok {
		for _, s := range sniffer {
			if str, ok := s.(string); ok {
				action.Sniffer = append(action.Sniffer, str)
			}
		}
	}
	if timeout, ok := actionMap["timeout"].(float64); ok {
		action.Timeout = uint32(timeout)
	}
	if server, ok := actionMap["server"].(string); ok {
		action.Server = server
	}
	if strategy, ok := actionMap["strategy"].(string); ok {
		action.Strategy = strategy
	}
	if disableCache, ok := actionMap["disable_cache"].(bool); ok {
		action.DisableCache = disableCache
	}
	if method, ok := actionMap["method"].(string); ok {
		action.Method = method
	}
	if noDrop, ok := actionMap["no_drop"].(bool); ok {
		action.NoDrop = noDrop
	}
	if overrideAddress, ok := actionMap["override_address"].(string); ok {
		action.OverrideAddress = overrideAddress
	}
	if overridePort, ok := actionMap["override_port"].(float64); ok {
		action.OverridePort = uint16(overridePort)
	}
	if udpConnect, ok := actionMap["udp_connect"].(bool); ok {
		action.UDPConnect = udpConnect
	}
	if tlsFragment, ok := actionMap["tls_fragment"].(bool); ok {
		action.TLSFragment = tlsFragment
	}
	if tlsRecordFragment, ok := actionMap["tls_record_fragment"].(bool); ok {
		action.TLSRecordFragment = tlsRecordFragment
	}
	if fallbackDelay, ok := actionMap["fallback_delay"].(float64); ok {
		action.FallbackDelay = uint32(fallbackDelay)
	}
	if udpTimeout, ok := actionMap["udp_timeout"].(float64); ok {
		action.UDPTimeout = uint32(udpTimeout)
	}
	if udpDisableDomainUnmapping, ok := actionMap["udp_disable_domain_unmapping"].(bool); ok {
		action.UDPDisableDomainUnmapping = udpDisableDomainUnmapping
	}
	if tlsFragmentFallbackDelay, ok := actionMap["tls_fragment_fallback_delay"].(float64); ok {
		action.TLSFragmentFallbackDelay = uint32(tlsFragmentFallbackDelay)
	}
	// Handle pointer fields
	if networkStrategy, ok := actionMap["network_strategy"].(string); ok && networkStrategy != "" {
		action.NetworkStrategy = &networkStrategy
	}
	if clientSubnet, ok := actionMap["client_subnet"].(string); ok && clientSubnet != "" {
		action.ClientSubnet = &clientSubnet
	}
	if rewriteTTL, ok := actionMap["rewrite_ttl"].(float64); ok {
		ttl := uint32(rewriteTTL)
		action.RewriteTTL = &ttl
	}

	return action
}

// handleRuleActionForm handles showing the rule action form
func (s *Server) handleRuleActionForm(w http.ResponseWriter, r *http.Request) {
	indexStr := r.URL.Query().Get("index")
	editMode := indexStr != ""

	outbounds, _ := s.getOutboundTags()
	data := map[string]interface{}{
		"EditMode":  editMode,
		"Outbounds": outbounds,
	}

	if editMode {
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}

		config, err := s.configManager.LoadConfig()
		if err != nil {
			http.Error(w, "Failed to get config", http.StatusInternalServerError)
			return
		}

		// Get rule action from config
		if config.Route != nil && config.Route.RuleAction != nil {
			if index >= 0 && index < len(config.Route.RuleAction) {
				if actionMap, ok := config.Route.RuleAction[index].(map[string]interface{}); ok {
					data["Action"] = s.parseRuleAction(actionMap)
					data["ActionIndex"] = index
				}
			}
		}
	} else {
		data["Action"] = RuleActionData{}
	}

	if err := s.renderTemplate(w, "rule-action-form.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleRuleActionCreate handles creating a new rule action
func (s *Server) handleRuleActionCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	action := s.buildRuleActionFromForm(r)

	// Get current config
	cfg, err := s.configManager.LoadConfig()
	if err != nil {
		http.Error(w, "Failed to get config", http.StatusInternalServerError)
		return
	}

	// Ensure route exists
	if cfg.Route == nil {
		cfg.Route = &config.RouteConfig{
			RuleAction: []interface{}{},
		}
	}

	// Ensure rule_action exists
	if cfg.Route.RuleAction == nil {
		cfg.Route.RuleAction = []interface{}{}
	}

	// Add new action
	cfg.Route.RuleAction = append(cfg.Route.RuleAction, action)

	// Save config
	if err := s.configManager.SaveConfig(cfg); err != nil {
		log.Printf("Error saving config: %v", err)
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	// Reload service
	if err := s.serviceManager.Reload(); err != nil {
		log.Printf("Warning: failed to reload service: %v", err)
	}

	// Return updated list
	s.handleRuleActionsList(w, r)
}

// handleRuleActionUpdate handles updating an existing rule action
func (s *Server) handleRuleActionUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	indexStr := r.FormValue("index")
	if indexStr == "" {
		http.Error(w, "No index provided", http.StatusBadRequest)
		return
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	action := s.buildRuleActionFromForm(r)

	// Get current config
	cfg, err := s.configManager.LoadConfig()
	if err != nil {
		http.Error(w, "Failed to get config", http.StatusInternalServerError)
		return
	}

	// Validate index
	if cfg.Route == nil || cfg.Route.RuleAction == nil || index < 0 || index >= len(cfg.Route.RuleAction) {
		http.Error(w, "Invalid action index", http.StatusBadRequest)
		return
	}

	// Update action
	cfg.Route.RuleAction[index] = action

	// Save config
	if err := s.configManager.SaveConfig(cfg); err != nil {
		log.Printf("Error saving config: %v", err)
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	// Reload service
	if err := s.serviceManager.Reload(); err != nil {
		log.Printf("Warning: failed to reload service: %v", err)
	}

	// Return updated list
	s.handleRuleActionsList(w, r)
}

// handleRuleActionDelete handles deleting a rule action
func (s *Server) handleRuleActionDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	indexStr := r.URL.Query().Get("index")
	if indexStr == "" {
		http.Error(w, "No index provided", http.StatusBadRequest)
		return
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	// Get current config
	cfg, err := s.configManager.LoadConfig()
	if err != nil {
		http.Error(w, "Failed to get config", http.StatusInternalServerError)
		return
	}

	// Validate index
	if cfg.Route == nil || cfg.Route.RuleAction == nil || index < 0 || index >= len(cfg.Route.RuleAction) {
		http.Error(w, "Invalid action index", http.StatusBadRequest)
		return
	}

	// Remove action
	cfg.Route.RuleAction = append(cfg.Route.RuleAction[:index], cfg.Route.RuleAction[index+1:]...)

	// Save config
	if err := s.configManager.SaveConfig(cfg); err != nil {
		log.Printf("Error saving config: %v", err)
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	// Reload service
	if err := s.serviceManager.Reload(); err != nil {
		log.Printf("Warning: failed to reload service: %v", err)
	}

	// Return updated list
	s.handleRuleActionsList(w, r)
}

// buildRuleActionFromForm builds a rule action map from form data
func (s *Server) buildRuleActionFromForm(r *http.Request) map[string]interface{} {
	action := make(map[string]interface{})

	// Get action type
	actionType := r.FormValue("action")
	if actionType != "" {
		action["action"] = actionType
	}

	// Add fields based on action type
	switch actionType {
	case "route":
		if outbound := r.FormValue("outbound"); outbound != "" {
			action["outbound"] = outbound
		}

	case "sniff":
		if sniffers := r.Form["sniffer[]"]; len(sniffers) > 0 {
			var validSniffers []string
			for _, s := range sniffers {
				s = strings.TrimSpace(s)
				if s != "" {
					validSniffers = append(validSniffers, s)
				}
			}
			if len(validSniffers) > 0 {
				action["sniffer"] = validSniffers
			}
		}
		if timeout := r.FormValue("timeout"); timeout != "" {
			if val, err := strconv.ParseUint(timeout, 10, 32); err == nil {
				action["timeout"] = uint32(val)
			}
		}

	case "resolve":
		if server := r.FormValue("server"); server != "" {
			action["server"] = server
		}
		if strategy := r.FormValue("strategy"); strategy != "" {
			action["strategy"] = strategy
		}
		if r.FormValue("disable_cache") == "on" {
			action["disable_cache"] = true
		}
		if rewriteTTL := r.FormValue("rewrite_ttl"); rewriteTTL != "" {
			if val, err := strconv.ParseUint(rewriteTTL, 10, 32); err == nil {
				ttl := uint32(val)
				action["rewrite_ttl"] = &ttl
			}
		}
		if clientSubnet := r.FormValue("client_subnet"); clientSubnet != "" {
			action["client_subnet"] = &clientSubnet
		}

	case "reject":
		if method := r.FormValue("method"); method != "" {
			action["method"] = method
		}
		if r.FormValue("no_drop") == "on" {
			action["no_drop"] = true
		}

	case "route-options":
		if outbound := r.FormValue("outbound"); outbound != "" {
			action["outbound"] = outbound
		}
		if overrideAddress := r.FormValue("override_address"); overrideAddress != "" {
			action["override_address"] = overrideAddress
		}
		if overridePort := r.FormValue("override_port"); overridePort != "" {
			if val, err := strconv.ParseUint(overridePort, 10, 16); err == nil {
				action["override_port"] = uint16(val)
			}
		}
		if networkStrategy := r.FormValue("network_strategy"); networkStrategy != "" {
			action["network_strategy"] = &networkStrategy
		}
		if fallbackDelay := r.FormValue("fallback_delay"); fallbackDelay != "" {
			if val, err := strconv.ParseUint(fallbackDelay, 10, 32); err == nil {
				action["fallback_delay"] = uint32(val)
			}
		}
		if udpTimeout := r.FormValue("udp_timeout"); udpTimeout != "" {
			if val, err := strconv.ParseUint(udpTimeout, 10, 32); err == nil {
				action["udp_timeout"] = uint32(val)
			}
		}
		if r.FormValue("udp_disable_domain_unmapping") == "on" {
			action["udp_disable_domain_unmapping"] = true
		}
		if r.FormValue("udp_connect") == "on" {
			action["udp_connect"] = true
		}
		if r.FormValue("tls_fragment") == "on" {
			action["tls_fragment"] = true
		}
		if tlsFragmentFallbackDelay := r.FormValue("tls_fragment_fallback_delay"); tlsFragmentFallbackDelay != "" {
			if val, err := strconv.ParseUint(tlsFragmentFallbackDelay, 10, 32); err == nil {
				action["tls_fragment_fallback_delay"] = uint32(val)
			}
		}
		if r.FormValue("tls_record_fragment") == "on" {
			action["tls_record_fragment"] = true
		}
	}

	return action
}
