#!/bin/bash

set -e

echo "Installing Sing-Box Web Config Manager..."

# Build the application
echo "Building..."
go build -o singbox-web-config-server cmd/server/main.go

# Create installation directory
INSTALL_DIR="/usr/local/bin/singbox-web-config"
echo "Creating installation directory: $INSTALL_DIR"
sudo mkdir -p "$INSTALL_DIR"
sudo mkdir -p "$INSTALL_DIR/web"

# Copy files
echo "Copying files..."
sudo cp singbox-web-config-server "$INSTALL_DIR/"
sudo cp -r web/* "$INSTALL_DIR/web/"

# Set permissions
sudo chmod +x "$INSTALL_DIR/singbox-web-config-server"

# Install systemd service
echo "Installing systemd service..."
sudo cp deploy/singbox-web-config.service /etc/systemd/system/
sudo systemctl daemon-reload

# Enable and start service
echo "Enabling service..."
sudo systemctl enable singbox-web-config

echo ""
echo "Installation complete!"
echo ""
echo "To start the service:"
echo "  sudo systemctl start singbox-web-config"
echo ""
echo "To view logs:"
echo "  sudo journalctl -u singbox-web-config -f"
echo ""
echo "The web interface will be available at http://localhost:8080"
echo ""
