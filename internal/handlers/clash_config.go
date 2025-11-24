package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/matinhimself/singbox-web-config/internal/clash"
)

// ClashConfigResponse represents the Clash configuration response
type ClashConfigResponse struct {
	URL         string `json:"url"`
	Secret      string `json:"secret,omitempty"`
	HasSecret   bool   `json:"hasSecret"`
	IsConnected bool   `json:"isConnected"`
}

// ClashTestRequest represents a request to test Clash connection
type ClashTestRequest struct {
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

// ClashTestResponse represents the response from testing a Clash connection
type ClashTestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ClashUpdateRequest represents a request to update Clash configuration
type ClashUpdateRequest struct {
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

// ClashUpdateResponse represents the response from updating Clash configuration
type ClashUpdateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// handleClashConfig returns the current Clash configuration
func (s *Server) handleClashConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := ClashConfigResponse{
		URL:         s.clashURL,
		HasSecret:   s.clashSecret != "",
		IsConnected: s.clashClient != nil,
	}

	// Only send the secret if explicitly requested and it exists
	if r.URL.Query().Get("include_secret") == "true" && s.clashSecret != "" {
		response.Secret = s.clashSecret
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleClashTest tests a Clash API connection
func (s *Server) handleClashTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ClashTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Format URL
	url := formatClashURL(req.URL)
	if url == "" {
		response := ClashTestResponse{
			Success: false,
			Message: "URL is required",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Test the connection
	err := clash.TestConnection(url, req.Secret)
	response := ClashTestResponse{
		Success: err == nil,
	}

	if err != nil {
		response.Message = err.Error()
	} else {
		response.Message = "Connection successful"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleClashUpdate updates the Clash API configuration
func (s *Server) handleClashUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ClashUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Format URL
	url := formatClashURL(req.URL)
	if url == "" {
		response := ClashUpdateResponse{
			Success: false,
			Message: "URL is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Test the connection first
	if err := clash.TestConnection(url, req.Secret); err != nil {
		response := ClashUpdateResponse{
			Success: false,
			Message: "Failed to connect: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Update the configuration
	s.clashURL = url
	s.clashSecret = req.Secret
	s.clashClient = clash.NewClient(url, req.Secret)

	// Save the configuration
	if s.clashConfigMgr != nil {
		config := &clash.Config{
			URL:    url,
			Secret: req.Secret,
		}
		if err := s.clashConfigMgr.Save(config); err != nil {
			log.Printf("Warning: failed to save Clash config: %v", err)
		}
	}

	log.Printf("Clash API configuration updated: %s", url)

	response := ClashUpdateResponse{
		Success: true,
		Message: "Configuration updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// formatClashURL ensures the Clash URL has proper http:// prefix
func formatClashURL(url string) string {
	if url == "" {
		return ""
	}

	url = strings.TrimSpace(url)

	// Add http:// prefix if no protocol is specified
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	return url
}
