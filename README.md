# Sing-Box Web Config Manager

A fully functional web UI configuration manager for sing-box with comprehensive route rules management, service control, and automatic backup capabilities.

## Features

### Core Capabilities

- **Complete Rules Management**: Full CRUD operations for sing-box routing rules with drag-and-drop reordering
- **Service Integration**: Start, stop, restart sing-box service directly from the web UI
- **Automatic Backups**: Every configuration change creates a backup with metadata tracking
- **Live Monitoring**: Watch service status and logs in real-time
- **Type Generation**: Automatically generates Go types from sing-box upstream to stay synchronized
- **Web UI**: Clean, responsive HTMX-powered interface without heavy JavaScript
- **Embedded Deployment**: Single binary with all templates and assets embedded
- **File Watching**: Detects external configuration changes with debouncing

### Rules Management

- **6 Supported Rule Types**: Default, Logical (AND/OR), DNS, DNS Logical, Local RuleSet, Remote RuleSet
- **41+ Rule Fields**: Support for all sing-box routing rule fields
- **Dynamic Forms**: Intelligent form generation with field type detection
- **Visual Ordering**: Drag-and-drop interface for rule priority
- **JSON Preview**: View rule configuration before saving
- **Smart Validation**: Form validation with type checking

### Service Management

- **Service Control**: Start, stop, restart sing-box service via systemd
- **Status Monitoring**: Real-time service status with auto-refresh
- **Log Viewer**: View recent service logs with configurable line counts
- **Auto-reload**: Automatically reloads service after configuration changes

### Configuration Management

- **Backup System**: Automatic and manual backups with descriptions
- **Backup Metadata**: Track backup name, description, timestamp, and version
- **Restore**: Restore any previous configuration (creates backup before restore)
- **Export**: Download current configuration as JSON
- **Import**: Restore configurations from backup files

## Project Status

**Production Ready** - All core features implemented and tested:
- ✅ Type generator from sing-box source (19 types extracted)
- ✅ Full web server with HTMX
- ✅ Complete rules CRUD with 6 rule types
- ✅ Service management and monitoring
- ✅ Backup and restore functionality
- ✅ Embedded assets for portable deployment
- ✅ Debian package builds via GitHub Actions

## Quick Start

### Installation

#### Option 1: Debian Package (Recommended)

Download the latest `.deb` package from [Releases](https://github.com/matinhimself/singbox-web-config/releases):

```bash
# Install the package
sudo dpkg -i singbox-web-config_*.deb

# Enable and start the service
sudo systemctl enable singbox-web-config
sudo systemctl start singbox-web-config

# Access the web UI at http://localhost:8080
```

The Debian package:
- Installs to `/usr/local/bin/singbox-web-config`
- Creates systemd service at `/etc/systemd/system/singbox-web-config.service`
- Runs on port 8080 by default
- Manages `/etc/sing-box/config.json` by default

#### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/matinhimself/singbox-web-config.git
cd singbox-web-config

# Build the server
go build -o singbox-web-config cmd/server/main.go

# Run the server
./singbox-web-config --addr localhost:8080 --config /etc/sing-box/config.json
```

### Usage

Once running, access the web UI at `http://localhost:8080`:

1. **Dashboard** - View configuration metadata and system status
2. **Rules** - Manage routing rules with full CRUD operations
3. **Service** - Control sing-box service and view logs

#### Command-line Options

```bash
singbox-web-config [options]

Options:
  --addr string       HTTP server address (default "localhost:8080")
  --config string     Path to sing-box config file (default "/etc/sing-box/config.json")
  --service string    Name of sing-box systemd service (default "sing-box")
```

### Type Generator

The type generator keeps the project synchronized with sing-box upstream:

```bash
# Generate types from latest sing-box dev-next branch
go run cmd/generator/main.go

# Generate from specific branch
go run cmd/generator/main.go --branch main

# Use local sing-box repository
go run cmd/generator/main.go --local /path/to/sing-box

# Skip repository update (use existing clone)
go run cmd/generator/main.go --skip-update
```

This generates:
- 19 rule-related types from sing-box source
- `internal/types/rules.go` - Struct definitions with JSON tags
- `internal/types/metadata.go` - Generation metadata (commit, timestamp, etc.)

### Generated Types

The generator extracts and converts sing-box types:

| Type | Fields | Description |
|------|--------|-------------|
| `RawDefaultRule` | 41 | Standard routing rules with domain, IP, port, protocol matching |
| `RawLogicalRule` | 3 | Compound rules with AND/OR logic |
| `RawDefaultDNSRule` | 20+ | DNS-specific routing rules |
| `LocalRuleSet` | 5 | File-based rule sets |
| `RemoteRuleSet` | 6 | URL-based rule sets |

Example:
```go
type RawDefaultRule struct {
    Inbound       []string `json:"inbound,omitempty"`
    Domain        []string `json:"domain,omitempty"`
    DomainSuffix  []string `json:"domain_suffix,omitempty"`
    GeoIP         []string `json:"geoip,omitempty"`
    Port          []uint16 `json:"port,omitempty"`
    Outbound      string   `json:"outbound,omitempty"`
    // ... and 35 more fields
}
```

## Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        Web Browser                          │
│  (HTMX + Vanilla JS for drag-and-drop & dynamic forms)    │
└────────────────┬────────────────────────────────────────────┘
                 │ HTTP
                 ▼
┌─────────────────────────────────────────────────────────────┐
│                     Go HTTP Server                          │
│  ┌──────────────┬─────────────┬──────────────────────────┐ │
│  │   Handlers   │   Forms     │      Config Manager      │ │
│  │  (Routing)   │  (Builder)  │  (Read/Write/Backup)     │ │
│  └──────────────┴─────────────┴──────────────────────────┘ │
└────────┬──────────────────────────────────┬────────────────┘
         │                                   │
         ▼                                   ▼
┌──────────────────┐              ┌─────────────────────────┐
│  Service Manager │              │  File System            │
│  (systemctl)     │              │  - config.json          │
│  - Start/Stop    │              │  - backups/             │
│  - Logs          │              │  - .meta files          │
└──────────────────┘              └─────────────────────────┘
         │
         ▼
┌──────────────────┐
│   Sing-Box       │
│   Service        │
└──────────────────┘
```

### Type Generator Pipeline

```
sing-box/option/*.go
         │
         ▼
    Go AST Parser ──→ Type Extractor ──→ Code Generator
         │                   │                  │
         │                   │                  ▼
         │                   │         internal/types/rules.go
         │                   │         internal/types/metadata.go
         │                   │
         │                   └──→ Tracks: commit hash, timestamp,
         │                         branch, type count
         │
         └──→ Handles generics: badoption.Listable[T] → []T
```

### Project Structure

```
singbox-web-config/
├── cmd/
│   ├── generator/          # Type generator CLI
│   └── server/             # Web server application
├── internal/
│   ├── generator/          # Type generation pipeline
│   │   ├── repository.go   # Git operations (clone, checkout, update)
│   │   ├── parser.go       # Go AST parsing
│   │   ├── extractor.go    # Type extraction from AST
│   │   └── generator.go    # Code generation with templates
│   ├── types/              # Generated sing-box types
│   │   ├── rules.go        # Rule type definitions (auto-generated)
│   │   └── metadata.go     # Generation metadata (auto-generated)
│   ├── handlers/           # HTTP handlers and routing
│   │   ├── server.go       # Server setup and lifecycle
│   │   ├── routes.go       # Route definitions
│   │   ├── rules.go        # Rules CRUD handlers
│   │   ├── service.go      # Service management handlers
│   │   └── config.go       # Config/backup handlers
│   ├── config/             # Configuration management
│   │   ├── manager.go      # Config file operations
│   │   └── backup.go       # Backup system
│   ├── service/            # Systemd service control
│   │   └── manager.go      # Service operations
│   ├── forms/              # Dynamic form generation
│   │   └── builder.go      # Reflection-based form builder
│   └── watcher/            # File change detection
│       └── watcher.go      # fsnotify-based file watcher
├── assets/                 # Embedded resources
│   ├── templates.go        # Embedded templates
│   └── static.go           # Embedded static files
├── web/
│   ├── templates/          # HTML templates (embedded)
│   │   ├── base.html       # Base layout
│   │   ├── index.html      # Dashboard
│   │   ├── rules.html      # Rules page
│   │   ├── service.html    # Service page
│   │   ├── rule-form.html  # Rule form modal
│   │   ├── rule-list.html  # Rules list with drag-and-drop
│   │   └── *.html          # Other components
│   └── static/             # CSS and assets (embedded)
│       ├── css/
│       └── images/
├── specs/                  # Detailed documentation
│   ├── 01-project-overview.md
│   ├── 02-generator-architecture.md
│   ├── 03-web-ui-architecture.md
│   ├── 04-singbox-route-rules.md
│   └── 05-implementation-plan.md
├── deploy/                 # Deployment configurations
│   └── debian/             # Debian package files
├── .github/
│   └── workflows/          # CI/CD pipelines
│       └── build-deb.yml   # Debian package build
└── testdata/               # Test fixtures
```

## Documentation

See the `specs/` directory for detailed documentation:

- [Project Overview](specs/01-project-overview.md) - Goals, tech stack, and principles
- [Generator Architecture](specs/02-generator-architecture.md) - How the type generator works
- [Web UI Architecture](specs/03-web-ui-architecture.md) - HTMX-based UI design
- [Sing-Box Route Rules](specs/04-singbox-route-rules.md) - Understanding route rules
- [Implementation Plan](specs/05-implementation-plan.md) - Roadmap and milestones

## Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Backend** | Go 1.21+ | Type-safe, high-performance server |
| **HTTP Server** | Go net/http | Standard library HTTP server |
| **Frontend** | HTMX 1.9.10 | Dynamic interactions without heavy JavaScript |
| **Templates** | html/template | Server-side rendering |
| **Type Generation** | go/ast, go/parser | AST parsing for type extraction |
| **File Watching** | fsnotify | Detect external config changes |
| **Service Control** | systemctl/journalctl | Linux systemd integration |
| **Styling** | Custom CSS | Clean, responsive design |
| **Packaging** | Debian .deb | Easy installation on Debian/Ubuntu |
| **CI/CD** | GitHub Actions | Automated builds and releases |

## Key Design Decisions

- **No Heavy Dependencies**: Uses Go standard library wherever possible
- **Embedded Assets**: Single binary deployment with embedded templates and static files
- **HTMX over SPA**: Server-side rendering with HTMX for simplicity and performance
- **Automatic Backups**: Every config change creates a backup for safety
- **Type Safety**: Generated types from sing-box source ensure compatibility
- **Systemd Integration**: Native Linux service management

## Roadmap

See [ROADMAP.md](ROADMAP.md) for detailed future improvements and planned features.

### Completed Features ✅

- Full CRUD for routing rules (6 rule types)
- Service management (start, stop, restart, logs)
- Automatic backup system with metadata
- Drag-and-drop rule reordering
- Real-time service monitoring
- File change detection
- Embedded deployment
- Debian package builds

### Coming Soon 🚧

- DNS rules management interface
- Inbound/Outbound configuration UI
- Rule templates and presets
- Advanced search and filtering
- Configuration validation and testing
- Multi-configuration support

## Contributing

Contributions, issues, and feature requests are welcome!

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) file for details

## Acknowledgments

- [Sing-Box](https://github.com/SagerNet/sing-box) - The amazing proxy platform this project builds upon
- [HTMX](https://htmx.org/) - High power tools for HTML
