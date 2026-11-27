#!/usr/bin/env bash
# Install the mf-digisac plugin from GitHub releases and register it with mf.
#
#   curl -fsSL https://raw.githubusercontent.com/mguilarducci/mf-digisac/main/scripts/get-digisac.sh | bash
#
# Adapted from entireio/entire-brain's scripts/get-brain.sh (MIT, see NOTICE).
set -euo pipefail

GITHUB_REPO="mguilarducci/mf-digisac"
BINARY="mf-digisac"
INSTALL_DIR="${MF_DIGISAC_INSTALL_DIR:-$HOME/.local/bin}"
WANT_VERSION=""

if [ -t 1 ] && [ -z "${NO_COLOR:-}" ]; then
  BOLD=$'\033[1m'; GREEN=$'\033[0;32m'; RED=$'\033[0;31m'; YELLOW=$'\033[0;33m'; NC=$'\033[0m'
else
  BOLD=""; GREEN=""; RED=""; YELLOW=""; NC=""
fi
info()  { printf '%s==>%s %s\n' "$GREEN" "$NC" "$*"; }
warn()  { printf '%s!%s   %s\n' "$YELLOW" "$NC" "$*" >&2; }
error() { printf '%serror:%s %s\n' "$RED" "$NC" "$*" >&2; exit 1; }

usage() {
  cat <<USAGE
Install the mf-digisac plugin.

  curl -fsSL https://raw.githubusercontent.com/${GITHUB_REPO}/main/scripts/get-digisac.sh | bash

Options:
  --version vX.Y.Z   Install a specific release (default: latest)
  --dir <path>       Install somewhere other than ~/.local/bin
  -h, --help         Show this message

The download is verified against the release checksums.
USAGE
}

while [ $# -gt 0 ]; do
  case "$1" in
    --version) WANT_VERSION="${2:?--version needs a value}"; shift 2 ;;
    --dir)     INSTALL_DIR="${2:?--dir needs a path}"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) error "unknown option: $1" ;;
  esac
done

for c in curl tar uname; do
  command -v "$c" >/dev/null || error "missing required command: $c"
done

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$os" in
  linux|darwin) : ;;
  *) error "unsupported OS: $os. Windows users: download the .zip from
  https://github.com/${GITHUB_REPO}/releases/latest" ;;
esac

arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) error "unsupported architecture: $arch" ;;
esac

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

version="$WANT_VERSION"
if [ -z "$version" ]; then
  status="$(curl -sL -o "${tmp}/api.json" -w '%{http_code}' \
    "https://api.github.com/repos/${GITHUB_REPO}/releases/latest")" || status="000"
  case "$status" in
    2??) : ;;
    403|429) error "GitHub's API rate limit is exhausted (HTTP ${status}). Wait and re-run, or pass --version vX.Y.Z." ;;
    *) error "could not resolve the latest release (HTTP ${status}). Pass --version vX.Y.Z.
  Releases: https://github.com/${GITHUB_REPO}/releases" ;;
  esac
  version="$({ grep -o '"tag_name": *"[^"]*"' "${tmp}/api.json" || true; } \
    | head -1 | sed 's/.*"\([^"]*\)"$/\1/')"
  [ -n "$version" ] || error "the GitHub API response named no release tag. Pass --version vX.Y.Z."
fi

info "Installing mf-digisac ${BOLD}${version}${NC} (${os}/${arch})"

asset="${BINARY}_${version#v}_${os}_${arch}.tar.gz"
base="https://github.com/${GITHUB_REPO}/releases/download/${version}"

curl -fsSL -o "${tmp}/${asset}" "${base}/${asset}" \
  || error "no build for ${os}/${arch} in ${version}
  Assets: https://github.com/${GITHUB_REPO}/releases/tag/${version}"
curl -fsSL -o "${tmp}/checksums.txt" "${base}/checksums.txt" \
  || error "could not fetch checksums.txt for ${version}; refusing to install an unverified binary."

if command -v sha256sum >/dev/null; then
  have="$(cd "$tmp" && sha256sum "$asset" | awk '{print $1}')"
elif command -v shasum >/dev/null; then
  have="$(cd "$tmp" && shasum -a 256 "$asset" | awk '{print $1}')"
else
  error "no sha256sum or shasum on PATH, so the download cannot be verified."
fi
want="$({ grep " ${asset}\$" "${tmp}/checksums.txt" || true; } | awk '{print $1}' | head -1)"
[ -n "$want" ] || error "checksums.txt has no entry for ${asset}"
[ "$have" = "$want" ] || error "checksum mismatch for ${asset}
  expected ${want}
  got      ${have}"
info "checksum verified"

tar -xzf "${tmp}/${asset}" -C "$tmp"
[ -f "${tmp}/${BINARY}" ] || error "archive did not contain ${BINARY}"
chmod +x "${tmp}/${BINARY}"
mkdir -p "$INSTALL_DIR"
mv "${tmp}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
info "installed to ${BOLD}${INSTALL_DIR}/${BINARY}${NC}"

if command -v mf >/dev/null; then
  if mf plugin install "${INSTALL_DIR}/${BINARY}" >/dev/null 2>&1; then
    info "registered with the mf CLI"
    mf digisac version 2>/dev/null || true
    printf '%sNext:%s  mf digisac auth login\n' "$BOLD" "$NC"
  else
    warn "could not register the plugin automatically. Run:
  mf plugin install ${INSTALL_DIR}/${BINARY}"
  fi
else
  warn "the mf CLI was not found on PATH. Install it first:
  curl -fsSL https://raw.githubusercontent.com/mguilarducci/mf-cli/main/scripts/install.sh | bash
then run:
  mf plugin install ${INSTALL_DIR}/${BINARY}"
fi

case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) : ;;
  *) warn "${INSTALL_DIR} is not on your PATH. Add:
  export PATH=\"${INSTALL_DIR}:\$PATH\"" ;;
esac
