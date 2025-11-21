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
	SourceDir string
	fset      *token.FileSet
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

		// Only parse rule-related files
		if !strings.HasPrefix(entry.Name(), "rule") {
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

// GetFileSet returns the file set used for parsing
func (p *Parser) GetFileSet() *token.FileSet {
	return p.fset
}
