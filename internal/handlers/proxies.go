package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// ProxyGroupData represents a proxy group with its members
type ProxyGroupData struct {
	Name      string
	Type      string
	Now       string
	Proxies   []ProxyNodeData
	CanSwitch bool
}

// ProxyNodeData represents a proxy node
type ProxyNodeData struct {
	Name  string
	Type  string
	Delay int
	IsNow bool
}

// handleProxiesPage handles the proxies management page
func (s *Server) handleProxiesPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Proxy Management",
		Data: map[string]interface{}{
			"ClashURL":    s.clashURL,
			"ClashSecret": s.clashSecret,
		},
	}

	if err := s.renderTemplate(w, "proxies.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleProxiesSettings displays current Clash API settings (read-only)
// Settings are configured via command-line arguments only
func (s *Server) handleProxiesSettings(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"ClashURL":       s.clashURL,
		"ClashSecret":    s.clashSecret,
		"HasClashClient": s.clashClient != nil,
	}

	if err := s.renderTemplate(w, "proxy-settings.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleProxiesGroups handles fetching all proxy groups
func (s *Server) handleProxiesGroups(w http.ResponseWriter, r *http.Request) {
	if s.clashClient == nil {
		http.Error(w, "Clash API not configured", http.StatusBadRequest)
		return
	}

	proxies, err := s.clashClient.GetProxies()
	if err != nil {
		log.Printf("Error fetching proxies: %v", err)
		http.Error(w, "Failed to fetch proxies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Process proxy groups
	var groups []ProxyGroupData
	for name, proxy := range proxies {
		if len(proxy.All) > 0 {
			group := ProxyGroupData{
				Name:      name,
				Type:      proxy.Type,
				Now:       proxy.Now,
				CanSwitch: proxy.Type == "Selector" || proxy.Type == "URLTest" || proxy.Type == "Fallback",
			}

			// Get proxy nodes for this group
			for _, proxyName := range proxy.All {
				node := ProxyNodeData{
					Name:  proxyName,
					IsNow: proxyName == proxy.Now,
				}

				// Try to get delay from history
				if proxyNode, ok := proxies[proxyName]; ok {
					node.Type = proxyNode.Type
					if len(proxyNode.History) > 0 {
						node.Delay = proxyNode.History[len(proxyNode.History)-1].Delay
					}
				}

				group.Proxies = append(group.Proxies, node)
			}

			groups = append(groups, group)
		}
	}

	data := map[string]interface{}{
		"Groups": groups,
	}

	if err := s.renderTemplate(w, "proxy-groups.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleProxySwitch handles switching the active proxy in a group
func (s *Server) handleProxySwitch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.clashClient == nil {
		http.Error(w, "Clash API not configured", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	groupName := r.FormValue("group")
	proxyName := r.FormValue("proxy")

	if groupName == "" || proxyName == "" {
		http.Error(w, "Group and proxy names are required", http.StatusBadRequest)
		return
	}

	if err := s.clashClient.SwitchProxy(groupName, proxyName); err != nil {
		log.Printf("Error switching proxy: %v", err)
		http.Error(w, "Failed to switch proxy: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated proxy groups
	w.Header().Set("HX-Trigger", "proxySwitched")
	s.handleProxiesGroups(w, r)
}

// handleProxyDelayTest handles testing proxy delay
func (s *Server) handleProxyDelayTest(w http.ResponseWriter, r *http.Request) {
	if s.clashClient == nil {
		http.Error(w, "Clash API not configured", http.StatusBadRequest)
		return
	}

	proxyName := r.URL.Query().Get("name")
	if proxyName == "" {
		http.Error(w, "Proxy name is required", http.StatusBadRequest)
		return
	}

	testURL := r.URL.Query().Get("url")
	timeoutStr := r.URL.Query().Get("timeout")
	timeout := 5000
	if timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = t
		}
	}

	delay, err := s.clashClient.TestProxyDelay(proxyName, testURL, timeout)
	if err != nil {
		// Return error but don't fail completely
		response := map[string]interface{}{
			"name":    proxyName,
			"delay":   0,
			"error":   err.Error(),
			"timeout": true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"name":  proxyName,
		"delay": delay,
		"error": nil,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleProxyGroupDelayTest handles testing all proxies in a group
func (s *Server) handleProxyGroupDelayTest(w http.ResponseWriter, r *http.Request) {
	if s.clashClient == nil {
		http.Error(w, "Clash API not configured", http.StatusBadRequest)
		return
	}

	groupName := r.URL.Query().Get("group")
	if groupName == "" {
		http.Error(w, "Group name is required", http.StatusBadRequest)
		return
	}

	proxy, err := s.clashClient.GetProxy(groupName)
	if err != nil {
		http.Error(w, "Failed to get proxy group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	testURL := r.URL.Query().Get("url")
	timeoutStr := r.URL.Query().Get("timeout")
	timeout := 5000
	if timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = t
		}
	}

	results := make([]map[string]interface{}, 0)
	for _, proxyName := range proxy.All {
		delay, err := s.clashClient.TestProxyDelay(proxyName, testURL, timeout)
		result := map[string]interface{}{
			"name": proxyName,
		}
		if err != nil {
			result["delay"] = 0
			result["error"] = err.Error()
			result["timeout"] = strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "context deadline exceeded")
		} else {
			result["delay"] = delay
		}
		results = append(results, result)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"group":   groupName,
		"results": results,
	})
}
