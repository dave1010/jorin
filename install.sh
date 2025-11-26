#!/usr/bin/env bash
set -euo pipefail

# detect_arch: sets global variables "os" and "arch"
detect_arch() {
  # Prefer an explicit Android detection since uname -s returns Linux on Android
  if [ -n "${ANDROID_ROOT:-}" ] || [ -e "/system/build.prop" ] || [ -e "/system/bin/getprop" ] || (uname -o 2>/dev/null || true | tr '[:upper:]' '[:lower:]' | grep -q android); then
    os=android
  else
    case "$(uname -s)" in
      Linux*) os=linux ;;
      Darwin*) os=darwin ;;
      FreeBSD*) os=freebsd ;;
      MINGW*|MSYS*|CYGWIN*) os=windows ;;
      *) echo "Unsupported OS"; exit 1 ;;
    esac
  fi

  case "$(uname -m)" in
    x86_64) arch=amd64 ;;
    aarch64|arm64) arch=arm64 ;;
    armv7*) arch=arm ;;
    i386|i686) arch=386 ;;
    *) echo "Unsupported architecture: $(uname -m)"; exit 1 ;;
  esac
}

detect_arch

asset="jorin-${os}-${arch}"
url="https://github.com/dave1010/jorin/releases/latest/download/${asset}"

echo "Downloading ${asset}..."

curl -fsSL --clobber "${url}" -o jorin
chmod +x jorin

echo "Installed ./jorin"
