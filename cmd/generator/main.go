package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/matinhimself/singbox-web-config/internal/generator"
)

func main() {
	var (
		repoURL    = flag.String("repo", generator.DefaultRepoURL, "Sing-box repository URL")
		branch     = flag.String("branch", generator.DefaultBranch, "Branch to use")
		localPath  = flag.String("local", "", "Use local repository path instead of cloning")
		outputDir  = flag.String("output", "internal/types", "Output directory for generated types")
		skipUpdate = flag.Bool("skip-update", false, "Skip repository update")
	)

	flag.Parse()

	fmt.Println("Sing-Box Type Generator")
	fmt.Println("=======================")
	fmt.Println()

	// Setup repository manager
	repoManager := generator.NewRepositoryManager().
		WithRepoURL(*repoURL).
		WithBranch(*branch)

	if *localPath != "" {
		repoManager.WithLocalPath(*localPath)
		fmt.Printf("Using local repository: %s\n", *localPath)
	} else {
		fmt.Printf("Repository: %s\n", *repoURL)
		fmt.Printf("Branch: %s\n", *branch)
	}

	// Update repository
	if !*skipUpdate {
		if err := repoManager.Update(); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating repository: %v\n", err)
			os.Exit(1)
		}
	}

	// Get repository info
	commit, err := repoManager.GetCommitHash()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to get commit hash: %v\n", err)
		commit = "unknown"
	}

	fmt.Printf("Commit: %s\n", commit)
	fmt.Println()

	// Parse source files
	rulePath := repoManager.GetRulePath()
	if _, err := os.Stat(rulePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: route/rule directory not found at %s\n", rulePath)
		fmt.Fprintf(os.Stderr, "Please ensure the sing-box repository is cloned correctly\n")
		os.Exit(1)
	}

	fmt.Printf("Parsing files from: %s\n", rulePath)
	parser := generator.NewParser(rulePath)
	files, err := parser.ParseDirectory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing files: %v\n", err)
		os.Exit(1)
	}

	// Extract types
	extractor := generator.NewTypeExtractor(files)
	types, err := extractor.ExtractRuleTypes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting types: %v\n", err)
		os.Exit(1)
	}

	// Print extracted types
	fmt.Println("\nExtracted types:")
	for _, t := range types {
		if t.IsInterface {
			fmt.Printf("  - %s (interface)\n", t.Name)
		} else {
			fmt.Printf("  - %s (%d fields)\n", t.Name, len(t.Fields))
		}
	}
	fmt.Println()

	// Generate code
	absOutputDir, err := filepath.Abs(*outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving output directory: %v\n", err)
		os.Exit(1)
	}

	codeGen := generator.NewCodeGenerator(absOutputDir)
	codeGen.Metadata.SingBoxCommit = commit
	codeGen.Metadata.SingBoxBranch = *branch
	codeGen.Metadata.FilesProcessed = len(files)

	if err := codeGen.Generate(types); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✓ Generation complete!")
	fmt.Printf("  Output: %s\n", absOutputDir)
	fmt.Printf("  Types: %d\n", len(types))
}
