# Debian Packaging

This directory contains information about the Debian packaging process for the Sing-Box Web Config Manager.

## Package Structure

The `.deb` packages are built automatically by GitHub Actions and include:

### Files Installed

- `/usr/local/bin/singbox-web-config/singbox-web-config-server` - Main binary
- `/usr/local/bin/singbox-web-config/web/` - Web assets (HTML, CSS, JS)
- `/etc/systemd/system/singbox-web-config.service` - Systemd service file
- `/usr/share/doc/singbox-web-config/` - Documentation

### Directories Created

- `/etc/sing-box/backups/` - Configuration backups
- `/var/log/singbox-web-config/` - Log files

## Package Metadata

- **Package Name**: singbox-web-config
- **Section**: net
- **Priority**: optional
- **Dependencies**: systemd
- **Architectures**: amd64, arm64, armhf

## Installation Scripts

The package includes the following maintainer scripts:

### postinst (Post-Installation)

Runs after package installation:
- Creates backup directory at `/etc/sing-box/backups/`
- Creates log directory at `/var/log/singbox-web-config/`
- Reloads systemd daemon
- Displays installation instructions

### prerm (Pre-Removal)

Runs before package removal:
- Stops the singbox-web-config service if running
- Disables the service if enabled

### postrm (Post-Removal)

Runs after package removal:
- On purge: removes log directory
- Reloads systemd daemon
- Preserves configuration files in `/etc/sing-box/`

## Building Locally

To build DEB packages locally:

```bash
# Install required tools
sudo apt-get install dpkg-dev

# Build for current architecture
./scripts/build-deb.sh

# The package will be created in the current directory
ls -lh singbox-web-config_*.deb
```

## GitHub Actions Workflow

The automated build process:

1. **Triggers**:
   - Push to main or release branches
   - Git tags starting with 'v'
   - Pull requests to main
   - Manual workflow dispatch

2. **Build Matrix**:
   - amd64 (64-bit x86)
   - arm64 (64-bit ARM)
   - armhf (32-bit ARM)

3. **Artifacts**:
   - `.deb` package
   - `.sha256` checksum
   - `.md5` checksum

4. **Release Creation**:
   - Automatically created for version tags
   - Includes all architecture packages
   - Contains checksums for verification
   - Provides installation instructions

5. **Testing**:
   - Tests installation on Ubuntu 22.04, 24.04
   - Tests installation on Debian 11, 12
   - Verifies file presence and systemd service

## Installation

### From DEB Package

```bash
# Download the package for your architecture
wget https://github.com/matinhimself/singbox-web-config/releases/download/v1.0.0/singbox-web-config_1.0.0_amd64.deb

# Verify checksum (optional but recommended)
wget https://github.com/matinhimself/singbox-web-config/releases/download/v1.0.0/singbox-web-config_1.0.0_amd64.deb.sha256
sha256sum -c singbox-web-config_1.0.0_amd64.deb.sha256

# Install
sudo dpkg -i singbox-web-config_1.0.0_amd64.deb

# Install dependencies if needed
sudo apt-get install -f

# Enable and start service
sudo systemctl enable singbox-web-config
sudo systemctl start singbox-web-config

# Check status
sudo systemctl status singbox-web-config
```

### Configuration

After installation, configure the service:

```bash
# Edit service file to customize settings
sudo systemctl edit singbox-web-config --full

# Common customizations:
# - Change listen address: -addr 127.0.0.1:8080
# - Change config path: -config /path/to/config.json
# - Change service name: -service your-singbox-service

# Reload and restart after changes
sudo systemctl daemon-reload
sudo systemctl restart singbox-web-config
```

## Upgrading

```bash
# Download new version
wget https://github.com/matinhimself/singbox-web-config/releases/download/v2.0.0/singbox-web-config_2.0.0_amd64.deb

# Install (will upgrade existing installation)
sudo dpkg -i singbox-web-config_2.0.0_amd64.deb

# Restart service
sudo systemctl restart singbox-web-config
```

## Removal

```bash
# Remove package but keep configuration
sudo dpkg -r singbox-web-config

# Remove package and configuration (purge)
sudo dpkg --purge singbox-web-config
```

## Supported Distributions

Tested and supported on:
- Ubuntu 22.04 LTS (Jammy)
- Ubuntu 24.04 LTS (Noble)
- Debian 11 (Bullseye)
- Debian 12 (Bookworm)

Should work on any Debian-based distribution with systemd support.

## Architecture Support

- **amd64**: Standard 64-bit x86 processors (Intel/AMD)
- **arm64**: 64-bit ARM processors (Raspberry Pi 4, AWS Graviton, etc.)
- **armhf**: 32-bit ARM processors (Raspberry Pi 3 and older)

## Troubleshooting

### Package Installation Fails

```bash
# Check for dependency issues
sudo apt-get install -f

# Force installation (not recommended)
sudo dpkg -i --force-all singbox-web-config_*.deb
```

### Service Won't Start

```bash
# Check service status
sudo systemctl status singbox-web-config

# View logs
sudo journalctl -u singbox-web-config -n 50

# Common issues:
# - Config file not found: Update -config path in service file
# - Port in use: Change -addr in service file
# - Permission denied: Check file permissions
```

### Verify Package Contents

```bash
# List package contents before installation
dpkg -c singbox-web-config_*.deb

# List installed files after installation
dpkg -L singbox-web-config

# Check package status
dpkg -s singbox-web-config
```

## Contributing

To improve the packaging:

1. Test on different distributions
2. Report issues with installation
3. Suggest improvements to maintainer scripts
4. Add support for more architectures

## References

- [Debian Policy Manual](https://www.debian.org/doc/debian-policy/)
- [Debian Maintainer Scripts](https://www.debian.org/doc/debian-policy/ch-maintainerscripts.html)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
