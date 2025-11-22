package handlers

import (
	"encoding/json"
	"html/template"
)

// Template helper functions

// FuncMap returns custom template functions
func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"add":        add,
		"marshal":    marshal,
		"derefString": derefString,
		"derefUint32": derefUint32,
		"strPtrEq":   strPtrEq,
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

// derefString safely dereferences a *string pointer, returning empty string if nil
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// derefUint32 safely dereferences a *uint32 pointer, returning 0 if nil
func derefUint32(u *uint32) uint32 {
	if u == nil {
		return 0
	}
	return *u
}

// strPtrEq compares a *string pointer with a string value
// Returns true if the pointer is not nil and the dereferenced value equals the string
func strPtrEq(s *string, val string) bool {
	if s == nil {
		return false
	}
	return *s == val
}
