package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for local development
	},
}

// handleConnectionsPage handles the connections monitoring page
func (s *Server) handleConnectionsPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Live Connections",
		Data:  nil,
	}

	if err := s.renderTemplate(w, "connections.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleConnectionsWebSocket handles WebSocket proxy to Clash API
func (s *Server) handleConnectionsWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get Clash API URL from query parameter or use default
	clashAPIURL := r.URL.Query().Get("clash_api")
	if clashAPIURL == "" {
		// Default Clash API URL - you can make this configurable
		clashAPIURL = "http://127.0.0.1:9090"
	}

	// Upgrade HTTP connection to WebSocket
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer clientConn.Close()

	// Parse the Clash API URL
	u, err := url.Parse(clashAPIURL)
	if err != nil {
		log.Printf("Invalid Clash API URL: %v", err)
		clientConn.WriteJSON(map[string]string{"error": "Invalid Clash API URL"})
		return
	}

	// Construct WebSocket URL for Clash API connections endpoint
	wsScheme := "ws"
	if u.Scheme == "https" {
		wsScheme = "wss"
	}
	clashWSURL := fmt.Sprintf("%s://%s/connections", wsScheme, u.Host)

	log.Printf("Connecting to Clash API WebSocket: %s", clashWSURL)

	// Connect to Clash API WebSocket
	clashConn, _, err := websocket.DefaultDialer.Dial(clashWSURL, nil)
	if err != nil {
		log.Printf("Failed to connect to Clash API: %v", err)
		clientConn.WriteJSON(map[string]string{"error": fmt.Sprintf("Failed to connect to Clash API: %v", err)})
		return
	}
	defer clashConn.Close()

	// Create channels for bidirectional communication
	done := make(chan struct{})
	errChan := make(chan error, 2)

	// Forward messages from Clash API to client
	go func() {
		defer close(done)
		for {
			_, message, err := clashConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Clash API WebSocket error: %v", err)
				}
				errChan <- err
				return
			}

			// Parse and validate the message
			var connMsg map[string]interface{}
			if err := json.Unmarshal(message, &connMsg); err != nil {
				log.Printf("Failed to parse Clash API message: %v", err)
				continue
			}

			// Forward to client
			if err := clientConn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Failed to write to client: %v", err)
				errChan <- err
				return
			}
		}
	}()

	// Forward messages from client to Clash API (if needed)
	go func() {
		for {
			_, message, err := clientConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Client WebSocket error: %v", err)
				}
				errChan <- err
				return
			}

			// Forward to Clash API
			if err := clashConn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Failed to write to Clash API: %v", err)
				errChan <- err
				return
			}
		}
	}()

	// Wait for done or error
	select {
	case <-done:
		log.Println("Clash API connection closed")
	case err := <-errChan:
		log.Printf("WebSocket proxy error: %v", err)
	}
}

// handleConnectionToRule handles creating a rule from connection data
func (s *Server) handleConnectionToRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Extract connection properties from form
	action := r.FormValue("action")
	sourceIP := r.FormValue("source_ip")
	destinationIP := r.FormValue("destination_ip")
	destinationPort := r.FormValue("destination_port")
	network := r.FormValue("network")
	domain := r.FormValue("domain")
	outbound := r.FormValue("outbound")

	// Validate action
	if action == "" {
		http.Error(w, "Action is required", http.StatusBadRequest)
		return
	}

	// Build rule from selected properties
	rule := make(map[string]interface{})

	// Add action field
	rule["action"] = action

	// Add matching fields
	if sourceIP != "" {
		rule["source_ip_cidr"] = []string{sourceIP + "/32"}
	}
	if destinationIP != "" {
		rule["ip_cidr"] = []string{destinationIP + "/32"}
	}
	if destinationPort != "" {
		rule["port"] = []string{destinationPort}
	}
	if network != "" {
		rule["network"] = []string{network}
	}
	if domain != "" {
		rule["domain_suffix"] = []string{domain}
	}

	// Add outbound for actions that need it
	if outbound != "" && (action == "route" || action == "route-options") {
		rule["outbound"] = outbound
	}

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

	// Return success
	w.Header().Set("HX-Trigger", "ruleCreated")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Rule created successfully")
}
