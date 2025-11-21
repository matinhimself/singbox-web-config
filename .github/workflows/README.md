# GitHub Actions Workflows

This directory contains automated workflows for the Sing-Box Web Config Manager project.

## Available Workflows

### build-deb.yml - DEB Package Builder

Automatically builds Debian packages for multiple architectures.

#### Triggers

- **Push Events**:
  - Pushes to `main` branch
  - Pushes to `release/**` branches
  - Tags starting with `v*` (e.g., v1.0.0)
- **Pull Requests**:
  - PRs targeting `main` branch
- **Manual**:
  - Workflow can be triggered manually via GitHub UI

#### Build Matrix

The workflow builds packages for three architectures:

| Architecture | GOARCH | Description | Use Case |
|--------------|--------|-------------|----------|
| amd64 | amd64 | 64-bit x86 | Standard desktop/server systems |
| arm64 | arm64 | 64-bit ARM | Raspberry Pi 4, AWS Graviton, Apple Silicon |
| armhf | arm (v7) | 32-bit ARM | Raspberry Pi 3 and older |

#### Jobs

##### 1. build-deb

Builds the DEB packages for all architectures.

**Steps:**
1. Checkout code
2. Set up Go 1.25
3. Determine version from git tag or commit
4. Download and verify Go dependencies
5. Build binary with optimizations (`-trimpath`, `-ldflags`)
6. Create DEB package structure
7. Generate control files and maintainer scripts
8. Build DEB package using `dpkg-deb`
9. Calculate SHA256 and MD5 checksums
10. Upload artifacts (retained for 30 days)

**Build Flags:**
- `CGO_ENABLED=0` - Static binary without CGO
- `-trimpath` - Remove file system paths from binary
- `-ldflags="-s -w"` - Strip debug info and symbol table
- `-X main.Version=...` - Embed version string

##### 2. create-release

Creates a GitHub release when a version tag is pushed.

**Steps:**
1. Download all build artifacts
2. Extract version from tag
3. Create GitHub release with:
   - All DEB packages (amd64, arm64, armhf)
   - Checksums for verification
   - Installation instructions
   - Changelog link

**Trigger:** Only runs for tags matching `refs/tags/v*`

**Permissions:** Requires `contents: write` to create releases

##### 3. test-install

Tests package installation on multiple distributions.

**Test Matrix:**

| Distribution | Version | Test Purpose |
|--------------|---------|--------------|
| Ubuntu | 22.04 LTS | Long-term support release |
| Ubuntu | 24.04 LTS | Latest LTS |
| Debian | 11 (Bullseye) | Stable release |
| Debian | 12 (Bookworm) | Current stable |

**Validation:**
- Package installs without errors
- Dependencies resolve correctly
- Files are placed in correct locations
- Systemd service file exists

## Usage

### Automatic Builds

Builds are triggered automatically on:

```bash
# Push to main
git push origin main

# Create and push a tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Push to release branch
git checkout -b release/v1.0
git push origin release/v1.0
```

### Manual Workflow Dispatch

1. Go to the Actions tab in GitHub
2. Select "Build DEB Packages" workflow
3. Click "Run workflow"
4. Select the branch
5. Click "Run workflow" button

### Downloading Artifacts

#### From Workflow Run

1. Go to Actions tab
2. Click on a workflow run
3. Scroll to Artifacts section
4. Download the desired architecture

#### From Release

1. Go to Releases page
2. Find the desired version
3. Download from Assets section

## Version Management

### Version Determination

The workflow determines version in two ways:

1. **Tagged Release**: Uses the git tag (e.g., `v1.0.0` → `1.0.0`)
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```

2. **Development Build**: Uses `0.0.0-dev.<commit-hash>`
   ```bash
   # Automatic for non-tagged commits
   # Example: 0.0.0-dev.a1b2c3d
   ```

### Version in Binary

The version is embedded in the binary during build:
```go
var Version string // Set via -ldflags at build time
```

## Package Contents

Each DEB package includes:

```
/usr/local/bin/singbox-web-config/
├── singbox-web-config-server          # Main binary
└── web/                                # Web assets
    ├── templates/
    └── static/

/etc/systemd/system/
└── singbox-web-config.service         # Systemd service

/usr/share/doc/singbox-web-config/
├── README.md                          # Project documentation
└── INSTALL.md                         # Installation guide
```

## Maintainer Scripts

### postinst (Post-Installation)

Executed after package installation:
- Creates `/etc/sing-box/backups/` directory
- Creates `/var/log/singbox-web-config/` directory
- Reloads systemd daemon
- Displays installation instructions

### prerm (Pre-Removal)

Executed before package removal:
- Stops the service if running
- Disables the service if enabled

### postrm (Post-Removal)

Executed after package removal:
- On purge: removes log directory
- Reloads systemd daemon
- Preserves config files in `/etc/sing-box/`

## Artifact Details

Each build produces:

```
singbox-web-config_<version>_<arch>.deb        # DEB package
singbox-web-config_<version>_<arch>.deb.sha256 # SHA256 checksum
singbox-web-config_<version>_<arch>.deb.md5    # MD5 checksum
```

**Retention**: Artifacts are kept for 30 days

## Security Considerations

### Build Security

- Builds run in GitHub-hosted runners (Ubuntu latest)
- Dependencies are verified with `go mod verify`
- Static binaries (CGO disabled) for portability
- No network access during build (vendored dependencies recommended)

### Package Security

- Packages are built with `--root-owner-group` flag
- Service runs as root (required for systemd management)
- File permissions are explicitly set
- Systemd security features enabled:
  - `PrivateTmp=true`
  - `ProtectSystem=strict`
  - `ProtectHome=true`

### Verification

Users can verify packages:
```bash
# Verify SHA256 checksum
sha256sum -c singbox-web-config_1.0.0_amd64.deb.sha256

# Verify MD5 checksum
md5sum -c singbox-web-config_1.0.0_amd64.deb.md5

# Inspect package contents
dpkg -c singbox-web-config_1.0.0_amd64.deb

# View package info
dpkg -I singbox-web-config_1.0.0_amd64.deb
```

## Troubleshooting

### Workflow Fails on Build

Check the build logs for:
- Go compilation errors
- Missing dependencies
- Network issues during `go mod download`

### Package Installation Fails in Tests

Common issues:
- Missing systemd in test container
- Architecture mismatch
- Dependency resolution problems

### Release Creation Fails

Requires:
- Valid tag format (v*)
- `contents: write` permission
- All build jobs to succeed

## Local Development

To test package building locally:

```bash
# Build package for current architecture
./scripts/build-deb.sh

# Test package installation
sudo dpkg -i singbox-web-config_*.deb

# Remove test installation
sudo dpkg -r singbox-web-config
```

## Future Improvements

Potential enhancements:

- [ ] Add RPM package support
- [ ] Support more architectures (mips, riscv64)
- [ ] Add package signing with GPG
- [ ] Create APT repository
- [ ] Add Docker image builds
- [ ] Implement package version checks
- [ ] Add performance benchmarks
- [ ] Create snap/flatpak packages

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Debian Package Management](https://www.debian.org/doc/manuals/debian-faq/pkg-basics)
- [Go Build Documentation](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies)
- [systemd Service Files](https://www.freedesktop.org/software/systemd/man/systemd.service.html)
