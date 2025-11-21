package handlers

import (
	"encoding/json"
	"html/template"
)

// Template helper functions

// FuncMap returns custom template functions
func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"add":     add,
		"marshal": marshal,
	}
}

// add adds two integers (for template indexing)
func add(a, b int) int {
	return a + b
}

// marshal converts an interface to JSON string for display
func marshal(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}
