# Sing-Box Web Config Manager - Roadmap

This document outlines the future development plans and potential improvements for the Sing-Box Web Config Manager.

## Current Status (v1.0)

The project has successfully completed its initial phase with a production-ready web interface for managing sing-box routing rules. All core features are functional and tested.

### Completed ✅

- ✅ Type generator from sing-box source (19 types)
- ✅ Full web server with HTMX
- ✅ Complete CRUD operations for 6 rule types
- ✅ Drag-and-drop rule reordering
- ✅ Service management (start, stop, restart)
- ✅ Real-time service monitoring and logs
- ✅ Automatic backup system with metadata
- ✅ Configuration export/import via backups
- ✅ File change detection and debouncing
- ✅ Embedded assets for single-binary deployment
- ✅ Debian package builds via GitHub Actions
- ✅ Comprehensive documentation

---

## Phase 2: Enhanced Configuration Management

### 2.1 DNS Configuration UI
**Priority:** High
**Effort:** Medium
**Timeline:** 2-3 weeks

- [ ] DNS servers management interface
- [ ] DNS rules CRUD operations (similar to routing rules)
- [ ] DNS strategy configuration
- [ ] Final DNS server selection
- [ ] Form builder for DNS-specific fields

**Benefits:**
- Complete DNS configuration without manual JSON editing
- Reuse existing form builder and CRUD patterns
- Consistent UI/UX with routing rules

### 2.2 Inbounds & Outbounds Management
**Priority:** High
**Effort:** High
**Timeline:** 3-4 weeks

- [ ] Inbound proxy configuration UI
- [ ] Outbound proxy configuration UI
- [ ] Support for all proxy types (SOCKS, HTTP, Shadowsocks, VMess, VLESS, Trojan, etc.)
- [ ] Type-specific form fields based on proxy protocol
- [ ] Connection testing for outbounds
- [ ] Import/export individual proxy configs

**Benefits:**
- Manage entire sing-box configuration through web UI
- Type-safe forms for each proxy protocol
- No need to manually edit complex proxy configurations

### 2.3 Log Configuration
**Priority:** Medium
**Effort:** Low
**Timeline:** 1 week

- [ ] Log level configuration (trace, debug, info, warn, error, fatal, panic)
- [ ] Log output destination (file path or stdout)
- [ ] Timestamp format options
- [ ] Disable color option
- [ ] Log rotation settings

**Benefits:**
- Complete configuration coverage
- Easy log management from UI

---

## Phase 3: Advanced Features

### 3.1 Rule Templates & Presets
**Priority:** High
**Effort:** Medium
**Timeline:** 2-3 weeks

- [ ] Common rule templates (block ads, bypass China, proxy international, etc.)
- [ ] User-defined custom templates
- [ ] Template marketplace/sharing
- [ ] One-click template application
- [ ] Template categories and tags

**Benefits:**
- Faster configuration setup
- Best practices sharing
- Reduced errors from common configurations

### 3.2 Search, Filter & Advanced Rule Management
**Priority:** Medium
**Effort:** Medium
**Timeline:** 2 weeks

- [ ] Search rules by content (domain, IP, outbound, etc.)
- [ ] Filter by rule type
- [ ] Filter by outbound target
- [ ] Bulk operations (delete, reorder, duplicate)
- [ ] Rule grouping and sections
- [ ] Copy/paste rules between configurations

**Benefits:**
- Manage large rule sets more efficiently
- Quick rule navigation and modification
- Better organization for complex configurations

### 3.3 Configuration Validation & Testing
**Priority:** High
**Effort:** High
**Timeline:** 3-4 weeks

- [ ] Real-time configuration validation
- [ ] Sing-box format check before saving
- [ ] Test configuration syntax (`sing-box check`)
- [ ] Domain/IP test tool (which rule matches?)
- [ ] Connection test for specific rules
- [ ] Validation warnings and suggestions

**Benefits:**
- Prevent invalid configurations
- Catch errors before deployment
- Debug routing issues

### 3.4 Configuration Comparison & Diff
**Priority:** Medium
**Effort:** Medium
**Timeline:** 2 weeks

- [ ] Visual diff between configurations
- [ ] Compare current config with backups
- [ ] Highlight changes in rules, DNS, proxies
- [ ] Side-by-side comparison view
- [ ] Merge capabilities

**Benefits:**
- Track what changed over time
- Understand impact of modifications
- Safer configuration updates

---

## Phase 4: User Experience Enhancements

### 4.1 Multi-Configuration Support
**Priority:** High
**Effort:** High
**Timeline:** 3 weeks

- [ ] Manage multiple sing-box configurations
- [ ] Switch between configurations
- [ ] Configuration profiles (home, work, travel, etc.)
- [ ] Quick-switch with service reload
- [ ] Per-profile backups

**Benefits:**
- Different configurations for different scenarios
- Easy A/B testing of configurations
- Profile-based setup management

### 4.2 Undo/Redo System
**Priority:** Medium
**Effort:** Medium
**Timeline:** 2 weeks

- [ ] Undo last change (rule add/edit/delete)
- [ ] Redo previously undone changes
- [ ] Change history with timestamps
- [ ] Rollback to any previous state
- [ ] Keyboard shortcuts (Ctrl+Z, Ctrl+Y)

**Benefits:**
- Safe experimentation
- Quick recovery from mistakes
- Better user confidence

### 4.3 Import/Export Enhancements
**Priority:** Medium
**Effort:** Low
**Timeline:** 1 week

- [ ] Import from URL (fetch remote config)
- [ ] Import sing-box config from clipboard
- [ ] Export individual sections (only rules, only DNS, etc.)
- [ ] Export to different formats (JSON, YAML)
- [ ] Bulk import rules from file

**Benefits:**
- Easier migration from existing configs
- Share configurations more easily
- Flexible export options

### 4.4 Dark Mode & Themes
**Priority:** Low
**Effort:** Medium
**Timeline:** 1-2 weeks

- [ ] Dark mode theme
- [ ] Light mode (current)
- [ ] System preference detection
- [ ] Theme persistence
- [ ] Custom theme colors

**Benefits:**
- Better user comfort
- Reduced eye strain
- Modern UI expectations

---

## Phase 5: Advanced Integration

### 5.1 GeoIP & GeoSite Database Integration
**Priority:** Medium
**Effort:** High
**Timeline:** 3-4 weeks

- [ ] Download and update GeoIP database
- [ ] Download and update GeoSite database
- [ ] Browse available GeoIP categories
- [ ] Browse available GeoSite categories
- [ ] Auto-complete for geoip/geosite fields
- [ ] Database update scheduler

**Benefits:**
- Better understanding of available geo categories
- Easier rule creation with auto-complete
- Keep databases up to date

### 5.2 Statistics & Monitoring
**Priority:** Medium
**Effort:** High
**Timeline:** 4-5 weeks

- [ ] Traffic statistics per rule
- [ ] Connection count monitoring
- [ ] Most-used rules analytics
- [ ] Outbound usage statistics
- [ ] Real-time connection viewer
- [ ] Historical data charts

**Benefits:**
- Understand traffic patterns
- Optimize rule order
- Identify unused rules

### 5.3 Rule Testing & Simulation
**Priority:** High
**Effort:** High
**Timeline:** 3 weeks

- [ ] Test which rule matches a domain/IP
- [ ] Simulate routing decisions
- [ ] Rule coverage analysis
- [ ] Conflict detection (overlapping rules)
- [ ] Performance impact estimates

**Benefits:**
- Debug routing issues quickly
- Optimize rule performance
- Prevent rule conflicts

---

## Phase 6: Advanced Administration

### 6.1 Multi-User Support
**Priority:** Low
**Effort:** Very High
**Timeline:** 5-6 weeks

- [ ] User authentication system
- [ ] Role-based access control (admin, editor, viewer)
- [ ] User activity logging
- [ ] Per-user preferences
- [ ] Audit trail for changes

**Benefits:**
- Multi-tenant environments
- Shared server management
- Accountability for changes

### 6.2 API & Automation
**Priority:** Medium
**Effort:** Medium
**Timeline:** 2-3 weeks

- [ ] RESTful API for all operations
- [ ] API authentication (token-based)
- [ ] API documentation (OpenAPI/Swagger)
- [ ] Webhook support for events
- [ ] CLI tool for API access

**Benefits:**
- Programmatic configuration management
- Integration with other tools
- Automation scripts

### 6.3 Backup Enhancements
**Priority:** Medium
**Effort:** Medium
**Timeline:** 2 weeks

- [ ] Scheduled automatic backups
- [ ] Backup retention policies (keep last N, or X days)
- [ ] Remote backup storage (S3, WebDAV, etc.)
- [ ] Backup compression
- [ ] Encrypted backups
- [ ] Backup size management

**Benefits:**
- Better disaster recovery
- Automated backup management
- Secure backup storage

---

## Phase 7: Performance & Scalability

### 7.1 Performance Optimizations
**Priority:** Medium
**Effort:** Medium
**Timeline:** 2-3 weeks

- [ ] Lazy loading for large rule sets
- [ ] Virtual scrolling for long lists
- [ ] Caching for frequently accessed data
- [ ] Optimistic UI updates
- [ ] Request debouncing and batching

**Benefits:**
- Handle large configurations smoothly
- Faster UI response times
- Better resource utilization

### 7.2 Testing & Quality Assurance
**Priority:** High
**Effort:** High
**Timeline:** 4-5 weeks

- [ ] Unit tests (target: 80%+ coverage)
- [ ] Integration tests for handlers
- [ ] E2E tests for critical workflows
- [ ] Performance benchmarks
- [ ] Automated testing in CI/CD
- [ ] Load testing

**Benefits:**
- Higher code quality
- Prevent regressions
- Confidence in releases

### 7.3 Documentation & Tutorials
**Priority:** High
**Effort:** Medium
**Timeline:** 2-3 weeks

- [ ] User guide with screenshots
- [ ] Video tutorials for common tasks
- [ ] API documentation
- [ ] Architecture documentation
- [ ] Troubleshooting guide
- [ ] FAQ section

**Benefits:**
- Lower barrier to entry
- Better user onboarding
- Reduced support burden

---

## Phase 8: Deployment & Distribution

### 8.1 Container Support
**Priority:** High
**Effort:** Medium
**Timeline:** 1-2 weeks

- [ ] Official Docker image
- [ ] Docker Compose example
- [ ] Container registry publishing
- [ ] Multi-arch builds (amd64, arm64)
- [ ] Docker deployment guide

**Benefits:**
- Easy deployment in containerized environments
- Cross-platform compatibility
- Standard deployment method

### 8.2 Additional Package Formats
**Priority:** Medium
**Effort:** Medium
**Timeline:** 2-3 weeks

- [ ] RPM packages (RedHat, CentOS, Fedora)
- [ ] Arch Linux AUR package
- [ ] Snap package
- [ ] Homebrew formula (macOS)
- [ ] Windows installer (MSI)

**Benefits:**
- Wider platform support
- Native package managers
- Easier installation

### 8.3 Release Automation
**Priority:** Medium
**Effort:** Low
**Timeline:** 1 week

- [ ] Automated version bumping
- [ ] Changelog generation from commits
- [ ] GitHub Releases automation
- [ ] Package upload automation
- [ ] Release notes generation

**Benefits:**
- Consistent release process
- Faster releases
- Better release documentation

---

## Long-Term Vision

### Potential Future Features

1. **Mobile App**: Native mobile app for iOS/Android for on-the-go management
2. **Configuration Marketplace**: Share and download community configurations
3. **AI-Powered Suggestions**: Smart rule recommendations based on usage patterns
4. **Cloud Sync**: Sync configurations across multiple devices
5. **Collaboration**: Real-time collaborative editing of configurations
6. **Plugin System**: Extensible architecture for custom features
7. **Multi-Language Support**: Internationalization (i18n) for global users
8. **Advanced Analytics**: ML-based traffic analysis and optimization
9. **Integration Hub**: Integrations with popular monitoring tools (Grafana, Prometheus)
10. **Configuration Generator**: Generate configs from high-level descriptions

---

## Timeline Summary

| Phase | Focus Area | Timeline | Priority |
|-------|-----------|----------|----------|
| Phase 2 | Enhanced Config Management | 6-8 weeks | High |
| Phase 3 | Advanced Features | 9-12 weeks | High |
| Phase 4 | User Experience | 7-10 weeks | Medium |
| Phase 5 | Advanced Integration | 10-13 weeks | Medium |
| Phase 6 | Advanced Administration | 9-11 weeks | Low |
| Phase 7 | Performance & Quality | 8-11 weeks | High |
| Phase 8 | Deployment & Distribution | 4-6 weeks | High |

**Total Estimated Timeline:** 9-12 months for complete feature set

---

## Contributing to the Roadmap

Have ideas for features not listed here? Open an issue or discussion on GitHub!

- **Feature Requests:** Use GitHub Issues with the `enhancement` label
- **Discussions:** Use GitHub Discussions for larger feature discussions
- **Pull Requests:** Contributions implementing roadmap features are always welcome

---

## Version Milestones

### v1.0 (Current) - Core Functionality ✅
- Basic rules management
- Service control
- Backup system

### v1.1 (Next) - Configuration Coverage
- DNS management
- Inbounds/Outbounds UI
- Log configuration

### v1.2 - Enhanced UX
- Rule templates
- Search & filtering
- Multi-configuration support

### v1.3 - Validation & Testing
- Configuration validation
- Rule testing tools
- Diff viewer

### v2.0 - Advanced Features
- GeoIP/GeoSite integration
- Statistics & monitoring
- API & automation

### v3.0 - Enterprise Features
- Multi-user support
- Advanced security
- High availability

---

**Last Updated:** 2025-01-21
**Status:** Active Development
