#!/usr/bin/env bash
set -euo pipefail

REPO="ASDFGHoney/planosh"
INSTALL_DIR="${HOME}/.local/bin"
BINARY="planosh"

# --- detect OS ---
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${OS}" in
  darwin|linux) ;;
  *) echo "Error: unsupported OS: ${OS}" >&2; exit 1 ;;
esac

# --- detect ARCH ---
ARCH="$(uname -m)"
case "${ARCH}" in
  x86_64|amd64)  ARCH="amd64" ;;
  arm64|aarch64)  ARCH="arm64" ;;
  *) echo "Error: unsupported architecture: ${ARCH}" >&2; exit 1 ;;
esac

# --- resolve version ---
if [ -n "${PLANOSH_VERSION:-}" ]; then
  VERSION="${PLANOSH_VERSION}"
else
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | head -1 | cut -d'"' -f4)"
fi

if [ -z "${VERSION}" ]; then
  echo "Error: failed to determine latest version" >&2
  exit 1
fi

# --- download & install ---
ARCHIVE="${BINARY}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}"' EXIT

echo "Downloading ${BINARY} ${VERSION} (${OS}/${ARCH})..."
curl -fsSL "${URL}" -o "${TMPDIR}/${ARCHIVE}"

tar -xzf "${TMPDIR}/${ARCHIVE}" -C "${TMPDIR}"

mkdir -p "${INSTALL_DIR}"
mv "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
chmod +x "${INSTALL_DIR}/${BINARY}"

echo "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"

# --- PATH hint ---
if ! echo "${PATH}" | tr ':' '\n' | grep -qx "${INSTALL_DIR}"; then
  echo ""
  echo "Add to your shell profile:"
  echo "  export PATH=\"${INSTALL_DIR}:\${PATH}\""
fi
