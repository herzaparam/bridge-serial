#!/usr/bin/env bash

set -euo pipefail

# Configuration
APP_NAME="rapier-bridge"
VERSION="1.0.0"
BUILD_DIR="bin"
ENTRY_POINT="cmd/app/main.go"

# Architecture configurations
ARCHITECTURES=(
    "amd64"
    "386"
    "arm"
    "arm64"
)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}')
    log_info "Found Go version: $GO_VERSION"
}

# Clean previous builds
clean_build() {
    log_info "Cleaning previous Windows builds..."
    
    # Create main build directory if it doesn't exist
    mkdir -p "$BUILD_DIR"
    
    # Remove only Windows-specific directories
    for arch in "${ARCHITECTURES[@]}"; do
        local arch_dir="${BUILD_DIR}/windows-${arch}"
        if [ -d "$arch_dir" ]; then
            log_info "Removing $arch_dir"
            rm -rf "$arch_dir"
        fi
    done
}

# Download dependencies
download_deps() {
    log_info "Downloading dependencies..."
    go mod download
    go mod tidy
}

# Build for specific Windows architecture
build_windows_arch() {
    local arch=$1
    local output_name="${APP_NAME}-${arch}.exe"
    local arch_dir="${BUILD_DIR}/windows-${arch}"
    
    log_info "Building for Windows ${arch}..."
    
    # Create architecture-specific directory
    mkdir -p "$arch_dir"
    
    # Set environment variables for cross-compilation
    export GOOS=windows
    export GOARCH="$arch"
    export CGO_ENABLED=1  # Enable CGO for Fyne/OpenGL
    export CC=x86_64-w64-mingw32-gcc  # Cross-compiler for Windows
    export FYNE_SCALE=1.0
    
    # Build the application
    go build \
        -ldflags="-s -w -X main.version=$VERSION -H windowsgui" \
        -o "$arch_dir/$output_name" \
        "$ENTRY_POINT"
    
    if [ $? -eq 0 ]; then
        log_success "Windows ${arch} build completed successfully!"
        log_info "Output: $arch_dir/$output_name"
        
        # Create build info for this architecture
        create_build_info_arch "$arch" "$arch_dir"
        
        # Show file information
        show_file_info_arch "$arch" "$arch_dir" "$output_name"
    else
        log_error "Windows ${arch} build failed!"
        return 1
    fi
}

# Create build info for specific architecture
create_build_info_arch() {
    local arch=$1
    local arch_dir=$2
    
    log_info "Creating build information for ${arch}..."
    
    cat > "$arch_dir/build-info.txt" << EOF
Build Information
=================
Application: $APP_NAME
Version: $VERSION
Architecture: Windows ${arch}
Build Date: $(date -u '+%Y-%m-%d %H:%M:%S UTC')
Build Platform: $(uname -s) $(uname -m)
Go Version: $(go version)
Target Platform: Windows ${arch}
Entry Point: $ENTRY_POINT
EOF

    log_success "Build information saved to $arch_dir/build-info.txt"
}

# Show file information for specific architecture
show_file_info_arch() {
    local arch=$1
    local arch_dir=$2
    local output_name=$3
    
    if [ -f "$arch_dir/$output_name" ]; then
        log_info "Build file information for ${arch}:"
        ls -lh "$arch_dir/$output_name"
    fi
}

# Build for all Windows architectures
build_all_windows() {
    log_info "Building for all Windows architectures..."
    
    local success_count=0
    local total_count=${#ARCHITECTURES[@]}
    
    for arch in "${ARCHITECTURES[@]}"; do
        echo ""
        if build_windows_arch "$arch"; then
            ((success_count++))
        else
            log_warning "Failed to build for ${arch}"
        fi
    done
    
    echo ""
    log_info "Build Summary: $success_count/$total_count architectures built successfully"
    
    if [ $success_count -eq $total_count ]; then
        log_success "All Windows architectures built successfully!"
    else
        log_warning "Some architectures failed to build"
    fi
}



# Show overall file information
show_overall_file_info() {
    log_info "Overall build directory contents:"
    echo ""
    find "$BUILD_DIR" -type f -name "*.exe" -exec ls -lh {} \;
    echo ""
    log_info "Directory structure:"
    tree "$BUILD_DIR" 2>/dev/null || find "$BUILD_DIR" -type d | sort
}

# Main build process
main() {
    log_info "Starting Windows build process for $APP_NAME v$VERSION"
    log_info "Building for architectures: ${ARCHITECTURES[*]}"
    echo ""
    
    check_go
    clean_build
    download_deps
    build_all_windows
    show_overall_file_info
    
    echo ""
    log_success "Windows build process completed!"
    log_info "You can find the executables in the '$BUILD_DIR' directory"
    log_info "Each architecture is in its own subdirectory:"
    for arch in "${ARCHITECTURES[@]}"; do
        echo "  - $BUILD_DIR/windows-$arch/"
    done
}

# Run the main build process
main


