# Sing-Box Web Config Manager - Project Overview

## Project Goal

Create a simple, user-friendly web UI configuration manager for sing-box, starting with route rules configuration. The project will provide an intuitive interface for managing complex sing-box configurations without manually editing JSON files.

## Phase 1: Route Rules Manager

Initial focus is on managing `route.rules` configurations, which are one of the most complex and frequently modified parts of sing-box configuration.

### Key Features

1. **Automatic Type Generation**: Stay synchronized with sing-box upstream by automatically generating Go types from the official sing-box repository
2. **Web UI**: Clean, responsive interface using HTMX for dynamic interactions without heavy JavaScript
3. **Type Safety**: Leverage Go's type system to ensure valid configurations
4. **Validation**: Built-in validation based on sing-box's actual rules
5. **Export**: Generate valid JSON configuration files for sing-box

## Technology Stack

### Backend
- **Go**: Strong typing, excellent performance, native sing-box compatibility
- **Standard library HTTP server**: Simple, reliable, no heavy frameworks needed initially
- **Go templates**: Server-side rendering for initial page loads

### Frontend
- **HTMX**: Modern approach to dynamic web pages without complex JavaScript
- **Pure CSS/Tailwind**: Clean, maintainable styling
- **HTML templates**: Go's html/template package for type-safe rendering

## Project Structure

```
singbox-web-config/
├── cmd/
│   ├── generator/          # Type generator tool
│   └── server/             # Web server application
├── internal/
│   ├── generator/          # Generator logic
│   ├── types/              # Generated sing-box types
│   ├── handlers/           # HTTP request handlers
│   ├── config/             # Configuration management
│   └── validator/          # Configuration validation
├── web/
│   ├── templates/          # HTML templates
│   └── static/             # CSS, JS, images
├── specs/                  # Project documentation
├── examples/               # Example configurations
└── testdata/              # Test fixtures
```

## Development Principles

1. **Upstream First**: Always use sing-box source code as the source of truth
2. **Simplicity**: Start simple, add complexity only when needed
3. **Type Safety**: Leverage Go's type system for correctness
4. **User Experience**: Prioritize ease of use over feature completeness
5. **Maintainability**: Clear code structure, good documentation

## Future Phases

- Phase 2: Additional configuration sections (DNS, inbounds, outbounds)
- Phase 3: Configuration templates and presets
- Phase 4: Import/export existing configurations
- Phase 5: Configuration validation and testing tools
- Phase 6: Multi-user support and configuration versioning

## Success Criteria for Phase 1

- [x] Generator successfully extracts and converts sing-box route rule types
- [x] Web UI can display all available route rule types
- [x] Users can create, edit, and delete route rules through the UI
- [x] Generated configuration is valid and works with sing-box
- [x] Documentation is clear and comprehensive

**Status:** ✅ Phase 1 Complete - Production Ready

## Current Implementation Status

### Completed Features ✅

**Type Generation:**
- Automatic type extraction from sing-box source
- 19 rule-related types generated
- Support for complex generic types
- Generation metadata tracking

**Web Server:**
- Full HTMX-powered web interface
- Responsive design with custom CSS
- Server-side rendering
- Embedded templates and static files
- Graceful shutdown handling

**Rules Management:**
- Complete CRUD operations for 6 rule types
- Dynamic form generation with type reflection
- Drag-and-drop rule reordering
- JSON preview for each rule
- Smart field validation

**Service Management:**
- Start, stop, restart sing-box service
- Real-time status monitoring
- Service log viewer with configurable line counts
- Auto-reload after configuration changes

**Configuration Management:**
- Automatic backup before every save
- Manual backup creation with descriptions
- Backup metadata tracking
- Restore from any backup
- Export/import via backups
- File change detection with debouncing

**Deployment:**
- Single binary with embedded assets
- Debian package builds
- GitHub Actions CI/CD
- systemd service integration

### Next Development Phase

See [ROADMAP.md](../ROADMAP.md) for detailed future plans including:
- DNS configuration UI
- Inbounds/Outbounds management
- Rule templates and presets
- Configuration validation and testing
- Multi-configuration support
