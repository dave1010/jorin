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

select_install_dir() {
  if [ -n "${JORIN_INSTALL_DIR:-}" ]; then
    install_dir="$JORIN_INSTALL_DIR"
    return
  fi

  if [ -n "${PREFIX:-}" ] && [ -w "${PREFIX}/bin" ]; then
    install_dir="${PREFIX}/bin"
    return
  fi

  if [ "$(id -u)" -eq 0 ]; then
    install_dir="/usr/local/bin"
    return
  fi

  if [ -w "/usr/local/bin" ]; then
    install_dir="/usr/local/bin"
    return
  fi

  if [ -w "/opt/homebrew/bin" ]; then
    install_dir="/opt/homebrew/bin"
    return
  fi

  install_dir="${HOME}/.local/bin"
}

profile_for_shell() {
  if [ -n "${JORIN_PROFILE:-}" ]; then
    profile="$JORIN_PROFILE"
    profile_shell="sh"
    return
  fi

  case "$(basename "${SHELL:-}" )" in
    zsh)
      profile="${HOME}/.zshrc"
      profile_shell="sh"
      ;;
    bash)
      profile="${HOME}/.bashrc"
      profile_shell="sh"
      ;;
    fish)
      profile="${HOME}/.config/fish/config.fish"
      profile_shell="fish"
      ;;
    *)
      profile="${HOME}/.profile"
      profile_shell="sh"
      ;;
  esac
}

ensure_path() {
  if [[ ":${PATH}:" == *":${install_dir}:"* ]]; then
    return
  fi

  if [[ "${install_dir}" != "${HOME}/"* ]]; then
    return
  fi

  profile_for_shell

  if [ "${profile_shell}" = "fish" ]; then
    path_line="set -gx PATH \"${install_dir}\" \$PATH"
    mkdir -p "$(dirname "${profile}")"
  else
    path_line="export PATH=\"${install_dir}:\$PATH\""
  fi

  if [ -f "${profile}" ] && grep -Fxq "${path_line}" "${profile}"; then
    return
  fi

  {
    printf '\n# Added by Jorin installer\n'
    printf '%s\n' "${path_line}"
  } >> "${profile}"

  echo "Added ${install_dir} to PATH in ${profile}."
  echo "Run 'source ${profile}' or start a new shell to use jorin."
}

detect_arch
select_install_dir

asset="jorin-${os}-${arch}"
url="https://github.com/dave1010/jorin/releases/latest/download/${asset}"

tmp_file="$(mktemp -t jorin.XXXXXX)"
trap 'rm -f "${tmp_file}"' EXIT

echo "Downloading ${asset}..."

curl -fsSL --clobber "${url}" -o "${tmp_file}"
chmod +x "${tmp_file}"

mkdir -p "${install_dir}"
install -m 0755 "${tmp_file}" "${install_dir}/jorin"

ensure_path

echo "Installed ${install_dir}/jorin"
