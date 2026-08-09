#!/bin/sh
#-----------------------------------------------------------------------------------------------------------------------
# Bluefin CLI Devcontainer Feature Install Script
#
# Installs the bluefin-cli binary from GitHub releases.
#
# Options (set via devcontainer-feature.json):
#   VERSION  - bluefin-cli version to install (default: latest)
#-----------------------------------------------------------------------------------------------------------------------
set -eu

VERSION="${VERSION:-latest}"
REPO="tuna-os/bluefin-cli"
BINARY="bluefin-cli"
BIN_DIR="/usr/local/bin"

# Resolve the tag to install.
if [ "${VERSION}" = "latest" ]; then
    echo "[bluefin-cli] Resolving latest release..."
    tag=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" |
        grep '"tag_name"' | head -1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
    if [ -z "${tag}" ]; then
        echo "[bluefin-cli] ERROR: could not determine latest release" >&2
        exit 1
    fi
else
    tag="v${VERSION#v}"
fi

version="${tag#v}"

# Detect OS and architecture.
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "${os}" in
    linux|darwin) ;;
    *)
        echo "[bluefin-cli] ERROR: unsupported OS: ${os}" >&2
        exit 1
        ;;
esac

arch=$(uname -m)
case "${arch}" in
    x86_64|amd64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    *)
        echo "[bluefin-cli] ERROR: unsupported architecture: ${arch}" >&2
        exit 1
        ;;
esac

# Download and install.
url="https://github.com/${REPO}/releases/download/${tag}/${BINARY}_${version}_${os}_${arch}.tar.gz"
echo "[bluefin-cli] Downloading ${BINARY} ${tag} (${os}/${arch})..."
echo "[bluefin-cli] URL: ${url}"

tmp=$(mktemp -d)
trap 'rm -rf "${tmp}"' EXIT

curl -fsSL "${url}" -o "${tmp}/bluefin-cli.tar.gz"
tar -xzf "${tmp}/bluefin-cli.tar.gz" -C "${tmp}"

if [ ! -f "${tmp}/${BINARY}" ]; then
    echo "[bluefin-cli] ERROR: ${BINARY} not found in archive" >&2
    exit 1
fi

mkdir -p "${BIN_DIR}"
install -m 0755 "${tmp}/${BINARY}" "${BIN_DIR}/${BINARY}"

echo "[bluefin-cli] Installed ${BINARY} ${tag} to ${BIN_DIR}/${BINARY}"
echo "[bluefin-cli] Run 'bluefin-cli' to get started."
