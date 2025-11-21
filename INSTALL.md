# Installation and Deployment Guide

## Sing-Box Web Config Manager

This guide will help you install and configure the Sing-Box Web Config Manager as a systemd service that manages your sing-box routing configuration.

## Prerequisites

- Go 1.21 or higher
- Sing-box installed and configured
- Linux system with systemd
- Root or sudo access

## Quick Installation

### 1. Build the Application

```bash
# Clone the repository
cd /path/to/singbox-web-config

# Build the application
go build -o singbox-web-config-server cmd/server/main.go
```

### 2. Install as Systemd Service

Use the provided installation script:

```bash
chmod +x deploy/install.sh
sudo ./deploy/install.sh
```

Or manually:

```bash
# Create installation directory
sudo mkdir -p /usr/local/bin/singbox-web-config
sudo mkdir -p /usr/local/bin/singbox-web-config/web

# Copy files
sudo cp singbox-web-config-server /usr/local/bin/singbox-web-config/
sudo cp -r web/* /usr/local/bin/singbox-web-config/web/

# Set permissions
sudo chmod +x /usr/local/bin/singbox-web-config/singbox-web-config-server

# Install systemd service
sudo cp deploy/singbox-web-config.service /etc/systemd/system/
sudo systemctl daemon-reload

# Enable and start service
sudo systemctl enable singbox-web-config
sudo systemctl start singbox-web-config
```

### 3. Verify Installation

```bash
# Check service status
sudo systemctl status singbox-web-config

# View logs
sudo journalctl -u singbox-web-config -f
```

## Configuration

### Systemd Service Configuration

Edit `/etc/systemd/system/singbox-web-config.service` to customize:

- **Listen Address**: Change `-addr 0.0.0.0:8080` to your preferred address
- **Config Path**: Change `-config /etc/sing-box/config.json` to your config location
- **Service Name**: Change `-service sing-box` if your sing-box service has a different name

Example:

```ini
[Service]
ExecStart=/usr/local/bin/singbox-web-config/singbox-web-config-server \
    -addr 127.0.0.1:8080 \
    -config /etc/sing-box/config.json \
    -service sing-box
```

After making changes:

```bash
sudo systemctl daemon-reload
sudo systemctl restart singbox-web-config
```

### Command-Line Options

When running manually:

```bash
singbox-web-config-server \
    -addr <address:port>      # HTTP server address (default: localhost:8080)
    -config <path>            # Path to sing-box config file (default: /etc/sing-box/config.json)
    -service <name>           # Sing-box systemd service name (default: sing-box)
```

## Features

### 1. Config File Watcher

The application automatically watches for changes to the sing-box config file and can notify you when external modifications occur.

### 2. Service Management

Control the sing-box service directly from the web interface:
- Start/Stop/Restart the service
- View service status
- Check service logs

### 3. Rule Management

- Add, edit, and delete routing rules through a dynamic web interface
- Forms are automatically generated from sing-box types
- Support for all rule types (Domain, GeoIP, CIDR, etc.)

### 4. Config Backup

- Automatic backups before any config changes
- Restore from previous backups
- Export current configuration

## Usage

### Access the Web Interface

Open your browser and navigate to:
- If using default settings: `http://localhost:8080`
- If using custom address: `http://your-address:your-port`

### Managing Rules

1. Click "Route Rules" in the navigation
2. Click "+ Add Rule" to create a new rule
3. Select the rule type from the dropdown
4. Fill in the form fields (arrays accept comma-separated values)
5. Click "Create Rule"

### Service Management

1. Click "Service" in the navigation
2. View current service status
3. Use buttons to Start/Stop/Restart the service
4. View recent logs

### Backup and Restore

1. Go to the Service page
2. Scroll to "Configuration Backups"
3. View available backups
4. Click "Restore" to restore a backup
5. Use "Export Current Config" to download the config

## Security Considerations

### Network Access

By default, the service listens on `0.0.0.0:8080`, making it accessible from any network interface. For better security:

1. **Listen only on localhost**:
   ```bash
   -addr 127.0.0.1:8080
   ```

2. **Use a reverse proxy** (nginx, Apache) with HTTPS and authentication:
   ```nginx
   server {
       listen 443 ssl;
       server_name config.example.com;

       ssl_certificate /path/to/cert.pem;
       ssl_certificate_key /path/to/key.pem;

       location / {
           auth_basic "Sing-Box Config";
           auth_basic_user_file /etc/nginx/.htpasswd;
           proxy_pass http://127.0.0.1:8080;
       }
   }
   ```

3. **Firewall rules**: Restrict access using iptables or firewalld:
   ```bash
   sudo firewall-cmd --add-rich-rule='rule family="ipv4" source address="192.168.1.0/24" port protocol="tcp" port="8080" accept'
   ```

### File Permissions

Ensure proper permissions:

```bash
# Config file should be readable by the service
sudo chmod 644 /etc/sing-box/config.json

# Backup directory needs write access
sudo chown -R root:root /etc/sing-box/backups
sudo chmod 755 /etc/sing-box/backups
```

## Troubleshooting

### Service Won't Start

```bash
# Check service status
sudo systemctl status singbox-web-config

# View detailed logs
sudo journalctl -u singbox-web-config -n 50

# Common issues:
# - Config file not found: Check path in service file
# - Permission denied: Ensure service user has read access to config
# - Port already in use: Change the listen address
```

### Config Changes Not Applied

```bash
# Manually reload sing-box
sudo systemctl restart sing-box

# Check sing-box status
sudo systemctl status sing-box

# View sing-box logs for errors
sudo journalctl -u sing-box -n 50
```

### File Watcher Not Working

```bash
# Check if fsnotify is working
# Verify the config file path is correct
# Ensure the parent directory exists
```

## Uninstallation

```bash
# Stop and disable service
sudo systemctl stop singbox-web-config
sudo systemctl disable singbox-web-config

# Remove service file
sudo rm /etc/systemd/system/singbox-web-config.service
sudo systemctl daemon-reload

# Remove application files
sudo rm -rf /usr/local/bin/singbox-web-config
```

## Updating

```bash
# Pull latest changes
git pull

# Rebuild
go build -o singbox-web-config-server cmd/server/main.go

# Reinstall
sudo cp singbox-web-config-server /usr/local/bin/singbox-web-config/
sudo systemctl restart singbox-web-config
```

## Support

For issues, questions, or contributions:
- GitHub Issues: https://github.com/matinhimself/singbox-web-config/issues
- Documentation: See specs/ directory for detailed architecture docs
