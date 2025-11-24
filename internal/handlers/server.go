package handlers

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/matinhimself/singbox-web-config/internal/clash"
	"github.com/matinhimself/singbox-web-config/internal/config"
	"github.com/matinhimself/singbox-web-config/internal/forms"
	"github.com/matinhimself/singbox-web-config/internal/service"
	"github.com/matinhimself/singbox-web-config/internal/watcher"
)

// Server represents the HTTP server
type Server struct {
	addr              string
	templates         *template.Template
	mux               *http.ServeMux
	configManager     *config.Manager
	serviceManager    *service.Manager
	formBuilder       *forms.Builder
	watcher           *watcher.Watcher
	templatesFS       embed.FS
	staticFS          embed.FS
	clashClient       *clash.Client
	clashURL          string
	clashSecret       string
	clashConfigMgr    *clash.ConfigManager
}

// NewServer creates a new HTTP server
func NewServer(addr string, configPath string, singboxService string, clashURL string, clashSecret string, templatesFS, staticFS embed.FS) (*Server, error) {
	// Create config manager
	configManager, err := config.NewManager(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager: %w", err)
	}

	// Create initial backup if config exists
	if err := configManager.CreateBackupWithName("Initial backup", "Automatic backup created on server startup"); err != nil {
		log.Printf("Warning: failed to create initial backup: %v", err)
	} else {
		log.Println("Created initial backup on startup")
	}

	// Create service manager
	serviceManager := service.NewManager(singboxService)

	// Create form builder
	formBuilder := forms.NewBuilder()

	// Create Clash config manager
	clashConfigMgr, err := clash.NewConfigManager()
	if err != nil {
		log.Printf("Warning: failed to create Clash config manager: %v", err)
	}

	// Determine Clash API configuration
	var formattedClashURL string
	var finalClashSecret string

	// Priority: 1. CLI args, 2. Saved config, 3. Auto-detect
	if clashURL != "" {
		// Use CLI arguments
		formattedClashURL = formatClashURL(clashURL)
		finalClashSecret = clashSecret
		log.Printf("Using Clash API from CLI arguments: %s", formattedClashURL)
	} else if clashConfigMgr != nil {
		// Try to load saved configuration
		savedConfig, err := clashConfigMgr.Load()
		if err != nil {
			log.Printf("Warning: failed to load Clash config: %v", err)
		} else if savedConfig.URL != "" {
			formattedClashURL = savedConfig.URL
			finalClashSecret = savedConfig.Secret
			log.Printf("Loaded Clash API from saved config: %s", formattedClashURL)
		}
	}

	// If still not configured, try auto-detection
	if formattedClashURL == "" {
		log.Println("Attempting to auto-detect Clash API on port 9090...")
		if detected := clash.AutoDetect(); detected != nil {
			formattedClashURL = detected.URL
			finalClashSecret = detected.Secret
			log.Printf("Auto-detected Clash API: %s", formattedClashURL)

			// Save the auto-detected configuration
			if clashConfigMgr != nil {
				if err := clashConfigMgr.Save(detected); err != nil {
					log.Printf("Warning: failed to save auto-detected config: %v", err)
				}
			}
		} else {
			log.Println("Clash API not found. You can configure it through the web interface.")
		}
	}

	s := &Server{
		addr:           addr,
		mux:            http.NewServeMux(),
		configManager:  configManager,
		serviceManager: serviceManager,
		formBuilder:    formBuilder,
		templatesFS:    templatesFS,
		staticFS:       staticFS,
		clashURL:       formattedClashURL,
		clashSecret:    finalClashSecret,
		clashConfigMgr: clashConfigMgr,
	}

	// Initialize Clash client if URL is provided
	if formattedClashURL != "" {
		s.clashClient = clash.NewClient(formattedClashURL, finalClashSecret)
		log.Printf("Clash API client initialized: %s", formattedClashURL)
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

// loadTemplates loads all HTML templates from embedded files
func (s *Server) loadTemplates() error {
	// Use ParseFS to parse templates from embedded filesystem
	// This properly handles nested template definitions
	tmpl, err := template.New("").Funcs(templateFuncMap()).ParseFS(
		s.templatesFS,
		"web/templates/*.html",
		"web/templates/components/*.html",
	)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	s.templates = tmpl
	return nil
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Static files from embedded filesystem
	staticSubFS, err := fs.Sub(s.staticFS, "web/static")
	if err != nil {
		log.Printf("Warning: failed to load static files: %v", err)
	} else {
		fileServer := http.FileServer(http.FS(staticSubFS))
		s.mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
	}

	// Page routes
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/rules", s.handleRulesPage)
	s.mux.HandleFunc("/rule-actions", s.handleRuleActionsPage)
	s.mux.HandleFunc("/outbounds", s.handleOutboundsPage)
	s.mux.HandleFunc("/connections", s.handleConnectionsPage)
	s.mux.HandleFunc("/proxies", s.handleProxiesPage)
	s.mux.HandleFunc("/service", s.handleServicePage)

	// API routes for rules (HTMX endpoints)
	s.mux.HandleFunc("/api/rules", s.handleRulesList)
	s.mux.HandleFunc("/api/rules/form", s.handleRuleForm)
	s.mux.HandleFunc("/api/rules/create", s.handleRuleCreate)
	s.mux.HandleFunc("/api/rules/delete", s.handleRuleDelete)
	s.mux.HandleFunc("/api/rules/update", s.handleRuleUpdate)
	s.mux.HandleFunc("/api/rules/reorder", s.handleRuleReorder)

	// API routes for outbounds (HTMX endpoints)
	s.mux.HandleFunc("/api/outbounds", s.handleOutboundsList)
	s.mux.HandleFunc("/api/outbounds/form", s.handleOutboundForm)
	s.mux.HandleFunc("/api/outbounds/create", s.handleOutboundCreate)
	s.mux.HandleFunc("/api/outbounds/update", s.handleOutboundUpdate)
	s.mux.HandleFunc("/api/outbounds/delete", s.handleOutboundDelete)
	s.mux.HandleFunc("/api/outbounds/reorder", s.handleOutboundReorder)
	s.mux.HandleFunc("/api/outbounds/rename", s.handleOutboundRename)
	s.mux.HandleFunc("/api/outbounds/group/manage", s.handleGroupManage)
	s.mux.HandleFunc("/api/outbounds/group/update", s.handleGroupUpdate)

	// API routes for rule actions (HTMX endpoints)
	s.mux.HandleFunc("/api/rule-actions", s.handleRuleActionsList)
	s.mux.HandleFunc("/api/rule-actions/form", s.handleRuleActionForm)
	s.mux.HandleFunc("/api/rule-actions/create", s.handleRuleActionCreate)
	s.mux.HandleFunc("/api/rule-actions/update", s.handleRuleActionUpdate)
	s.mux.HandleFunc("/api/rule-actions/delete", s.handleRuleActionDelete)

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
	s.mux.HandleFunc("/api/config/create-backup", s.handleConfigCreateBackup)

	// WebSocket and API routes for connections
	s.mux.HandleFunc("/ws/connections", s.handleConnectionsWebSocket)
	s.mux.HandleFunc("/api/connections/create-rule", s.handleConnectionToRule)

	// API routes for proxies
	s.mux.HandleFunc("/api/proxies/settings", s.handleProxiesSettings)
	s.mux.HandleFunc("/api/proxies/groups", s.handleProxiesGroups)
	s.mux.HandleFunc("/api/proxies/switch", s.handleProxySwitch)
	s.mux.HandleFunc("/api/proxies/delay-test", s.handleProxyDelayTest)
	s.mux.HandleFunc("/api/proxies/group-delay-test", s.handleProxyGroupDelayTest)

	// API routes for Clash configuration
	s.mux.HandleFunc("/api/clash/config", s.handleClashConfig)
	s.mux.HandleFunc("/api/clash/test", s.handleClashTest)
	s.mux.HandleFunc("/api/clash/update", s.handleClashUpdate)
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
