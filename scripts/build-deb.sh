#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Sing-Box Web Config Manager - DEB Package Builder${NC}"
echo "=================================================="
echo ""

# Check if running on Linux
if [[ "$OSTYPE" != "linux-gnu"* ]]; then
    echo -e "${RED}Error: This script must be run on Linux${NC}"
    exit 1
fi

# Check for required tools
if ! command -v dpkg-deb &> /dev/null; then
    echo -e "${RED}Error: dpkg-deb is not installed${NC}"
    echo "Install it with: sudo apt-get install dpkg-dev"
    exit 1
fi

# Get version from git tag or use dev version
if git describe --tags --exact-match 2>/dev/null; then
    VERSION=$(git describe --tags --exact-match | sed 's/^v//')
else
    VERSION="0.0.0-dev.$(git rev-parse --short HEAD)"
fi

# Architecture mapping
ARCH=$(dpkg --print-architecture)
GOARCH=""
GOARM=""

case $ARCH in
    amd64)
        GOARCH="amd64"
        ;;
    arm64)
        GOARCH="arm64"
        ;;
    armhf)
        GOARCH="arm"
        GOARM="7"
        ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

echo -e "Version: ${YELLOW}$VERSION${NC}"
echo -e "Architecture: ${YELLOW}$ARCH${NC}"
echo ""

# Build Go binary
echo -e "${GREEN}Building Go binary...${NC}"
CGO_ENABLED=0 GOOS=linux GOARCH=$GOARCH GOARM=$GOARM go build -v -trimpath \
    -ldflags="-s -w -X main.Version=$VERSION" \
    -o singbox-web-config-server \
    ./cmd/server/main.go

if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Failed to build binary${NC}"
    exit 1
fi

echo -e "${GREEN}Binary built successfully${NC}"
echo ""

# Create package directory structure
PKG_DIR="singbox-web-config_${VERSION}_${ARCH}"
echo -e "${GREEN}Creating package structure: $PKG_DIR${NC}"

rm -rf "$PKG_DIR"
mkdir -p "${PKG_DIR}/DEBIAN"
mkdir -p "${PKG_DIR}/usr/local/bin/singbox-web-config"
mkdir -p "${PKG_DIR}/usr/local/bin/singbox-web-config/web"
mkdir -p "${PKG_DIR}/etc/systemd/system"
mkdir -p "${PKG_DIR}/usr/share/doc/singbox-web-config"

# Copy binary
echo "Copying binary..."
cp singbox-web-config-server "${PKG_DIR}/usr/local/bin/singbox-web-config/"
chmod +x "${PKG_DIR}/usr/local/bin/singbox-web-config/singbox-web-config-server"

# Copy web assets
echo "Copying web assets..."
cp -r web/* "${PKG_DIR}/usr/local/bin/singbox-web-config/web/"

# Copy systemd service
echo "Copying systemd service..."
cp deploy/singbox-web-config.service "${PKG_DIR}/etc/systemd/system/"

# Copy documentation
echo "Copying documentation..."
cp README.md "${PKG_DIR}/usr/share/doc/singbox-web-config/"
cp INSTALL.md "${PKG_DIR}/usr/share/doc/singbox-web-config/"

# Create control file
echo "Creating control file..."
cat > "${PKG_DIR}/DEBIAN/control" <<EOF
Package: singbox-web-config
Version: $VERSION
Section: net
Priority: optional
Architecture: $ARCH
Maintainer: Matin <matin@example.com>
Homepage: https://github.com/matinhimself/singbox-web-config
Description: Web-based configuration manager for sing-box
 A simple web UI configuration manager for sing-box, starting with
 route rules management. Features automatic type generation, clean
 responsive interface using HTMX, and type-safe configurations.
Depends: systemd
EOF

# Create postinst script
echo "Creating postinst script..."
cat > "${PKG_DIR}/DEBIAN/postinst" <<'EOF'
#!/bin/bash
set -e

# Create backup directory if it doesn't exist
mkdir -p /etc/sing-box/backups
chmod 755 /etc/sing-box/backups

# Create log directory
mkdir -p /var/log/singbox-web-config
chmod 755 /var/log/singbox-web-config

# Reload systemd daemon
systemctl daemon-reload

echo "Sing-Box Web Config Manager installed successfully!"
echo ""
echo "To enable and start the service:"
echo "  sudo systemctl enable singbox-web-config"
echo "  sudo systemctl start singbox-web-config"
echo ""
echo "The web interface will be available at http://0.0.0.0:8080"
echo ""
echo "Documentation: /usr/share/doc/singbox-web-config/"
echo ""

exit 0
EOF
chmod +x "${PKG_DIR}/DEBIAN/postinst"

# Create prerm script
echo "Creating prerm script..."
cat > "${PKG_DIR}/DEBIAN/prerm" <<'EOF'
#!/bin/bash
set -e

# Stop service if running
if systemctl is-active --quiet singbox-web-config; then
    systemctl stop singbox-web-config
fi

# Disable service if enabled
if systemctl is-enabled --quiet singbox-web-config; then
    systemctl disable singbox-web-config
fi

exit 0
EOF
chmod +x "${PKG_DIR}/DEBIAN/prerm"

# Create postrm script
echo "Creating postrm script..."
cat > "${PKG_DIR}/DEBIAN/postrm" <<'EOF'
#!/bin/bash
set -e

if [ "$1" = "purge" ]; then
    # Remove log directory on purge
    rm -rf /var/log/singbox-web-config

    echo "Sing-Box Web Config Manager removed."
    echo "Note: Configuration files in /etc/sing-box/ were preserved."
fi

# Reload systemd daemon
systemctl daemon-reload

exit 0
EOF
chmod +x "${PKG_DIR}/DEBIAN/postrm"

echo ""
echo -e "${GREEN}Building DEB package...${NC}"
dpkg-deb --build --root-owner-group "$PKG_DIR"

if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Failed to build DEB package${NC}"
    exit 1
fi

# Calculate checksums
echo ""
echo -e "${GREEN}Calculating checksums...${NC}"
sha256sum "${PKG_DIR}.deb" > "${PKG_DIR}.deb.sha256"
md5sum "${PKG_DIR}.deb" > "${PKG_DIR}.deb.md5"

# Display results
echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""
echo "Package: ${PKG_DIR}.deb"
echo "Size: $(du -h "${PKG_DIR}.deb" | cut -f1)"
echo ""
echo "Checksums:"
cat "${PKG_DIR}.deb.sha256"
cat "${PKG_DIR}.deb.md5"
echo ""
echo -e "${YELLOW}To install:${NC}"
echo "  sudo dpkg -i ${PKG_DIR}.deb"
echo "  sudo apt-get install -f  # If dependencies are missing"
echo ""
echo -e "${YELLOW}To test package contents:${NC}"
echo "  dpkg -c ${PKG_DIR}.deb"
echo ""
echo -e "${YELLOW}To verify checksums:${NC}"
echo "  sha256sum -c ${PKG_DIR}.deb.sha256"
echo ""

# Clean up
rm -f singbox-web-config-server

echo -e "${GREEN}Done!${NC}"
