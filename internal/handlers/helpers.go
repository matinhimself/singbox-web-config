package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
)

// Template helper functions

// FuncMap returns custom template functions
func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"add":         add,
		"marshal":     marshal,
		"derefString": derefString,
		"derefUint32": derefUint32,
		"strPtrEq":    strPtrEq,
		"dict":        dict,
		"list":        list,
		"has":         has,
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

// dict creates a map from alternating key-value pairs
// Example: dict "key1" "value1" "key2" "value2" -> map[string]interface{}{"key1": "value1", "key2": "value2"}
func dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict requires an even number of arguments")
	}

	result := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings")
		}
		result[key] = values[i+1]
	}

	return result, nil
}

// list creates a slice from the given arguments
// Example: list "item1" "item2" "item3" -> []string{"item1", "item2", "item3"}
func list(values ...string) []string {
	return values
}

// has checks if a value exists in a slice
// Example: has "needle" (list "hay" "needle" "stack") -> true
func has(value string, slice []string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
