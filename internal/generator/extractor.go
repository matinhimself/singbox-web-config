package generator

import (
	"fmt"
	"go/ast"
	"strings"
)

// RuleType represents an extracted rule type definition
type RuleType struct {
	Name        string
	Doc         string
	Fields      []*Field
	SourceFile  string
	IsInterface bool
}

// Field represents a struct field
type Field struct {
	Name     string
	Type     string
	JSONTag  string
	Doc      string
	Required bool
}

// TypeExtractor extracts type information from parsed AST files
type TypeExtractor struct {
	files map[string]*ast.File
}

// NewTypeExtractor creates a new type extractor
func NewTypeExtractor(files map[string]*ast.File) *TypeExtractor {
	return &TypeExtractor{
		files: files,
	}
}

// ExtractRuleTypes extracts all rule type definitions
func (e *TypeExtractor) ExtractRuleTypes() ([]*RuleType, error) {
	var ruleTypes []*RuleType

	for fileName, file := range e.files {
		types := e.extractTypesFromFile(fileName, file)
		ruleTypes = append(ruleTypes, types...)
	}

	if len(ruleTypes) == 0 {
		return nil, fmt.Errorf("no types found")
	}

	fmt.Printf("Extracted %d types\n", len(ruleTypes))
	return ruleTypes, nil
}

// extractTypesFromFile extracts types from a single file
func (e *TypeExtractor) extractTypesFromFile(fileName string, file *ast.File) []*RuleType {
	var types []*RuleType

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			ruleType := e.extractType(fileName, typeSpec, genDecl.Doc)
			if ruleType != nil {
				types = append(types, ruleType)
			}
		}
	}

	return types
}

// extractType extracts a single type definition
func (e *TypeExtractor) extractType(fileName string, typeSpec *ast.TypeSpec, doc *ast.CommentGroup) *RuleType {
	typeName := typeSpec.Name.Name

	// Skip unexported types
	if !ast.IsExported(typeName) {
		return nil
	}

	// Skip types that are just aliases to external types
	if ident, ok := typeSpec.Type.(*ast.Ident); ok {
		if !ast.IsExported(ident.Name) {
			return nil
		}
	}

	// Check if it's a struct
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		// Check if it's an interface
		if _, isInterface := typeSpec.Type.(*ast.InterfaceType); isInterface {
			return &RuleType{
				Name:        typeName,
				Doc:         extractDoc(doc),
				SourceFile:  fileName,
				IsInterface: true,
			}
		}
		return nil
	}

	// Extract fields
	fields := e.extractFields(structType)

	return &RuleType{
		Name:       typeName,
		Doc:        extractDoc(doc),
		Fields:     fields,
		SourceFile: fileName,
	}
}

// extractFields extracts fields from a struct type
func (e *TypeExtractor) extractFields(structType *ast.StructType) []*Field {
	var fields []*Field

	for _, field := range structType.Fields.List {
		// Skip embedded fields
		if len(field.Names) == 0 {
			continue
		}

		for _, name := range field.Names {
			// Skip unexported fields
			if !ast.IsExported(name.Name) {
				continue
			}

			f := &Field{
				Name: name.Name,
				Type: e.typeToString(field.Type),
				Doc:  extractDoc(field.Doc),
			}

			// Extract JSON tag
			if field.Tag != nil {
				tag := field.Tag.Value
				f.JSONTag = extractJSONTag(tag)
				f.Required = !strings.Contains(tag, "omitempty")

				// Skip fields with json:"-" (internal implementation fields)
				if f.JSONTag == "-" {
					continue
				}
			}

			fields = append(fields, f)
		}
	}

	return fields
}

// typeToString converts an AST type expression to a string
func (e *TypeExtractor) typeToString(expr ast.Expr) string {
	typeStr := e.typeToStringRaw(expr)
	return e.simplifyType(typeStr)
}

// typeToStringRaw converts an AST type expression to a raw string
func (e *TypeExtractor) typeToStringRaw(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.ArrayType:
		return "[]" + e.typeToStringRaw(t.Elt)
	case *ast.StarExpr:
		return "*" + e.typeToStringRaw(t.X)
	case *ast.SelectorExpr:
		return e.typeToStringRaw(t.X) + "." + t.Sel.Name
	case *ast.MapType:
		return "map[" + e.typeToStringRaw(t.Key) + "]" + e.typeToStringRaw(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.IndexExpr:
		// Handle generic types like Listable[string]
		base := e.typeToStringRaw(t.X)
		index := e.typeToStringRaw(t.Index)
		return base + "[" + index + "]"
	case *ast.IndexListExpr:
		// Handle multiple type parameters
		base := e.typeToStringRaw(t.X)
		var indices []string
		for _, idx := range t.Indices {
			indices = append(indices, e.typeToStringRaw(idx))
		}
		return base + "[" + strings.Join(indices, ",") + "]"
	default:
		return "unknown"
	}
}

// simplifyType simplifies complex sing-box types to basic Go types
func (e *TypeExtractor) simplifyType(typeStr string) string {
	// Map sing-box specific types to simpler equivalents

	// badoption.Listable[T] -> []T
	if strings.HasPrefix(typeStr, "badoption.Listable[") {
		inner := strings.TrimPrefix(typeStr, "badoption.Listable[")
		inner = strings.TrimSuffix(inner, "]")
		return "[]" + e.simplifyType(inner)
	}

	// badoption.Prefixable -> string (for IP prefixes)
	if typeStr == "badoption.Prefixable" || typeStr == "*badoption.Prefixable" {
		if strings.HasPrefix(typeStr, "*") {
			return "*string"
		}
		return "string"
	}

	// badoption.Duration -> uint32 (milliseconds)
	if typeStr == "badoption.Duration" {
		return "uint32"
	}

	// badjson.TypedMap[K,V] -> map[string]interface{}
	if strings.HasPrefix(typeStr, "badjson.TypedMap[") || strings.HasPrefix(typeStr, "*badjson.TypedMap[") {
		if strings.HasPrefix(typeStr, "*") {
			return "*map[string]interface{}"
		}
		return "map[string]interface{}"
	}

	// Handle external package types - replace with interface{} or appropriate basic type
	if strings.Contains(typeStr, ".") {
		// Check for known external packages
		if strings.HasPrefix(typeStr, "domain.") ||
		   strings.HasPrefix(typeStr, "netipx.") ||
		   strings.HasPrefix(typeStr, "json.") ||
		   strings.HasPrefix(typeStr, "badjson.") {
			// Replace with interface{} to avoid import issues
			if strings.HasPrefix(typeStr, "*") {
				return "*interface{}"
			}
			return "interface{}"
		}
	}

	// Remove package prefixes for known types
	typeStr = strings.ReplaceAll(typeStr, "option.", "")

	// Handle remaining complex generic types with "bad" prefix
	if strings.HasPrefix(typeStr, "bad") && strings.Contains(typeStr, "[") {
		return "interface{}"
	}

	// Handle specific sing-box option types
	knownTypes := map[string]string{
		"InterfaceType":          "string",
		"NetworkStrategy":        "string",
		"*NetworkStrategy":       "*string",
		"DomainStrategy":         "string",
		"DNSQueryType":           "string",
		"DNSRCode":               "uint16",
		"*DNSRCode":              "*uint16",
		"DNSRecordOptions":       "interface{}",
		"Rule":                   "interface{}",
		"[]Rule":                 "[]interface{}",
		"DNSRule":                "interface{}",
		"[]DNSRule":              "[]interface{}",
		"HeadlessRule":           "interface{}",
		"[]HeadlessRule":         "[]interface{}",
		"DNSServerOptions":       "interface{}",
		"[]DNSServerOptions":     "[]interface{}",
		"badPrefix":              "string",
		"*badPrefix":             "*string",
		"badAddr":                "string",
		"*badAddr":               "*string",
		"badHTTPHeader":          "map[string]string",
		"*badHTTPHeader":         "*map[string]string",
		"FwMark":                 "uint32",
		"UDPTimeoutCompat":       "uint32",
		"DebugOptions":           "interface{}",
		"*DebugOptions":          "*interface{}",
		"DomainResolveOptions":   "interface{}",
		"*DomainResolveOptions":  "*interface{}",
		"Inbound":                "interface{}",
		"[]Inbound":              "[]interface{}",
		"Outbound":               "interface{}",
		"[]Outbound":             "[]interface{}",
		"Endpoint":               "interface{}",
		"[]Endpoint":             "[]interface{}",
		"Service":                "interface{}",
		"[]Service":              "[]interface{}",
		"RuleSet":                "interface{}",
		"[]RuleSet":              "[]interface{}",
	}

	if replacement, ok := knownTypes[typeStr]; ok {
		return replacement
	}

	// Replace unknown types with interface{}
	if typeStr == "unknown" {
		return "interface{}"
	}

	return typeStr
}

// extractDoc extracts documentation from comment group
func extractDoc(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}

	var lines []string
	for _, comment := range cg.List {
		text := comment.Text
		// Remove // or /*  */
		text = strings.TrimPrefix(text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)
		if text != "" {
			lines = append(lines, text)
		}
	}

	return strings.Join(lines, " ")
}

// extractJSONTag extracts the JSON tag name from a struct tag
func extractJSONTag(tag string) string {
	// Remove backticks
	tag = strings.Trim(tag, "`")

	// Find json:"name"
	parts := strings.Split(tag, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, "json:") {
			// Extract value between quotes
			jsonTag := strings.TrimPrefix(part, "json:")
			jsonTag = strings.Trim(jsonTag, "\"")
			// Remove omitempty and other options
			jsonTag = strings.Split(jsonTag, ",")[0]
			return jsonTag
		}
	}

	return ""
}
