# Build Automation Script

This Python script (`build.py`) provides automated building capabilities for the bilibili-ticket-go project.

## Features

- **Dependency Checking**: Validates that all required tools (Go, Rust, Git) are installed
- **Submodule Management**: Automatically initializes and updates git submodules
- **Rust Component Building**: Builds the required captcha component written in Rust
- **Go Application Building**: Builds the main Go application with multiple modes
- **Cross-Platform Support**: Can build for multiple operating systems and architectures
- **Clean Builds**: Option to clean previous build artifacts
- **Flexible Output**: Configurable output directories

## Prerequisites

Before using the build script, ensure you have the following installed:

- **Python 3.6+**
- **Go 1.18+** (for building the Go application)
- **Rust/Cargo** (for building the captcha component)
- **Git** (for submodule management)

## Usage

### Basic Usage

```bash
# Development build (default)
python3 build.py

# Release build with optimizations
python3 build.py --mode release

# Cross-platform build for all supported platforms
python3 build.py --mode cross

# Clean build (removes previous artifacts)
python3 build.py --mode release --clean
```

### Advanced Usage

```bash
# Build with custom output directory
python3 build.py --mode release --output ./bin

# Skip the Rust captcha component (if having issues)
python3 build.py --skip-captcha

# Verbose output for debugging
python3 build.py --mode dev --verbose

# Check dependencies only
python3 build.py --check-deps
```

### Command-Line Options

| Option | Description |
|--------|-------------|
| `--mode {dev,release,cross}` | Build mode (default: dev) |
| `--output DIR` | Output directory for built binaries (default: ./dist) |
| `--clean` | Clean build artifacts before building |
| `--verbose` | Enable verbose output with detailed logging |
| `--check-deps` | Only check dependencies and exit |
| `--skip-captcha` | Skip building the Rust captcha component |
| `--help` | Show help message with all options |

## Build Modes

### Development Mode (`--mode dev`)
- Fast compilation
- Includes debugging information
- No optimizations
- Single platform build (current system)

### Release Mode (`--mode release`)
- Optimized compilation
- Stripped binaries (smaller size)
- Uses `garble` for code obfuscation if available
- Single platform build (current system)

### Cross-Platform Mode (`--mode cross`)
- Builds for multiple platforms:
  - Linux (amd64, arm64)
  - Windows (amd64, arm64)  
  - macOS (amd64, arm64)
- Optimized compilation
- Creates separate directories for each platform

## Output Structure

```
dist/                          # Default output directory
├── bilibili-ticket-go         # Development/Release build
├── linux_amd64/              # Cross-platform builds
│   └── bilibili-ticket-go
├── linux_arm64/
│   └── bilibili-ticket-go
├── windows_amd64/
│   └── bilibili-ticket-go.exe
├── windows_arm64/
│   └── bilibili-ticket-go.exe
├── darwin_amd64/
│   └── bilibili-ticket-go
└── darwin_arm64/
    └── bilibili-ticket-go
```

## Troubleshooting

### Common Issues

1. **Missing Dependencies**
   ```bash
   python3 build.py --check-deps
   ```
   This will show which tools are missing.

2. **Submodule Issues**
   If the captcha component fails to build, try:
   ```bash
   git submodule update --init --recursive
   python3 build.py --skip-captcha
   ```

3. **Go Build Failures**
   Check that CGO is properly configured and all dependencies are available:
   ```bash
   go mod tidy
   go mod download
   ```

4. **Rust Build Failures**
   Ensure Rust is properly installed and updated:
   ```bash
   rustup update
   cargo --version
   ```

### Error Messages

- **"bindings.h: No such file or directory"**: The Rust captcha component hasn't been built. Use `--skip-captcha` or ensure the submodule is properly initialized.
- **"Command not found"**: A required tool is not installed or not in PATH.
- **"Go build failed"**: Check Go version and dependencies with `go mod tidy`.

## Integration with CI/CD

The build script can be easily integrated into CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Build Application
  run: |
    python3 build.py --mode release --clean --verbose
    
# For cross-platform builds
- name: Cross-Platform Build
  run: |
    python3 build.py --mode cross --output ./release --clean
```

## Development Notes

- The script automatically sets `CGO_ENABLED=1` as required by the project
- Uses `garble` for code obfuscation in release builds if available
- Supports both development and production build configurations
- Handles platform-specific binary naming (`.exe` for Windows)
- Provides detailed logging for debugging build issues

## Contributing

If you encounter issues or want to add features to the build script:

1. Test your changes with different build modes
2. Ensure backward compatibility
3. Add appropriate error handling
4. Update this documentation if needed