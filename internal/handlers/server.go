package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// Server represents the HTTP server
type Server struct {
	addr      string
	templates *template.Template
	mux       *http.ServeMux
}

// NewServer creates a new HTTP server
func NewServer(addr string) (*Server, error) {
	s := &Server{
		addr: addr,
		mux:  http.NewServeMux(),
	}

	// Load templates
	if err := s.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	// Setup routes
	s.setupRoutes()

	return s, nil
}

// loadTemplates loads all HTML templates
func (s *Server) loadTemplates() error {
	// Parse all templates in the web/templates directory
	templatesPath := filepath.Join("web", "templates", "*.html")
	tmpl, err := template.ParseGlob(templatesPath)
	if err != nil {
		return err
	}

	s.templates = tmpl
	return nil
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Static files
	fs := http.FileServer(http.Dir("web/static"))
	s.mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Page routes
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/rules", s.handleRulesPage)

	// API routes (HTMX endpoints)
	s.mux.HandleFunc("/api/rules", s.handleRulesList)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.addr)
	log.Printf("Visit http://%s in your browser", s.addr)
	return http.ListenAndServe(s.addr, s.mux)
}

// renderTemplate renders a template with the given data
func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	return s.templates.ExecuteTemplate(w, name, data)
}
