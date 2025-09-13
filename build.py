#!/usr/bin/env python3
"""
Bilibili Ticket Go - Automated Build Script

This script automates the build process for bilibili-ticket-go, including:
- Dependency checks for required tools
- Git submodule initialization and updates
- Rust captcha component building
- Go application building with various modes
- Cross-platform compilation support

Usage:
    python3 build.py [options]

Options:
    --mode MODE        Build mode: dev, release, or cross (default: dev)
    --output DIR       Output directory for built binaries (default: ./dist)
    --clean           Clean build artifacts before building
    --verbose         Enable verbose output
    --check-deps      Only check dependencies and exit
    --help            Show this help message

Examples:
    python3 build.py --mode dev                    # Development build
    python3 build.py --mode release --clean        # Clean release build
    python3 build.py --mode cross --output ./bin   # Cross-platform build
    python3 build.py --check-deps                  # Check dependencies only
"""

import argparse
import logging
import os
import platform
import subprocess
import sys
import shutil
from pathlib import Path
from typing import List, Optional, Tuple


class BuildError(Exception):
    """Custom exception for build errors"""
    pass


class BuildAutomation:
    """Main build automation class"""
    
    def __init__(self, verbose: bool = False):
        self.verbose = verbose
        self.setup_logging()
        self.project_root = Path(__file__).parent.absolute()
        self.captcha_dir = self.project_root / "captcha" / "biliTicker"
        
    def setup_logging(self):
        """Setup logging configuration"""
        level = logging.DEBUG if self.verbose else logging.INFO
        logging.basicConfig(
            level=level,
            format='%(asctime)s - %(levelname)s - %(message)s',
            datefmt='%Y-%m-%d %H:%M:%S'
        )
        self.logger = logging.getLogger(__name__)
        
    def run_command(self, cmd: List[str], cwd: Optional[Path] = None, 
                   capture_output: bool = False) -> Tuple[int, str, str]:
        """Run a shell command and return exit code, stdout, stderr"""
        if cwd is None:
            cwd = self.project_root
            
        self.logger.debug(f"Running command: {' '.join(cmd)} in {cwd}")
        
        try:
            result = subprocess.run(
                cmd,
                cwd=cwd,
                capture_output=capture_output,
                text=True,
                check=False
            )
            
            if not capture_output:
                return result.returncode, "", ""
            else:
                return result.returncode, result.stdout, result.stderr
                
        except FileNotFoundError as e:
            raise BuildError(f"Command not found: {cmd[0]}. Please ensure it's installed and in PATH.")
        except Exception as e:
            raise BuildError(f"Failed to execute command: {e}")
    
    def check_command_exists(self, command: str) -> bool:
        """Check if a command exists in PATH"""
        return shutil.which(command) is not None
    
    def check_dependencies(self) -> bool:
        """Check if all required dependencies are available"""
        self.logger.info("Checking build dependencies...")
        
        required_tools = {
            'git': 'Git version control system',
            'go': 'Go programming language',
            'cargo': 'Rust package manager (Cargo)',
            'rustc': 'Rust compiler'
        }
        
        missing_tools = []
        
        for tool, description in required_tools.items():
            if self.check_command_exists(tool):
                # Get version info
                try:
                    if tool == 'git':
                        _, stdout, _ = self.run_command(['git', '--version'], capture_output=True)
                    elif tool == 'go':
                        _, stdout, _ = self.run_command(['go', 'version'], capture_output=True)
                    elif tool == 'cargo':
                        _, stdout, _ = self.run_command(['cargo', '--version'], capture_output=True)
                    elif tool == 'rustc':
                        _, stdout, _ = self.run_command(['rustc', '--version'], capture_output=True)
                    
                    version = stdout.strip().split('\n')[0] if stdout else "version unknown"
                    self.logger.info(f"✓ {tool}: {version}")
                except:
                    self.logger.info(f"✓ {tool}: available")
            else:
                missing_tools.append(f"{tool} ({description})")
                self.logger.error(f"✗ {tool}: not found")
        
        if missing_tools:
            self.logger.error("Missing required dependencies:")
            for tool in missing_tools:
                self.logger.error(f"  - {tool}")
            self.logger.error("\nPlease install the missing tools and try again.")
            return False
        
        self.logger.info("All dependencies are available!")
        return True
    
    def init_submodules(self) -> None:
        """Initialize and update git submodules"""
        self.logger.info("Initializing git submodules...")
        
        # Check if we're in a git repository
        if not (self.project_root / ".git").exists():
            raise BuildError("Not in a git repository. Please ensure you're in the project root.")
        
        # Initialize submodules
        ret_code, _, stderr = self.run_command(['git', 'submodule', 'init'], capture_output=True)
        if ret_code != 0:
            raise BuildError(f"Failed to initialize submodules: {stderr}")
        
        # Update submodules
        ret_code, _, stderr = self.run_command(['git', 'submodule', 'update', '--recursive'], capture_output=True)
        if ret_code != 0:
            raise BuildError(f"Failed to update submodules: {stderr}")
        
        self.logger.info("Submodules initialized and updated successfully")
    
    def build_rust_captcha(self) -> None:
        """Build the Rust captcha component"""
        self.logger.info("Building Rust captcha component...")
        
        if not self.captcha_dir.exists():
            self.logger.warning(f"Captcha directory not found: {self.captcha_dir}")
            self.logger.warning("Skipping Rust captcha build. The Go build may fail if it requires the captcha component.")
            return
        
        if not (self.captcha_dir / "Cargo.toml").exists():
            self.logger.warning(f"Cargo.toml not found in {self.captcha_dir}")
            self.logger.warning("Skipping Rust captcha build. The Go build may fail if it requires the captcha component.")
            return
        
        # Build the Rust component
        ret_code, _, stderr = self.run_command(
            ['cargo', 'build', '--release'],
            cwd=self.captcha_dir,
            capture_output=True
        )
        
        if ret_code != 0:
            self.logger.error(f"Rust build failed: {stderr}")
            self.logger.warning("Continuing without Rust captcha component. The Go build may fail.")
            return
        
        self.logger.info("Rust captcha component built successfully")
    
    def prepare_go_environment(self) -> None:
        """Prepare Go environment and dependencies"""
        self.logger.info("Preparing Go environment...")
        
        # Run go mod tidy
        ret_code, _, stderr = self.run_command(['go', 'mod', 'tidy'], capture_output=True)
        if ret_code != 0:
            raise BuildError(f"go mod tidy failed: {stderr}")
        
        # Install garble for obfuscated builds
        self.logger.info("Installing garble...")
        ret_code, _, stderr = self.run_command(
            ['go', 'install', 'mvdan.cc/garble@latest'],
            capture_output=True
        )
        if ret_code != 0:
            self.logger.warning(f"Failed to install garble: {stderr}")
            self.logger.warning("Continuing without garble (obfuscated builds won't be available)")
        else:
            self.logger.info("Garble installed successfully")
    
    def build_go_application(self, mode: str, output_dir: Path) -> None:
        """Build the Go application"""
        self.logger.info(f"Building Go application in {mode} mode...")
        
        # Create output directory
        output_dir.mkdir(parents=True, exist_ok=True)
        
        # Set build flags based on mode
        if mode == "dev":
            # Development build - fast compilation, debugging info
            build_cmd = ['go', 'build', '-o', str(output_dir / 'bilibili-ticket-go')]
            if platform.system() == "Windows":
                build_cmd[-1] += '.exe'
            build_cmd.append('./main.go')
            
        elif mode == "release":
            # Release build - optimized, stripped
            output_path = output_dir / 'bilibili-ticket-go'
            if platform.system() == "Windows":
                output_path = output_path.with_suffix('.exe')
                
            # Check if garble is available
            if self.check_command_exists('garble'):
                build_cmd = [
                    'garble', '-tiny', 'build',
                    '-trimpath',
                    '-ldflags', '-s -w',
                    '-o', str(output_path),
                    './main.go'
                ]
            else:
                build_cmd = [
                    'go', 'build',
                    '-trimpath',
                    '-ldflags', '-s -w',
                    '-o', str(output_path),
                    './main.go'
                ]
                
        elif mode == "cross":
            # Cross-platform build
            self.build_cross_platform(output_dir)
            return
        else:
            raise BuildError(f"Unknown build mode: {mode}")
        
        # Set CGO_ENABLED=1 as required by the project
        env = os.environ.copy()
        env['CGO_ENABLED'] = '1'
        
        # Execute the build command
        self.logger.debug(f"Build command: {' '.join(build_cmd)}")
        result = subprocess.run(build_cmd, cwd=self.project_root, env=env, capture_output=True, text=True)
        
        if result.returncode != 0:
            self.logger.error(f"Go build failed with exit code {result.returncode}")
            self.logger.error(f"stdout: {result.stdout}")
            self.logger.error(f"stderr: {result.stderr}")
            
            # Check if the error is related to missing captcha bindings
            if "bindings.h" in result.stderr or "captcha" in result.stderr.lower():
                self.logger.error("Build failed due to missing captcha component.")
                self.logger.error("This usually means the Rust submodule wasn't built successfully.")
                self.logger.error("Try running: git submodule update --init --recursive")
                
            raise BuildError(f"Go build failed with exit code {result.returncode}")
        
        self.logger.info(f"Go application built successfully in {output_dir}")
    
    def build_cross_platform(self, output_dir: Path) -> None:
        """Build for multiple platforms and architectures"""
        self.logger.info("Building for multiple platforms...")
        
        platforms = [
            ("linux", "amd64"),
            ("linux", "arm64"),
            ("windows", "amd64"),
            ("windows", "arm64"),
            ("darwin", "amd64"),
            ("darwin", "arm64")
        ]
        
        for goos, goarch in platforms:
            self.logger.info(f"Building for {goos}/{goarch}...")
            
            # Create platform-specific output directory
            platform_dir = output_dir / f"{goos}_{goarch}"
            platform_dir.mkdir(parents=True, exist_ok=True)
            
            # Set output filename
            binary_name = "bilibili-ticket-go"
            if goos == "windows":
                binary_name += ".exe"
            output_path = platform_dir / binary_name
            
            # Set environment variables
            env = os.environ.copy()
            env.update({
                'GOOS': goos,
                'GOARCH': goarch,
                'CGO_ENABLED': '1'
            })
            
            # Build command
            build_cmd = [
                'go', 'build',
                '-trimpath',
                '-ldflags', '-s -w',
                '-o', str(output_path),
                './main.go'
            ]
            
            # Execute build
            result = subprocess.run(build_cmd, cwd=self.project_root, env=env)
            
            if result.returncode == 0:
                self.logger.info(f"✓ Built for {goos}/{goarch}")
            else:
                self.logger.error(f"✗ Failed to build for {goos}/{goarch}")
        
        self.logger.info("Cross-platform build completed")
    
    def clean_artifacts(self, output_dir: Path) -> None:
        """Clean build artifacts"""
        self.logger.info("Cleaning build artifacts...")
        
        # Remove output directory
        if output_dir.exists():
            shutil.rmtree(output_dir)
            self.logger.info(f"Removed {output_dir}")
        
        # Clean Go build cache
        ret_code, _, _ = self.run_command(['go', 'clean', '-cache'], capture_output=True)
        if ret_code == 0:
            self.logger.info("Cleaned Go build cache")
        
        # Clean Rust build artifacts
        if self.captcha_dir.exists():
            rust_target_dir = self.captcha_dir / "target"
            if rust_target_dir.exists():
                shutil.rmtree(rust_target_dir)
                self.logger.info("Cleaned Rust build artifacts")
    
    def build(self, mode: str, output_dir: Path, clean: bool = False, skip_captcha: bool = False) -> None:
        """Main build process"""
        try:
            self.logger.info(f"Starting build process (mode: {mode})")
            
            # Clean if requested
            if clean:
                self.clean_artifacts(output_dir)
            
            # Check dependencies
            if not self.check_dependencies():
                raise BuildError("Dependency check failed")
            
            # Initialize submodules
            self.init_submodules()
            
            # Build Rust captcha component (unless skipped)
            if not skip_captcha:
                self.build_rust_captcha()
            else:
                self.logger.info("Skipping Rust captcha component build")
            
            # Prepare Go environment
            self.prepare_go_environment()
            
            # Build Go application
            self.build_go_application(mode, output_dir)
            
            self.logger.info("Build completed successfully!")
            
        except BuildError as e:
            self.logger.error(f"Build failed: {e}")
            sys.exit(1)
        except KeyboardInterrupt:
            self.logger.info("Build interrupted by user")
            sys.exit(130)
        except Exception as e:
            self.logger.error(f"Unexpected error: {e}")
            sys.exit(1)


def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(
        description="Automated build script for bilibili-ticket-go",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__
    )
    
    parser.add_argument(
        '--mode',
        choices=['dev', 'release', 'cross'],
        default='dev',
        help='Build mode (default: dev)'
    )
    
    parser.add_argument(
        '--output',
        type=Path,
        default=Path('./dist'),
        help='Output directory for built binaries (default: ./dist)'
    )
    
    parser.add_argument(
        '--clean',
        action='store_true',
        help='Clean build artifacts before building'
    )
    
    parser.add_argument(
        '--verbose',
        action='store_true',
        help='Enable verbose output'
    )
    
    parser.add_argument(
        '--check-deps',
        action='store_true',
        help='Only check dependencies and exit'
    )
    
    parser.add_argument(
        '--skip-captcha',
        action='store_true',
        help='Skip building the Rust captcha component'
    )
    
    args = parser.parse_args()
    
    # Initialize build automation
    builder = BuildAutomation(verbose=args.verbose)
    
    # Check dependencies only if requested
    if args.check_deps:
        if builder.check_dependencies():
            print("All dependencies are available!")
            sys.exit(0)
        else:
            print("Some dependencies are missing!")
            sys.exit(1)
    
    # Run the build process
    builder.build(args.mode, args.output, args.clean, args.skip_captcha)


if __name__ == '__main__':
    main()