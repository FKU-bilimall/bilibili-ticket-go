import os
import platform
import subprocess
import sys
import argparse
import zipfile
import shutil
import datetime

GO_MAIN = "main.go"
GO_OUTPUT = "bilibili-ticket-go"
ARCH_MAP = {
    "x86_64": "amd64",
    "aarch64": "arm64",
    "arm64": "arm64"
}
OS_MAP = {
    "linux": "linux",
    "windows": "windows",
    "darwin": "darwin"
}

def run(cmd, env=None, cwd=None):
    print(f"==> {cmd}")
    result = subprocess.run(cmd, shell=True, env=env, cwd=cwd)
    if result.returncode != 0:
        sys.exit(result.returncode)

def parse_target(target):
    parts = target.split('-')
    arch = ARCH_MAP.get(parts[0], parts[0])
    osname = OS_MAP.get(parts[2], parts[2])
    return osname, arch

def detect_platform_arch():
    sys_platform = platform.system().lower()
    machine = platform.machine().lower()
    goos = OS_MAP.get(sys_platform, sys_platform)
    goarch = ARCH_MAP.get(machine, machine)
    return goos, goarch

def build_go(goos, goarch, out_dir, ldflags=None):
    env = os.environ.copy()
    env["GOOS"] = goos
    env["GOARCH"] = goarch
    env["CGO_ENABLED"] = "1"
    out_name = GO_OUTPUT + (".exe" if goos == "windows" else "")
    out_path = os.path.join(out_dir, out_name)
    os.makedirs(out_dir, exist_ok=True)
    ldflags_str = f'-ldflags "{ldflags}"' if ldflags else ''
    run(f"go build {ldflags_str} -o {out_path} {GO_MAIN}", env=env)
    print(f"Go build finished: {out_path}")
    return out_path

def zip_with_deps(output_path, deps_path=None, zip_dir=None, zip_name=None):
    if zip_dir is None:
        zip_dir = os.path.dirname(output_path)
    if zip_name is None:
        zip_name = os.path.splitext(os.path.basename(output_path))[0] + ".zip"
    zip_path = os.path.join(zip_dir, zip_name)
    lib_exts = [".so", ".dll", ".dylib"]
    with zipfile.ZipFile(zip_path, "w", zipfile.ZIP_DEFLATED) as zipf:
        # Add the executable to the zip
        zipf.write(output_path, os.path.basename(output_path))
        # Add dependency libraries
        if deps_path and os.path.isdir(deps_path):
            for file in os.listdir(deps_path):
                abs_path = os.path.join(deps_path, file)
                if os.path.isfile(abs_path) and any(file.endswith(ext) for ext in lib_exts):
                    zipf.write(abs_path, file)
    print(f"Packaged as: {zip_path}")


def get_git_commit():
    try:
        result = subprocess.run(["git", "rev-parse", "--short", "HEAD"], stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=True, text=True)
        return result.stdout.strip()
    except Exception:
        return ""

def get_build_time():
    return str(int(datetime.datetime.utcnow().timestamp()))

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--target", help="Target triple (e.g. x86_64-unknown-linux-gnu)")
    parser.add_argument("--os", help="Target OS (windows/linux/darwin)")
    parser.add_argument("--arch", help="Target architecture (amd64/arm64)")
    parser.add_argument("--deps", help="Dependency directory path (optional)")
    parser.add_argument("--outdir", help="Output directory (optional, default: output)")
    parser.add_argument("--commit", help="Git short commit hash for ldflags")
    parser.add_argument("--buildtime", help="Build timestamp for ldflags")
    args = parser.parse_args()

    if args.target:
        goos, goarch = parse_target(args.target)
    elif args.os and args.arch:
        goos = args.os
        goarch = args.arch
    else:
        goos, goarch = detect_platform_arch()
    project_name = GO_OUTPUT
    out_dir = args.outdir or "output"
    os.makedirs(out_dir, exist_ok=True)
    print(f"Build target: {goos}, arch: {goarch}, output dir: {out_dir}")

    # Auto get commit/buildtime if not provided
    commit = getattr(args, 'commit', None) or get_git_commit()
    buildtime = getattr(args, 'buildtime', None) or get_build_time()

    ldflags = f'-s -w -X "bilibili-ticket-go/global.GitCommit={commit}" -X "bilibili-ticket-go/global.BuildTime={buildtime}" -X "bilibili-ticket-go/global.LoggerLevel=4"'
    out_path = build_go(goos, goarch, out_dir, ldflags=ldflags)
    # zip file name: {{ProjectName}}_{{Os}}_{{Arch}}.zip, output to output dir (no subfolder)
    zip_name = f"{project_name}_{goos}_{goarch}.zip"
    zip_with_deps(out_path, args.deps, zip_dir=out_dir, zip_name=zip_name)
    # Remove the executable after packaging
    if os.path.exists(out_path):
        os.remove(out_path)
    print("Build and packaging complete!")
if __name__ == "__main__":
    main()