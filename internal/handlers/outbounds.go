package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// handleOutboundsPage handles the outbounds management page
func (s *Server) handleOutboundsPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Outbound Management",
		Data:  map[string]interface{}{},
	}

	if err := s.renderTemplate(w, "outbounds.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleOutboundsList handles the HTMX endpoint for outbounds list
func (s *Server) handleOutboundsList(w http.ResponseWriter, r *http.Request) {
	outbounds, err := s.configManager.GetOutbounds()
	if err != nil {
		log.Printf("Error getting outbounds: %v", err)
		http.Error(w, "Failed to load outbounds", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Outbounds": outbounds,
	}

	if err := s.renderTemplate(w, "outbound-list.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleOutboundForm handles the HTMX endpoint for outbound forms
func (s *Server) handleOutboundForm(w http.ResponseWriter, r *http.Request) {
	outboundType := r.URL.Query().Get("type")
	indexStr := r.URL.Query().Get("index")
	editMode := indexStr != ""

	var outboundData map[string]interface{}
	var originalTag string

	if editMode {
		// Get existing outbound for editing
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}

		outbounds, err := s.configManager.GetOutbounds()
		if err != nil {
			log.Printf("Error getting outbounds: %v", err)
			http.Error(w, "Failed to get outbounds", http.StatusInternalServerError)
			return
		}

		if index < 0 || index >= len(outbounds) {
			http.Error(w, "Index out of range", http.StatusBadRequest)
			return
		}

		// Convert outbound to map
		if outbound, ok := outbounds[index].(map[string]interface{}); ok {
			outboundData = outbound
			if tag, ok := outbound["tag"].(string); ok {
				originalTag = tag
			}
			// Get type from outbound data
			if outboundType == "" {
				if t, ok := outbound["type"].(string); ok {
					outboundType = t
				}
			}
		} else {
			http.Error(w, "Invalid outbound format", http.StatusInternalServerError)
			return
		}
	}

	if outboundType == "" {
		outboundType = "direct" // Default type
	}

	// Get all outbound tags for selector/urltest outbounds
	allOutbounds, err := s.configManager.GetOutboundTags()
	if err != nil {
		log.Printf("Warning: failed to get outbound tags: %v", err)
		allOutbounds = []string{}
	}

	// Build form structure based on outbound type
	formFields := s.buildOutboundFormFields(outboundType, allOutbounds)

	// Populate form with existing values if editing
	if editMode && outboundData != nil {
		populateOutboundFormValues(formFields, outboundData)
	}

	data := map[string]interface{}{
		"Fields":        formFields,
		"OutboundType":  outboundType,
		"OutboundTypes": getAvailableOutboundTypes(),
		"EditMode":      editMode,
		"OriginalTag":   originalTag,
		"AllOutbounds":  allOutbounds,
	}

	if err := s.renderTemplate(w, "outbound-form.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleOutboundCreate handles creating a new outbound
func (s *Server) handleOutboundCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Build outbound from form data
	outbound := buildOutboundFromForm(r.Form)

	// Validate required fields
	if err := validateOutbound(outbound); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get current outbounds
	outbounds, err := s.configManager.GetOutbounds()
	if err != nil {
		log.Printf("Error getting outbounds: %v", err)
		http.Error(w, "Failed to get outbounds", http.StatusInternalServerError)
		return
	}

	// Add new outbound
	outbounds = append(outbounds, outbound)

	// Save updated outbounds
	if err := s.configManager.UpdateOutbounds(outbounds); err != nil {
		log.Printf("Error updating outbounds: %v", err)
		http.Error(w, "Failed to save outbounds", http.StatusInternalServerError)
		return
	}

	// Reload service
	if err := s.serviceManager.Reload(); err != nil {
		log.Printf("Warning: failed to reload service: %v", err)
	}

	// Return updated list
	w.Header().Set("HX-Trigger", "outboundCreated")
	s.handleOutboundsList(w, r)
}

// handleOutboundUpdate handles updating an existing outbound
func (s *Server) handleOutboundUpdate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	originalTag := r.FormValue("original_tag")
	if originalTag == "" {
		http.Error(w, "Missing original_tag", http.StatusBadRequest)
		return
	}

	// Build outbound from form data
	updatedOutbound := buildOutboundFromForm(r.Form)

	// Validate required fields
	if err := validateOutbound(updatedOutbound); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get current outbounds
	outbounds, err := s.configManager.GetOutbounds()
	if err != nil {
		log.Printf("Error getting outbounds: %v", err)
		http.Error(w, "Failed to get outbounds", http.StatusInternalServerError)
		return
	}

	newTag, _ := updatedOutbound["tag"].(string)
	updateIndex := -1
	for i, outbound := range outbounds {
		if outboundMap, ok := outbound.(map[string]interface{}); ok {
			if tag, ok := outboundMap["tag"].(string); ok && tag == originalTag {
				updateIndex = i
				break
			}
		}
	}

	if updateIndex == -1 {
		http.Error(w, "Outbound to update not found", http.StatusBadRequest)
		return
	}

	// Update outbound
	outbounds[updateIndex] = updatedOutbound

	// If tag was changed, update references in other outbounds
	if newTag != "" && newTag != originalTag {
		for _, outbound := range outbounds {
			if outboundMap, ok := outbound.(map[string]interface{}); ok {
				if outboundType, ok := outboundMap["type"].(string); ok && (outboundType == "selector" || outboundType == "urltest") {
					// Update members
					if members, ok := outboundMap["outbounds"].([]interface{}); ok {
						for j, member := range members {
							if memberStr, ok := member.(string); ok && memberStr == originalTag {
								members[j] = newTag
							}
						}
						outboundMap["outbounds"] = members
					}
					// Update default selection
					if def, ok := outboundMap["default"].(string); ok && def == originalTag {
						outboundMap["default"] = newTag
					}
				}
			}
		}
	}

	// Save updated outbounds
	if err := s.configManager.UpdateOutbounds(outbounds); err != nil {
		log.Printf("Error updating outbounds: %v", err)
		http.Error(w, "Failed to save outbounds", http.StatusInternalServerError)
		return
	}

	// Reload service
	if err := s.serviceManager.Reload(); err != nil {
		log.Printf("Warning: failed to reload service: %v", err)
	}

	// Return updated list
	w.Header().Set("HX-Trigger", "outboundUpdated")
	s.handleOutboundsList(w, r)
}

// handleOutboundDelete handles deleting an outbound
func (s *Server) handleOutboundDelete(w http.ResponseWriter, r *http.Request) {
	tagToDelete := r.URL.Query().Get("tag")
	if tagToDelete == "" {
		http.Error(w, "Invalid tag", http.StatusBadRequest)
		return
	}

	// Get current outbounds
	outbounds, err := s.configManager.GetOutbounds()
	if err != nil {
		log.Printf("Error getting outbounds: %v", err)
		http.Error(w, "Failed to get outbounds", http.StatusInternalServerError)
		return
	}

	// Remove references from groups
	for _, outbound := range outbounds {
		if outboundMap, ok := outbound.(map[string]interface{}); ok {
			outboundType, isString := outboundMap["type"].(string)
			if !isString {
				continue
			}

			if outboundType == "selector" || outboundType == "urltest" {
				// Remove from 'outbounds' list
				if members, ok := outboundMap["outbounds"].([]interface{}); ok {
					var newMembers []interface{}
					for _, member := range members {
						if memberStr, ok := member.(string); ok && memberStr != tagToDelete {
							newMembers = append(newMembers, member)
						}
					}
					outboundMap["outbounds"] = newMembers
				}

				// If it's a selector, check the 'default' field
				if outboundType == "selector" {
					if def, ok := outboundMap["default"].(string); ok && def == tagToDelete {
						outboundMap["default"] = "" // Reset default
						if newMembers, ok := outboundMap["outbounds"].([]interface{}); ok && len(newMembers) > 0 {
							if newDefault, ok := newMembers[0].(string); ok {
								outboundMap["default"] = newDefault
							}
						}
					}
				}
			}
		}
	}

	deleteIndex := -1
	for i, outbound := range outbounds {
		if outboundMap, ok := outbound.(map[string]interface{}); ok {
			if tag, ok := outboundMap["tag"].(string); ok && tag == tagToDelete {
				deleteIndex = i
				break
			}
		}
	}

	if deleteIndex == -1 {
		http.Error(w, "Outbound not found", http.StatusBadRequest)
		return
	}

	// Remove outbound
	outbounds = append(outbounds[:deleteIndex], outbounds[deleteIndex+1:]...)

	// Save updated outbounds
	if err := s.configManager.UpdateOutbounds(outbounds); err != nil {
		log.Printf("Error updating outbounds: %v", err)
		http.Error(w, "Failed to save outbounds", http.StatusInternalServerError)
		return
	}

	// Reload service
	if err := s.serviceManager.Reload(); err != nil {
		log.Printf("Warning: failed to reload service: %v", err)
	}

	// Return updated list
	w.Header().Set("HX-Trigger", "outboundDeleted")
	s.handleOutboundsList(w, r)
}

// handleOutboundReorder handles reordering outbounds (drag-and-drop)
func (s *Server) handleOutboundReorder(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from, err := strconv.Atoi(fromStr)
	if err != nil {
		http.Error(w, "Invalid from index", http.StatusBadRequest)
		return
	}

	to, err := strconv.Atoi(toStr)
	if err != nil {
		http.Error(w, "Invalid to index", http.StatusBadRequest)
		return
	}

	// Get current outbounds
	outbounds, err := s.configManager.GetOutbounds()
	if err != nil {
		log.Printf("Error getting outbounds: %v", err)
		http.Error(w, "Failed to get outbounds", http.StatusInternalServerError)
		return
	}

	if from < 0 || from >= len(outbounds) || to < 0 || to >= len(outbounds) {
		http.Error(w, "Index out of range", http.StatusBadRequest)
		return
	}

	// Reorder
	item := outbounds[from]
	outbounds = append(outbounds[:from], outbounds[from+1:]...)
	if to > from {
		to--
	}
	outbounds = append(outbounds[:to], append([]interface{}{item}, outbounds[to:]...)...)

	// Save updated outbounds
	if err := s.configManager.UpdateOutbounds(outbounds); err != nil {
		log.Printf("Error updating outbounds: %v", err)
		http.Error(w, "Failed to save outbounds", http.StatusInternalServerError)
		return
	}

	// Reload service
	if err := s.serviceManager.Reload(); err != nil {
		log.Printf("Warning: failed to reload service: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleOutboundRename handles renaming an outbound and updating all references
func (s *Server) handleOutboundRename(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	oldTag := r.FormValue("old_tag")
	newTag := r.FormValue("new_tag")

	if oldTag == "" || newTag == "" {
		http.Error(w, "Missing tag parameters", http.StatusBadRequest)
		return
	}

	// Rename outbound and update all references
	if err := s.configManager.RenameOutbound(oldTag, newTag); err != nil {
		log.Printf("Error renaming outbound: %v", err)
		http.Error(w, "Failed to rename outbound", http.StatusInternalServerError)
		return
	}

	// Reload service
	if err := s.serviceManager.Reload(); err != nil {
		log.Printf("Warning: failed to reload service: %v", err)
	}

	// Return updated list
	w.Header().Set("HX-Trigger", "outboundRenamed")
	s.handleOutboundsList(w, r)
}

// handleGroupManage handles managing outbound groups (selector/urltest)
func (s *Server) handleGroupManage(w http.ResponseWriter, r *http.Request) {
	indexStr := r.URL.Query().Get("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	outbounds, err := s.configManager.GetOutbounds()
	if err != nil {
		log.Printf("Error getting outbounds: %v", err)
		http.Error(w, "Failed to get outbounds", http.StatusInternalServerError)
		return
	}

	if index < 0 || index >= len(outbounds) {
		http.Error(w, "Index out of range", http.StatusBadRequest)
		return
	}

	outbound, ok := outbounds[index].(map[string]interface{})
	if !ok {
		http.Error(w, "Invalid outbound format", http.StatusInternalServerError)
		return
	}

	outboundType, _ := outbound["type"].(string)
	if outboundType != "selector" && outboundType != "urltest" {
		http.Error(w, "Outbound is not a group type", http.StatusBadRequest)
		return
	}

	// Get all available outbounds for selection
	allTags, err := s.configManager.GetOutboundTags()
	if err != nil {
		log.Printf("Error getting outbound tags: %v", err)
		allTags = []string{}
	}

	// Get current group members
	var groupMembers []string
	if members, ok := outbound["outbounds"].([]interface{}); ok {
		for _, m := range members {
			if tag, ok := m.(string); ok {
				groupMembers = append(groupMembers, tag)
			}
		}
	}

	// Filter out the current outbound tag from available outbounds
	currentTag, _ := outbound["tag"].(string)
	var availableOutbounds []string
	for _, tag := range allTags {
		if tag != currentTag && !contains(groupMembers, tag) {
			availableOutbounds = append(availableOutbounds, tag)
		}
	}

	data := map[string]interface{}{
		"Outbound":           outbound,
		"Index":              index,
		"GroupMembers":       groupMembers,
		"AvailableOutbounds": availableOutbounds,
	}

	if err := s.renderTemplate(w, "group-manage.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleGroupUpdate handles updating group members
func (s *Server) handleGroupUpdate(w http.ResponseWriter, r *http.Request) {
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

	// Get selected members
	members := r.Form["members[]"]
	if len(members) == 0 {
		http.Error(w, "At least one member is required", http.StatusBadRequest)
		return
	}

	outbounds, err := s.configManager.GetOutbounds()
	if err != nil {
		log.Printf("Error getting outbounds: %v", err)
		http.Error(w, "Failed to get outbounds", http.StatusInternalServerError)
		return
	}

	if index < 0 || index >= len(outbounds) {
		http.Error(w, "Index out of range", http.StatusBadRequest)
		return
	}

	outbound, ok := outbounds[index].(map[string]interface{})
	if !ok {
		http.Error(w, "Invalid outbound format", http.StatusInternalServerError)
		return
	}

	// Update group members
	var memberInterfaces []interface{}
	for _, m := range members {
		memberInterfaces = append(memberInterfaces, m)
	}
	outbound["outbounds"] = memberInterfaces

	// Save updated outbounds
	if err := s.configManager.UpdateOutbounds(outbounds); err != nil {
		log.Printf("Error updating outbounds: %v", err)
		http.Error(w, "Failed to save outbounds", http.StatusInternalServerError)
		return
	}

	// Reload service
	if err := s.serviceManager.Reload(); err != nil {
		log.Printf("Warning: failed to reload service: %v", err)
	}

	// Return updated list
	w.Header().Set("HX-Trigger", "groupUpdated")
	s.handleOutboundsList(w, r)
}

// Helper functions

func getAvailableOutboundTypes() []map[string]string {
	return []map[string]string{
		{"value": "direct", "label": "Direct", "description": "Direct connection"},
		{"value": "block", "label": "Block", "description": "Block connection"},
		{"value": "dns", "label": "DNS", "description": "DNS outbound"},
		{"value": "socks", "label": "SOCKS", "description": "SOCKS proxy"},
		{"value": "http", "label": "HTTP", "description": "HTTP proxy"},
		{"value": "shadowsocks", "label": "Shadowsocks", "description": "Shadowsocks protocol"},
		{"value": "vmess", "label": "VMess", "description": "VMess protocol"},
		{"value": "vless", "label": "VLESS", "description": "VLESS protocol"},
		{"value": "trojan", "label": "Trojan", "description": "Trojan protocol"},
		{"value": "wireguard", "label": "WireGuard", "description": "WireGuard VPN"},
		{"value": "hysteria", "label": "Hysteria", "description": "Hysteria protocol"},
		{"value": "hysteria2", "label": "Hysteria2", "description": "Hysteria2 protocol"},
		{"value": "tuic", "label": "TUIC", "description": "TUIC protocol"},
		{"value": "ssh", "label": "SSH", "description": "SSH tunnel"},
		{"value": "tor", "label": "Tor", "description": "Tor network"},
		{"value": "selector", "label": "Selector", "description": "Manual selection group"},
		{"value": "urltest", "label": "URLTest", "description": "Auto selection group"},
	}
}

func buildOutboundFromForm(form map[string][]string) map[string]interface{} {
	outbound := make(map[string]interface{})

	for key, values := range form {
		if key == "index" || key == "original_tag" {
			continue // Skip index and original_tag fields
		}

		if len(values) == 0 || values[0] == "" {
			continue // Skip empty values
		}

		// Handle array fields (ending with [])
		if strings.HasSuffix(key, "[]") {
			actualKey := strings.TrimSuffix(key, "[]")
			var arrayValues []interface{}
			for _, v := range values {
				if v != "" {
					// Handle comma-separated values
					if strings.Contains(v, ",") {
						for _, item := range strings.Split(v, ",") {
							item = strings.TrimSpace(item)
							if item != "" {
								arrayValues = append(arrayValues, item)
							}
						}
					} else {
						arrayValues = append(arrayValues, v)
					}
				}
			}
			if len(arrayValues) > 0 {
				outbound[actualKey] = arrayValues
			}
			continue
		}

		value := values[0]

		// Try to parse as number
		if intVal, err := strconv.Atoi(value); err == nil {
			outbound[key] = intVal
			continue
		}

		// Try to parse as boolean
		if value == "true" {
			outbound[key] = true
			continue
		} else if value == "false" {
			outbound[key] = false
			continue
		}

		// Handle JSON objects
		if strings.HasPrefix(value, "{") || strings.HasPrefix(value, "[") {
			var jsonValue interface{}
			if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
				outbound[key] = jsonValue
				continue
			}
		}

		// Default to string
		outbound[key] = value
	}

	return outbound
}

func validateOutbound(outbound map[string]interface{}) error {
	outboundType, ok := outbound["type"].(string)
	if !ok || outboundType == "" {
		return fmt.Errorf("outbound type is required")
	}

	tag, ok := outbound["tag"].(string)
	if !ok || tag == "" {
		return fmt.Errorf("outbound tag is required")
	}

	// Type-specific validation
	switch outboundType {
	case "socks", "http", "shadowsocks", "vmess", "vless", "trojan", "wireguard", "hysteria", "hysteria2", "tuic", "ssh":
		if _, ok := outbound["server"]; !ok {
			return fmt.Errorf("server is required for %s outbound", outboundType)
		}
		if _, ok := outbound["server_port"]; !ok {
			return fmt.Errorf("server_port is required for %s outbound", outboundType)
		}
	case "selector", "urltest":
		outbounds, ok := outbound["outbounds"].([]interface{})
		if !ok || len(outbounds) == 0 {
			return fmt.Errorf("at least one outbound is required for %s", outboundType)
		}
	}

	return nil
}

func populateOutboundFormValues(fields []FormField, data map[string]interface{}) {
	for i := range fields {
		field := &fields[i]
		if value, ok := data[field.Name]; ok {
			if field.IsArray {
				if arrayValue, ok := value.([]interface{}); ok {
					var strValues []string
					for _, v := range arrayValue {
						strValues = append(strValues, fmt.Sprintf("%v", v))
					}
					field.Values = strValues
				}
			} else {
				field.Value = value
			}
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// FormField represents a field in the outbound form
type FormField struct {
	Name        string
	Label       string
	Type        string
	Placeholder string
	Required    bool
	IsArray     bool
	Options     []string
	Description string
	Value       interface{}
	Values      []string
}

func (s *Server) buildOutboundFormFields(outboundType string, allOutbounds []string) []FormField {
	commonFields := []FormField{
		{Name: "type", Label: "Type", Type: "hidden", Value: outboundType, Required: true},
		{Name: "tag", Label: "Tag", Type: "text", Placeholder: "my-outbound", Required: true, Description: "Unique identifier for this outbound"},
	}

	var specificFields []FormField

	switch outboundType {
	case "direct":
		specificFields = []FormField{
			{Name: "override_address", Label: "Override Address", Type: "text", Placeholder: "1.1.1.1", Description: "Override destination address"},
			{Name: "override_port", Label: "Override Port", Type: "number", Placeholder: "53", Description: "Override destination port"},
		}
	case "block", "dns":
		// No additional fields
	case "socks":
		specificFields = []FormField{
			{Name: "server", Label: "Server", Type: "text", Placeholder: "127.0.0.1", Required: true},
			{Name: "server_port", Label: "Server Port", Type: "number", Placeholder: "1080", Required: true},
			{Name: "version", Label: "Version", Type: "select", Options: []string{"5", "4", "4a"}, Description: "SOCKS protocol version"},
			{Name: "username", Label: "Username", Type: "text"},
			{Name: "password", Label: "Password", Type: "password"},
			{Name: "network", Label: "Network", Type: "select", Options: []string{"tcp", "udp", "tcp,udp"}},
		}
	case "http":
		specificFields = []FormField{
			{Name: "server", Label: "Server", Type: "text", Placeholder: "127.0.0.1", Required: true},
			{Name: "server_port", Label: "Server Port", Type: "number", Placeholder: "8080", Required: true},
			{Name: "username", Label: "Username", Type: "text"},
			{Name: "password", Label: "Password", Type: "password"},
			{Name: "path", Label: "Path", Type: "text", Placeholder: "/"},
		}
	case "shadowsocks":
		specificFields = []FormField{
			{Name: "server", Label: "Server", Type: "text", Placeholder: "example.com", Required: true},
			{Name: "server_port", Label: "Server Port", Type: "number", Placeholder: "8388", Required: true},
			{Name: "method", Label: "Method", Type: "select", Required: true, Options: []string{
				"2022-blake3-aes-128-gcm", "2022-blake3-aes-256-gcm", "2022-blake3-chacha20-poly1305",
				"aes-128-gcm", "aes-256-gcm", "chacha20-ietf-poly1305",
			}},
			{Name: "password", Label: "Password", Type: "password", Required: true},
			{Name: "network", Label: "Network", Type: "select", Options: []string{"tcp", "udp", "tcp,udp"}},
		}
	case "vmess":
		specificFields = []FormField{
			{Name: "server", Label: "Server", Type: "text", Placeholder: "example.com", Required: true},
			{Name: "server_port", Label: "Server Port", Type: "number", Placeholder: "443", Required: true},
			{Name: "uuid", Label: "UUID", Type: "text", Placeholder: "uuid-here", Required: true},
			{Name: "security", Label: "Security", Type: "select", Options: []string{"auto", "none", "aes-128-gcm", "chacha20-poly1305"}},
			{Name: "alter_id", Label: "Alter ID", Type: "number", Placeholder: "0"},
		}
	case "vless":
		specificFields = []FormField{
			{Name: "server", Label: "Server", Type: "text", Placeholder: "example.com", Required: true},
			{Name: "server_port", Label: "Server Port", Type: "number", Placeholder: "443", Required: true},
			{Name: "uuid", Label: "UUID", Type: "text", Placeholder: "uuid-here", Required: true},
			{Name: "flow", Label: "Flow", Type: "select", Options: []string{"", "xtls-rprx-vision"}},
		}
	case "trojan":
		specificFields = []FormField{
			{Name: "server", Label: "Server", Type: "text", Placeholder: "example.com", Required: true},
			{Name: "server_port", Label: "Server Port", Type: "number", Placeholder: "443", Required: true},
			{Name: "password", Label: "Password", Type: "password", Required: true},
		}
	case "wireguard":
		specificFields = []FormField{
			{Name: "server", Label: "Server", Type: "text", Placeholder: "example.com", Required: true},
			{Name: "server_port", Label: "Server Port", Type: "number", Placeholder: "51820", Required: true},
			{Name: "local_address[]", Label: "Local Address", Type: "array", Placeholder: "10.0.0.2/32", Required: true, IsArray: true, Description: "Local IP address(es)"},
			{Name: "private_key", Label: "Private Key", Type: "text", Required: true},
			{Name: "peer_public_key", Label: "Peer Public Key", Type: "text", Required: true},
			{Name: "pre_shared_key", Label: "Pre-Shared Key", Type: "text"},
			{Name: "mtu", Label: "MTU", Type: "number", Placeholder: "1408"},
		}
	case "hysteria":
		specificFields = []FormField{
			{Name: "server", Label: "Server", Type: "text", Placeholder: "example.com", Required: true},
			{Name: "server_port", Label: "Server Port", Type: "number", Placeholder: "443", Required: true},
			{Name: "up_mbps", Label: "Upload (Mbps)", Type: "number", Placeholder: "10"},
			{Name: "down_mbps", Label: "Download (Mbps)", Type: "number", Placeholder: "50"},
			{Name: "auth_str", Label: "Auth String", Type: "password"},
			{Name: "obfs", Label: "Obfuscation", Type: "text"},
		}
	case "hysteria2":
		specificFields = []FormField{
			{Name: "server", Label: "Server", Type: "text", Placeholder: "example.com", Required: true},
			{Name: "server_port", Label: "Server Port", Type: "number", Placeholder: "443", Required: true},
			{Name: "up_mbps", Label: "Upload (Mbps)", Type: "number", Placeholder: "10"},
			{Name: "down_mbps", Label: "Download (Mbps)", Type: "number", Placeholder: "50"},
			{Name: "password", Label: "Password", Type: "password"},
		}
	case "tuic":
		specificFields = []FormField{
			{Name: "server", Label: "Server", Type: "text", Placeholder: "example.com", Required: true},
			{Name: "server_port", Label: "Server Port", Type: "number", Placeholder: "443", Required: true},
			{Name: "uuid", Label: "UUID", Type: "text", Placeholder: "uuid-here", Required: true},
			{Name: "password", Label: "Password", Type: "password"},
			{Name: "congestion_control", Label: "Congestion Control", Type: "select", Options: []string{"cubic", "new_reno", "bbr"}},
		}
	case "ssh":
		specificFields = []FormField{
			{Name: "server", Label: "Server", Type: "text", Placeholder: "example.com", Required: true},
			{Name: "server_port", Label: "Server Port", Type: "number", Placeholder: "22", Required: true},
			{Name: "user", Label: "User", Type: "text", Required: true},
			{Name: "password", Label: "Password", Type: "password"},
			{Name: "private_key", Label: "Private Key", Type: "textarea", Description: "SSH private key content"},
		}
	case "tor":
		specificFields = []FormField{
			{Name: "executable_path", Label: "Tor Executable Path", Type: "text", Placeholder: "/usr/bin/tor"},
			{Name: "data_directory", Label: "Data Directory", Type: "text"},
		}
	case "selector":
		specificFields = []FormField{
			{Name: "outbounds[]", Label: "Outbounds", Type: "multiselect", Required: true, IsArray: true, Options: allOutbounds, Description: "Select outbounds for this group"},
			{Name: "default", Label: "Default", Type: "select", Options: allOutbounds, Description: "Default outbound to use"},
			{Name: "interrupt_exist_connections", Label: "Interrupt Existing Connections", Type: "checkbox"},
		}
	case "urltest":
		specificFields = []FormField{
			{Name: "outbounds[]", Label: "Outbounds", Type: "multiselect", Required: true, IsArray: true, Options: allOutbounds, Description: "Select outbounds to test"},
			{Name: "url", Label: "Test URL", Type: "text", Placeholder: "https://www.gstatic.com/generate_204"},
			{Name: "interval", Label: "Test Interval (seconds)", Type: "number", Placeholder: "180"},
			{Name: "tolerance", Label: "Tolerance (ms)", Type: "number", Placeholder: "50"},
			{Name: "interrupt_exist_connections", Label: "Interrupt Existing Connections", Type: "checkbox"},
		}
	}

	// Add common dialer options for applicable types
	if outboundType != "block" && outboundType != "dns" && outboundType != "selector" && outboundType != "urltest" {
		dialerFields := []FormField{
			{Name: "detour", Label: "Detour", Type: "select", Options: allOutbounds, Description: "Use another outbound as proxy chain"},
			{Name: "bind_interface", Label: "Bind Interface", Type: "text", Description: "Bind to specific network interface"},
			{Name: "connect_timeout", Label: "Connect Timeout (seconds)", Type: "number", Placeholder: "5"},
			{Name: "tcp_fast_open", Label: "TCP Fast Open", Type: "checkbox"},
		}
		specificFields = append(specificFields, dialerFields...)
	}

	return append(commonFields, specificFields...)
}
