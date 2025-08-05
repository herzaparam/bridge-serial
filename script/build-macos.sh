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

# Check if running on macOS
check_macos() {
    if [[ "$OSTYPE" != "darwin"* ]]; then
        log_warning "This script is designed for macOS. Running on: $OSTYPE"
        log_info "The script will still attempt to build, but results may vary."
    else
        log_info "Running on macOS: $(sw_vers -productName) $(sw_vers -productVersion)"
    fi
}

# Clean previous builds
clean_build() {
    log_info "Cleaning previous macOS builds..."
    
    # Create main build directory if it doesn't exist
    mkdir -p "$BUILD_DIR"
    
    # Remove only macOS-specific directories
    for arch in "${ARCHITECTURES[@]}"; do
        local arch_dir="${BUILD_DIR}/macos-${arch}"
        if [ -d "$arch_dir" ]; then
            log_info "Removing $arch_dir"
            rm -rf "$arch_dir"
        fi
    done
    
    # Remove universal binary directory
    local universal_dir="${BUILD_DIR}/macos-universal"
    if [ -d "$universal_dir" ]; then
        log_info "Removing $universal_dir"
        rm -rf "$universal_dir"
    fi
}

# Download dependencies
download_deps() {
    log_info "Downloading dependencies..."
    go mod download
    go mod tidy
}

# Build for specific macOS architecture
build_macos_arch() {
    local arch=$1
    local output_name="${APP_NAME}-${arch}"
    local arch_dir="${BUILD_DIR}/macos-${arch}"
    
    log_info "Building for macOS ${arch}..."
    
    # Create architecture-specific directory
    mkdir -p "$arch_dir"
    
    # Set environment variables for cross-compilation
    export GOOS=darwin
    export GOARCH="$arch"
    export CGO_ENABLED=1  # Enable CGO for Fyne/OpenGL
    export FYNE_SCALE=1.0
    
    # Build the application
    go build \
        -ldflags="-s -w -X main.version=$VERSION" \
        -o "$arch_dir/$output_name" \
        "$ENTRY_POINT"
    
    if [ $? -eq 0 ]; then
        log_success "macOS ${arch} build completed successfully!"
        log_info "Output: $arch_dir/$output_name"
        
        # Create build info for this architecture
        create_build_info_arch "$arch" "$arch_dir"
        
        # Show file information
        show_file_info_arch "$arch" "$arch_dir" "$output_name"
        
        # Make executable
        chmod +x "$arch_dir/$output_name"
    else
        log_error "macOS ${arch} build failed!"
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
Architecture: macOS ${arch}
Build Date: $(date -u '+%Y-%m-%d %H:%M:%S UTC')
Build Platform: $(uname -s) $(uname -m)
Go Version: $(go version)
Target Platform: macOS ${arch}
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
        
        # Show file type information
        log_info "File type information:"
        file "$arch_dir/$output_name"
    fi
}

# Build for all macOS architectures
build_all_macos() {
    log_info "Building for all macOS architectures..."
    
    local success_count=0
    local total_count=${#ARCHITECTURES[@]}
    
    for arch in "${ARCHITECTURES[@]}"; do
        echo ""
        if build_macos_arch "$arch"; then
            ((success_count++))
        else
            log_warning "Failed to build for ${arch}"
        fi
    done
    
    echo ""
    log_info "Build Summary: $success_count/$total_count architectures built successfully"
    
    if [ $success_count -eq $total_count ]; then
        log_success "All macOS architectures built successfully!"
    else
        log_warning "Some architectures failed to build"
    fi
}

# Create universal binary (if both architectures built successfully)
create_universal_binary() {
    local amd64_binary="${BUILD_DIR}/macos-amd64/${APP_NAME}-amd64"
    local arm64_binary="${BUILD_DIR}/macos-arm64/${APP_NAME}-arm64"
    local universal_binary="${BUILD_DIR}/macos-universal/${APP_NAME}"
    
    if [ -f "$amd64_binary" ] && [ -f "$arm64_binary" ]; then
        log_info "Creating universal binary..."
        
        mkdir -p "${BUILD_DIR}/macos-universal"
        
        # Create universal binary using lipo
        lipo -create "$amd64_binary" "$arm64_binary" -output "$universal_binary"
        
        if [ $? -eq 0 ]; then
            log_success "Universal binary created successfully!"
            log_info "Output: $universal_binary"
            
            # Make executable
            chmod +x "$universal_binary"
            
            # Show file information
            log_info "Universal binary information:"
            ls -lh "$universal_binary"
            file "$universal_binary"
            
            # Create build info for universal binary
            create_universal_build_info
        else
            log_error "Failed to create universal binary!"
        fi
    else
        log_warning "Cannot create universal binary - missing one or both architecture builds"
    fi
}

# Create build info for universal binary
create_universal_build_info() {
    local universal_dir="${BUILD_DIR}/macos-universal"
    
    log_info "Creating build information for universal binary..."
    
    cat > "$universal_dir/build-info.txt" << EOF
Build Information
=================
Application: $APP_NAME
Version: $VERSION
Architecture: macOS Universal (AMD64 + ARM64)
Build Date: $(date -u '+%Y-%m-%d %H:%M:%S UTC')
Build Platform: $(uname -s) $(uname -m)
Go Version: $(go version)
Target Platform: macOS Universal
Entry Point: $ENTRY_POINT

This universal binary contains both AMD64 and ARM64 architectures
and will run natively on both Intel and Apple Silicon Macs.
EOF

    log_success "Universal binary build information saved to $universal_dir/build-info.txt"
}



# Show overall file information
show_overall_file_info() {
    log_info "Overall build directory contents:"
    echo ""
    find "$BUILD_DIR" -type f -name "${APP_NAME}*" -exec ls -lh {} \;
    echo ""
    log_info "Directory structure:"
    tree "$BUILD_DIR" 2>/dev/null || find "$BUILD_DIR" -type d | sort
}

# Main build process
main() {
    log_info "Starting macOS build process for $APP_NAME v$VERSION"
    log_info "Building for architectures: ${ARCHITECTURES[*]}"
    echo ""
    
    check_go
    check_macos
    clean_build
    download_deps
    build_all_macos
    create_universal_binary
    show_overall_file_info
    
    echo ""
    log_success "macOS build process completed!"
    log_info "You can find the executables in the '$BUILD_DIR' directory"
    log_info "Each architecture is in its own subdirectory:"
    for arch in "${ARCHITECTURES[@]}"; do
        echo "  - $BUILD_DIR/macos-$arch/"
    done
    echo "  - $BUILD_DIR/macos-universal/ (universal binary)"
}

# Run the main build process
main 