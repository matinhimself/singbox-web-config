package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/matinhimself/singbox-web-config/internal/config"
	"github.com/matinhimself/singbox-web-config/internal/forms"
	"github.com/matinhimself/singbox-web-config/internal/service"
	"github.com/matinhimself/singbox-web-config/internal/watcher"
)

// Server represents the HTTP server
type Server struct {
	addr           string
	templates      *template.Template
	mux            *http.ServeMux
	configManager  *config.Manager
	serviceManager *service.Manager
	formBuilder    *forms.Builder
	watcher        *watcher.Watcher
}

// NewServer creates a new HTTP server
func NewServer(addr string, configPath string, singboxService string) (*Server, error) {
	// Create config manager
	configManager, err := config.NewManager(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager: %w", err)
	}

	// Create service manager
	serviceManager := service.NewManager(singboxService)

	// Create form builder
	formBuilder := forms.NewBuilder()

	s := &Server{
		addr:           addr,
		mux:            http.NewServeMux(),
		configManager:  configManager,
		serviceManager: serviceManager,
		formBuilder:    formBuilder,
	}

	// Load templates
	if err := s.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	// Setup file watcher
	fileWatcher, err := watcher.NewWatcher(configPath, func() {
		log.Println("Config file changed externally, reloading...")
		// You could add logic here to notify connected clients via SSE or WebSockets
	})
	if err != nil {
		log.Printf("Warning: failed to setup file watcher: %v", err)
	} else {
		s.watcher = fileWatcher
		s.watcher.Start()
	}

	// Setup routes
	s.setupRoutes()

	return s, nil
}

// loadTemplates loads all HTML templates
func (s *Server) loadTemplates() error {
	// Parse all templates in the web/templates directory with custom functions
	templatesPath := filepath.Join("web", "templates", "*.html")
	tmpl, err := template.New("").Funcs(templateFuncMap()).ParseGlob(templatesPath)
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
	s.mux.HandleFunc("/service", s.handleServicePage)

	// API routes for rules (HTMX endpoints)
	s.mux.HandleFunc("/api/rules", s.handleRulesList)
	s.mux.HandleFunc("/api/rules/form", s.handleRuleForm)
	s.mux.HandleFunc("/api/rules/create", s.handleRuleCreate)
	s.mux.HandleFunc("/api/rules/delete", s.handleRuleDelete)
	s.mux.HandleFunc("/api/rules/update", s.handleRuleUpdate)

	// API routes for service management
	s.mux.HandleFunc("/api/service/status", s.handleServiceStatus)
	s.mux.HandleFunc("/api/service/start", s.handleServiceStart)
	s.mux.HandleFunc("/api/service/stop", s.handleServiceStop)
	s.mux.HandleFunc("/api/service/restart", s.handleServiceRestart)
	s.mux.HandleFunc("/api/service/logs", s.handleServiceLogs)

	// API routes for config management
	s.mux.HandleFunc("/api/config/export", s.handleConfigExport)
	s.mux.HandleFunc("/api/config/backups", s.handleConfigBackups)
	s.mux.HandleFunc("/api/config/restore", s.handleConfigRestore)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.addr)
	log.Printf("Visit http://%s in your browser", s.addr)
	return http.ListenAndServe(s.addr, s.mux)
}

// Stop stops the server and cleanup
func (s *Server) Stop() {
	if s.watcher != nil {
		s.watcher.Stop()
	}
}

// renderTemplate renders a template with the given data
func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	return s.templates.ExecuteTemplate(w, name, data)
}
