package handlers

import (
	"log"
	"net/http"

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

// handleRulesList handles the HTMX endpoint for rules list
func (s *Server) handleRulesList(w http.ResponseWriter, r *http.Request) {
	// For now, return empty list
	// Later this will return actual rules from storage

	if err := s.renderTemplate(w, "rule-list.html", nil); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
