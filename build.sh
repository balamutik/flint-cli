#!/bin/bash

# Flint Vault Cross-Platform Build Script
# Builds releases for multiple platforms and architectures

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="flint-vault"
BUILD_DIR="build"
DIST_DIR="dist"
VERSION=${1:-"dev"}
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build targets (platform:goos:goarch format)
TARGETS=(
    "linux-amd64:linux:amd64"
    "linux-arm64:linux:arm64"
    "darwin-amd64:darwin:amd64"
    "darwin-arm64:darwin:arm64"
    "windows-amd64:windows:amd64"
)

echo -e "${BLUE}🚀 Flint Vault Cross-Platform Build${NC}"
echo -e "${BLUE}====================================${NC}"
echo "Version: $VERSION"
echo "Commit: $COMMIT_HASH"
echo "Build Time: $BUILD_TIME"
echo ""

# Clean previous builds
echo -e "${YELLOW}🧹 Cleaning previous builds...${NC}"
rm -rf "$BUILD_DIR" "$DIST_DIR"
mkdir -p "$BUILD_DIR" "$DIST_DIR"

# Go module cleanup and verification
echo -e "${YELLOW}🔍 Verifying Go modules...${NC}"
go mod tidy
go mod verify

# Run tests before building
echo -e "${YELLOW}🧪 Running tests...${NC}"
if ! go test ./... -short; then
    echo -e "${RED}❌ Tests failed! Aborting build.${NC}"
    exit 1
fi
echo -e "${GREEN}✅ All tests passed!${NC}"

# Build function
build_target() {
    local target_spec=$1
    
    # Parse target specification
    IFS=':' read -r target goos goarch <<< "$target_spec"
    
    echo -e "${BLUE}🔨 Building $target...${NC}"
    
    # Set environment variables
    export GOOS=$goos
    export GOARCH=$goarch
    export CGO_ENABLED=0
    
    # Determine binary name (add .exe for Windows)
    local binary_name="$APP_NAME"
    if [ "$goos" = "windows" ]; then
        binary_name="${APP_NAME}.exe"
    fi
    
    # Build with ldflags for version info
    local ldflags="-s -w -X main.Version=$VERSION -X main.GitCommit=$COMMIT_HASH -X main.BuildTime=$BUILD_TIME"
    local output_path="$BUILD_DIR/$target/$binary_name"
    
    # Create target directory
    mkdir -p "$BUILD_DIR/$target"
    
    # Build binary
    if go build -ldflags "$ldflags" -o "$output_path" ./cmd; then
        echo -e "${GREEN}  ✅ Built: $output_path${NC}"
        
        # Get binary size
        local size=$(du -h "$output_path" | cut -f1)
        echo -e "     📦 Size: $size"
        
        # Create archive
        create_archive "$target" "$binary_name" "$goos"
        
    else
        echo -e "${RED}  ❌ Failed to build $target${NC}"
        return 1
    fi
}

# Archive creation function
create_archive() {
    local target=$1
    local binary_name=$2
    local goos=$3
    
    local archive_dir="$BUILD_DIR/$target"
    local archive_name="$APP_NAME-$VERSION-$target"
    
    # Copy additional files
    cp README.md "$archive_dir/"
    cp LICENSE "$archive_dir/" 2>/dev/null || echo "LICENSE file not found, skipping..."
    
    # Create installation script for Unix systems
    if [ "$goos" != "windows" ]; then
        cat > "$archive_dir/install.sh" << 'EOF'
#!/bin/bash
# Flint Vault Installation Script

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="flint-vault"

echo "Installing Flint Vault to $INSTALL_DIR..."

# Check if directory exists and is writable
if [ ! -d "$INSTALL_DIR" ]; then
    echo "Creating directory: $INSTALL_DIR"
    sudo mkdir -p "$INSTALL_DIR"
fi

# Copy binary
if sudo cp "$BINARY_NAME" "$INSTALL_DIR/"; then
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    echo "✅ Flint Vault installed successfully!"
    echo "Run 'flint-vault --help' to get started."
else
    echo "❌ Installation failed. Check permissions."
    exit 1
fi
EOF
        chmod +x "$archive_dir/install.sh"
    fi
    
    # Create archive based on platform
    cd "$BUILD_DIR"
    
    if [ "$goos" = "windows" ]; then
        # ZIP for Windows
        zip -r "../$DIST_DIR/${archive_name}.zip" "$target/" > /dev/null
        echo -e "  📦 Created: ${archive_name}.zip"
    else
        # TAR.GZ for Unix systems
        tar -czf "../$DIST_DIR/${archive_name}.tar.gz" "$target/"
        echo -e "  📦 Created: ${archive_name}.tar.gz"
    fi
    
    cd ..
}

# Build all targets
echo -e "${YELLOW}🔨 Building for all platforms...${NC}"
echo ""

failed_builds=()

for target_spec in "${TARGETS[@]}"; do
    if ! build_target "$target_spec"; then
        # Extract target name for failed builds tracking
        target=$(echo "$target_spec" | cut -d':' -f1)
        failed_builds+=("$target")
    fi
    echo ""
done

# Generate checksums
echo -e "${YELLOW}🔐 Generating checksums...${NC}"
cd "$DIST_DIR"

# Check if we have any files to checksum
if ls ./* 1> /dev/null 2>&1; then
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum * > checksums.txt
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 * > checksums.txt
    else
        echo "Warning: No checksum utility found"
    fi
    echo -e "${GREEN}✅ Checksums generated: checksums.txt${NC}"
else
    echo -e "${YELLOW}⚠️  No files to generate checksums for${NC}"
fi

cd ..

# Build summary
echo ""
echo -e "${BLUE}📋 Build Summary${NC}"
echo -e "${BLUE}================${NC}"

if [ ${#failed_builds[@]} -eq 0 ]; then
    echo -e "${GREEN}✅ All builds completed successfully!${NC}"
    echo ""
    echo "📦 Release artifacts:"
    ls -la "$DIST_DIR"
    
    echo ""
    echo -e "${BLUE}🚀 Ready for GitHub Release!${NC}"
    echo "Upload the files in '$DIST_DIR' to your GitHub release."
    
else
    echo -e "${RED}❌ Some builds failed:${NC}"
    for failed in "${failed_builds[@]}"; do
        echo -e "${RED}  - $failed${NC}"
    done
    exit 1
fi

echo ""
echo -e "${GREEN}🎉 Build completed successfully!${NC}"
echo -e "${YELLOW}💡 Next steps:${NC}"
echo "1. Create a new release on GitHub"
echo "2. Upload files from '$DIST_DIR' directory"
echo "3. Use checksums.txt for verification" 