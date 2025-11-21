package forms

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/matinhimself/singbox-web-config/internal/types"
)

// FieldType represents the type of form field
type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypeTextarea FieldType = "textarea"
	FieldTypeNumber   FieldType = "number"
	FieldTypeCheckbox FieldType = "checkbox"
	FieldTypeSelect   FieldType = "select"
	FieldTypeArray    FieldType = "array"
)

// FormField represents a single form field
type FormField struct {
	Name        string
	Label       string
	Type        FieldType
	JSONTag     string
	Placeholder string
	Required    bool
	IsArray     bool
	ArrayType   string // For array fields
	Options     []string
	Description string
}

// FormDefinition represents a complete form
type FormDefinition struct {
	Name   string
	Title  string
	Fields []FormField
}

// Builder builds forms from struct types
type Builder struct{}

// NewBuilder creates a new form builder
func NewBuilder() *Builder {
	return &Builder{}
}

// BuildForm generates a form definition from a rule type
func (b *Builder) BuildForm(ruleTypeName string) (*FormDefinition, error) {
	var value interface{}

	// Map rule type names to their structs
	switch ruleTypeName {
	case "RawDefaultRule":
		value = types.RawDefaultRule{}
	case "RawLogicalRule":
		value = types.RawLogicalRule{}
	case "RawDefaultDNSRule":
		value = types.RawDefaultDNSRule{}
	case "RawLogicalDNSRule":
		value = types.RawLogicalDNSRule{}
	case "LocalRuleSet":
		value = types.LocalRuleSet{}
	case "RemoteRuleSet":
		value = types.RemoteRuleSet{}
	default:
		return nil, fmt.Errorf("unsupported rule type: %s", ruleTypeName)
	}

	t := reflect.TypeOf(value)
	fields := []FormField{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse JSON tag
		jsonName := strings.Split(jsonTag, ",")[0]

		formField := FormField{
			Name:    field.Name,
			Label:   b.fieldNameToLabel(field.Name),
			JSONTag: jsonName,
		}

		// Determine field type and properties
		b.determineFieldType(&formField, field.Type)

		// Add description for common fields
		formField.Description = b.getFieldDescription(field.Name)

		fields = append(fields, formField)
	}

	return &FormDefinition{
		Name:   ruleTypeName,
		Title:  b.typeNameToTitle(ruleTypeName),
		Fields: fields,
	}, nil
}

// determineFieldType determines the appropriate form field type
func (b *Builder) determineFieldType(formField *FormField, t reflect.Type) {
	kind := t.Kind()

	switch kind {
	case reflect.Slice, reflect.Array:
		formField.IsArray = true
		elemType := t.Elem()
		formField.ArrayType = elemType.Kind().String()

		// Determine the element type
		switch elemType.Kind() {
		case reflect.String:
			formField.Type = FieldTypeArray
		case reflect.Uint16, reflect.Int, reflect.Int32:
			formField.Type = FieldTypeArray
			formField.Placeholder = "e.g., 80, 443, 8080"
		default:
			formField.Type = FieldTypeTextarea
		}

	case reflect.String:
		// Check if it's a specific type that should be a select
		if b.isSelectField(formField.Name) {
			formField.Type = FieldTypeSelect
			formField.Options = b.getSelectOptions(formField.Name)
		} else {
			formField.Type = FieldTypeText
		}

	case reflect.Bool:
		formField.Type = FieldTypeCheckbox

	case reflect.Int, reflect.Int32, reflect.Uint16, reflect.Uint32:
		formField.Type = FieldTypeNumber

	case reflect.Ptr:
		// For pointer types, recurse on the element type
		b.determineFieldType(formField, t.Elem())

	default:
		formField.Type = FieldTypeText
	}
}

// isSelectField checks if a field should be a select dropdown
func (b *Builder) isSelectField(fieldName string) bool {
	selectFields := []string{"Mode", "ClashMode", "Strategy"}
	for _, sf := range selectFields {
		if fieldName == sf {
			return true
		}
	}
	return false
}

// getSelectOptions returns options for select fields
func (b *Builder) getSelectOptions(fieldName string) []string {
	switch fieldName {
	case "Mode":
		return []string{"and", "or"}
	case "ClashMode":
		return []string{"direct", "global", "rule"}
	case "Strategy":
		return []string{"prefer_ipv4", "prefer_ipv6", "ipv4_only", "ipv6_only"}
	default:
		return []string{}
	}
}

// fieldNameToLabel converts a field name to a human-readable label
func (b *Builder) fieldNameToLabel(name string) string {
	// Add spaces before capital letters
	var result []rune
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, ' ')
		}
		result = append(result, r)
	}
	return string(result)
}

// typeNameToTitle converts a type name to a title
func (b *Builder) typeNameToTitle(name string) string {
	// Remove "Raw" prefix if present
	name = strings.TrimPrefix(name, "Raw")

	// Add spaces before capital letters
	var result []rune
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, ' ')
		}
		result = append(result, r)
	}
	return string(result)
}

// getFieldDescription returns a description for common fields
func (b *Builder) getFieldDescription(fieldName string) string {
	descriptions := map[string]string{
		"Domain":              "Exact domain names to match (e.g., google.com)",
		"DomainSuffix":        "Domain suffixes to match (e.g., .google.com matches google.com and all subdomains)",
		"DomainKeyword":       "Keywords that must appear in the domain",
		"DomainRegex":         "Regular expressions for domain matching",
		"Geosite":             "Geosite categories (e.g., cn, google, facebook)",
		"GeoIP":               "Country codes for destination IP (e.g., CN, US)",
		"SourceGeoIP":         "Country codes for source IP",
		"IPCIDR":              "IP CIDR ranges for destination (e.g., 192.168.0.0/16)",
		"SourceIPCIDR":        "IP CIDR ranges for source",
		"Port":                "Destination ports to match (e.g., 80, 443)",
		"SourcePort":          "Source ports to match",
		"PortRange":           "Destination port ranges (e.g., 1000:2000)",
		"SourcePortRange":     "Source port ranges",
		"Protocol":            "Network protocols (e.g., tcp, udp)",
		"Network":             "Network types (e.g., tcp, udp)",
		"Inbound":             "Inbound tags to match",
		"Outbound":            "Target outbound for this rule",
		"ProcessName":         "Process names to match",
		"ProcessPath":         "Process paths to match",
		"User":                "User names to match",
		"RuleSet":             "Rule set references",
		"Mode":                "Logical mode: 'and' (all rules must match) or 'or' (any rule must match)",
		"Invert":              "Invert the rule match result",
		"IPIsPrivate":         "Match private IP addresses",
		"SourceIPIsPrivate":   "Match private source IP addresses",
	}

	if desc, ok := descriptions[fieldName]; ok {
		return desc
	}
	return ""
}

// GetAvailableRuleTypes returns all rule types that can have forms
func (b *Builder) GetAvailableRuleTypes() []string {
	return []string{
		"RawDefaultRule",
		"RawLogicalRule",
		"RawDefaultDNSRule",
		"RawLogicalDNSRule",
		"LocalRuleSet",
		"RemoteRuleSet",
	}
}
