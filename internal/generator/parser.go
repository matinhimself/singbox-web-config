package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Parser handles parsing Go source files
type Parser struct {
	SourceDir   string
	FileFilter  func(string) bool // Optional filter function for file names
	fset        *token.FileSet
}

// NewParser creates a new parser for the given directory
func NewParser(sourceDir string) *Parser {
	return &Parser{
		SourceDir: sourceDir,
		fset:      token.NewFileSet(),
	}
}

// ParseDirectory parses all Go files in the directory
func (p *Parser) ParseDirectory() (map[string]*ast.File, error) {
	files := make(map[string]*ast.File)

	entries, err := os.ReadDir(p.SourceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		// Skip test files
		if strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		// Apply file filter if provided
		if p.FileFilter != nil && !p.FileFilter(entry.Name()) {
			continue
		}

		filePath := filepath.Join(p.SourceDir, entry.Name())
		astFile, err := parser.ParseFile(p.fset, filePath, nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("Warning: failed to parse %s: %v\n", entry.Name(), err)
			continue
		}

		files[entry.Name()] = astFile
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no Go files found in %s", p.SourceDir)
	}

	fmt.Printf("Successfully parsed %d Go files\n", len(files))
	return files, nil
}

// WithFileFilter sets a custom file filter
func (p *Parser) WithFileFilter(filter func(string) bool) *Parser {
	p.FileFilter = filter
	return p
}

// GetFileSet returns the file set used for parsing
func (p *Parser) GetFileSet() *token.FileSet {
	return p.fset
}

// FileFilterByPrefix creates a filter that matches files with the given prefix
func FileFilterByPrefix(prefix string) func(string) bool {
	return func(name string) bool {
		return strings.HasPrefix(name, prefix)
	}
}

// FileFilterByNames creates a filter that matches specific filenames
func FileFilterByNames(names ...string) func(string) bool {
	nameSet := make(map[string]bool)
	for _, name := range names {
		nameSet[name] = true
	}
	return func(name string) bool {
		return nameSet[name]
	}
}
