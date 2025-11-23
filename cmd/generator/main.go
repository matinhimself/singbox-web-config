package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/matinhimself/singbox-web-config/internal/generator"
)

// ConfigCategory represents a category of config types to generate
type ConfigCategory struct {
	Name       string
	FileFilter func(string) bool
	OutputFile string
}

var configCategories = []ConfigCategory{
	{
		Name: "Main",
		FileFilter: generator.FileFilterByNames("options.go"),
		OutputFile: "config.go",
	},
	{
		Name:       "Rules",
		FileFilter: generator.FileFilterByPrefix("rule"),
		OutputFile: "rules.go",
	},
	{
		Name: "DNS",
		FileFilter: generator.FileFilterByNames("dns.go"),
		OutputFile: "dns.go",
	},
	{
		Name: "Inbounds",
		FileFilter: generator.FileFilterByPrefix("inbound"),
		OutputFile: "inbounds.go",
	},
	{
		Name: "Outbounds",
		FileFilter: generator.FileFilterByPrefix("outbound"),
		OutputFile: "outbounds.go",
	},
	{
		Name: "Route",
		FileFilter: generator.FileFilterByNames("route.go", "route_action.go"),
		OutputFile: "route.go",
	},
	{
		Name: "NTP",
		FileFilter: generator.FileFilterByNames("ntp.go"),
		OutputFile: "ntp.go",
	},
	{
		Name: "Experimental",
		FileFilter: generator.FileFilterByNames("experimental.go"),
		OutputFile: "experimental.go",
	},
}

func main() {
	var (
		repoURL    = flag.String("repo", generator.DefaultRepoURL, "Sing-box repository URL")
		branch     = flag.String("branch", generator.DefaultBranch, "Branch to use")
		localPath  = flag.String("local", "", "Use local repository path instead of cloning")
		outputDir  = flag.String("output", "internal/types", "Output directory for generated types")
		skipUpdate = flag.Bool("skip-update", false, "Skip repository update")
		categories = flag.String("categories", "all", "Comma-separated list of categories to generate (all, main, rules, dns, inbounds, outbounds, route, ntp, experimental)")
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

	// Parse requested categories
	requestedCategories := parseCategories(*categories)

	optionPath := repoManager.GetRulePath()
	if _, err := os.Stat(optionPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: option directory not found at %s\n", optionPath)
		fmt.Fprintf(os.Stderr, "Please ensure the sing-box repository is cloned correctly\n")
		os.Exit(1)
	}

	// Generate code
	absOutputDir, err := filepath.Abs(*outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving output directory: %v\n", err)
		os.Exit(1)
	}

	codeGen := generator.NewCodeGenerator(absOutputDir)
	codeGen.Metadata.SingBoxCommit = commit
	codeGen.Metadata.SingBoxBranch = *branch

	totalTypes := 0
	totalFiles := 0

	// Process each category
	for _, category := range requestedCategories {
		fmt.Printf("\n=== Processing %s ===\n", category.Name)
		fmt.Printf("Parsing files from: %s\n", optionPath)

		parser := generator.NewParser(optionPath).WithFileFilter(category.FileFilter)
		files, err := parser.ParseDirectory()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing files for %s: %v\n", category.Name, err)
			continue
		}

		if len(files) == 0 {
			fmt.Printf("No files found for %s, skipping...\n", category.Name)
			continue
		}

		// Extract types
		extractor := generator.NewTypeExtractor(files)
		types, err := extractor.ExtractRuleTypes()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error extracting types for %s: %v\n", category.Name, err)
			continue
		}

		if len(types) == 0 {
			fmt.Printf("No types extracted for %s, skipping...\n", category.Name)
			continue
		}

		// Print extracted types
		fmt.Printf("\nExtracted types for %s:\n", category.Name)
		for _, t := range types {
			if t.IsInterface {
				fmt.Printf("  - %s (interface)\n", t.Name)
			} else {
				fmt.Printf("  - %s (%d fields)\n", t.Name, len(t.Fields))
			}
		}

		// Generate to specific file
		if err := codeGen.GenerateToFile(types, category.OutputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating code for %s: %v\n", category.Name, err)
			continue
		}

		totalTypes += len(types)
		totalFiles += len(files)
	}

	// Generate metadata
	codeGen.Metadata.TypesGenerated = totalTypes
	codeGen.Metadata.FilesProcessed = totalFiles
	if err := codeGen.GenerateMetadata(); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating metadata: %v\n", err)
	}

	fmt.Println()
	fmt.Println("✓ Generation complete!")
	fmt.Printf("  Output: %s\n", absOutputDir)
	fmt.Printf("  Categories: %d\n", len(requestedCategories))
	fmt.Printf("  Types: %d\n", totalTypes)
}

func parseCategories(input string) []ConfigCategory {
	if input == "all" {
		return configCategories
	}

	names := strings.Split(input, ",")
	var result []ConfigCategory

	for _, name := range names {
		name = strings.TrimSpace(strings.ToLower(name))
		for _, cat := range configCategories {
			if strings.ToLower(cat.Name) == name {
				result = append(result, cat)
				break
			}
		}
	}

	if len(result) == 0 {
		return configCategories
	}

	return result
}
