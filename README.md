# Sing-Box Web Config Manager

A simple web UI configuration manager for sing-box, starting with route rules management.

## Features

- **Automatic Type Generation**: Stays synchronized with sing-box upstream by automatically generating Go types from the official sing-box repository
- **Web UI**: Clean, responsive interface using HTMX for dynamic interactions
- **Type Safety**: Leverages Go's type system to ensure valid configurations
- **Easy to Use**: Manage complex sing-box route rules without manually editing JSON files

## Project Status

Currently in early development. Completed:
- ✅ Project structure and documentation
- ✅ Type generator from sing-box source
- ✅ Successfully extracted 28 rule-related types with 41 fields in RawDefaultRule
- 🚧 Web server with HTMX (in progress)

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Git

### Generate Types from Sing-Box

```bash
# Generate types from latest sing-box dev-next branch
go run cmd/generator/main.go

# Generate from specific branch
go run cmd/generator/main.go --branch main

# Use local sing-box repository
go run cmd/generator/main.go --local /path/to/sing-box
```

This will:
1. Clone/update the sing-box repository
2. Parse the option package for rule definitions
3. Extract type information
4. Generate Go types in `internal/types/`

### Generated Types

The generator creates types for:
- Default routing rules (`RawDefaultRule` - 41 fields)
- Logical rules (AND/OR)
- DNS rules
- Rule actions
- Rule sets

Example generated type:
```go
type RawDefaultRule struct {
    Inbound       []string `json:"inbound,omitempty"`
    Domain        []string `json:"domain,omitempty"`
    DomainSuffix  []string `json:"domain_suffix,omitempty"`
    GeoIP         []string `json:"geoip,omitempty"`
    Port          []uint16 `json:"port,omitempty"`
    // ... and 36 more fields
}
```

## Architecture

### Type Generator

The generator uses Go's AST parser to extract type definitions from sing-box source:

```
sing-box/option/*.go → Parser → Type Extractor → Code Generator → internal/types/
```

Key features:
- Handles complex generic types (converts `badoption.Listable[string]` to `[]string`)
- Preserves JSON tags and documentation
- Tracks generation metadata (commit hash, timestamp, etc.)

### Project Structure

```
singbox-web-config/
├── cmd/
│   ├── generator/          # Type generator CLI
│   └── server/             # Web server (coming soon)
├── internal/
│   ├── generator/          # Generator logic
│   │   ├── repository.go   # Git operations
│   │   ├── parser.go       # AST parsing
│   │   ├── extractor.go    # Type extraction
│   │   └── generator.go    # Code generation
│   ├── types/              # Generated sing-box types
│   ├── handlers/           # HTTP handlers (coming soon)
│   ├── config/             # Configuration management
│   └── validator/          # Validation logic
├── web/
│   ├── templates/          # HTML templates
│   └── static/             # CSS, JS, images
├── specs/                  # Detailed documentation
│   ├── 01-project-overview.md
│   ├── 02-generator-architecture.md
│   ├── 03-web-ui-architecture.md
│   ├── 04-singbox-route-rules.md
│   └── 05-implementation-plan.md
├── examples/               # Example configurations
└── testdata/              # Test fixtures
```

## Documentation

See the `specs/` directory for detailed documentation:

- [Project Overview](specs/01-project-overview.md) - Goals, tech stack, and principles
- [Generator Architecture](specs/02-generator-architecture.md) - How the type generator works
- [Web UI Architecture](specs/03-web-ui-architecture.md) - HTMX-based UI design
- [Sing-Box Route Rules](specs/04-singbox-route-rules.md) - Understanding route rules
- [Implementation Plan](specs/05-implementation-plan.md) - Roadmap and milestones

## Technology Stack

- **Backend**: Go with standard library HTTP server
- **Frontend**: HTMX for dynamic interactions
- **Templates**: Go html/template
- **Type Generation**: Go AST parser

## Next Steps

1. Implement basic web server with HTMX
2. Create rule management UI (CRUD operations)
3. Add configuration export functionality
4. Implement validation
5. Add import functionality

## Contributing

This is an early-stage project. Contributions, ideas, and feedback are welcome!

## License

TBD

## Acknowledgments

- [Sing-Box](https://github.com/SagerNet/sing-box) - The amazing proxy platform this project builds upon
- [HTMX](https://htmx.org/) - High power tools for HTML
