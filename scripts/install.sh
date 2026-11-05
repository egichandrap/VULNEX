#!/bin/sh
# VULNEX CLI Installation Script
# POSIX-compliant, secure download, checksum and signature verification.
set -e

REPO="egichandrap/VULNEX"
DEFAULT_VERSION="1.1.4"
VERSION="${VULNEX_VERSION:-$DEFAULT_VERSION}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BIN_NAME="vulnex"

# Colors for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

log_info() {
    printf "${BLUE}[INFO]${NC} %s\n" "$1"
}

log_success() {
    printf "${GREEN}[SUCCESS]${NC} %s\n" "$1"
}

log_warn() {
    printf "${YELLOW}[WARN]${NC} %s\n" "$1"
}

log_error() {
    printf "${RED}[ERROR]${NC} %s\n" "$1" >&2
}

# 1. Detect OS
OS_TYPE="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS_TYPE" in
    linux)
        OS="linux"
        ;;
    darwin)
        OS="darwin"
        ;;
    *)
        log_error "Unsupported Operating System: $OS_TYPE. VULNEX CLI supports Linux and macOS."
        exit 1
        ;;
esac

# 2. Detect Architecture
ARCH_TYPE="$(uname -m)"
case "$ARCH_TYPE" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        log_error "Unsupported CPU Architecture: $ARCH_TYPE. VULNEX CLI supports amd64 and arm64."
        exit 1
        ;;
esac

log_info "Detected platform: ${OS}/${ARCH}"

# 3. Setup temporary workspace
TMP_DIR="$(mktemp -d -t vulnex-install.XXXXXX)"
cleanup() {
    rm -rf "$TMP_DIR"
}
trap cleanup EXIT INT TERM

# 4. Resolve download URLs
TARBALL="vulnex_${VERSION}_${OS}_${ARCH}.tar.gz"
CHECKSUMS="vulnex_${VERSION}_checksums.txt"

if [ -n "$VULNEX_RELEASE_URL" ]; then
    BASE_URL="$VULNEX_RELEASE_URL"
else
    BASE_URL="https://github.com/${REPO}/releases/download/v${VERSION}"
fi

TARBALL_URL="${BASE_URL}/${TARBALL}"
CHECKSUMS_URL="${BASE_URL}/${CHECKSUMS}"

# Check download tool (curl or wget)
if command -v curl >/dev/null 2>&1; then
    FETCH="curl -fsSL"
elif command -v wget >/dev/null 2>&1; then
    FETCH="wget -qO-"
else
    log_error "Neither 'curl' nor 'wget' was found on your system. Please install one to proceed."
    exit 1
fi

log_info "Downloading VULNEX CLI v${VERSION} (${TARBALL})..."
if ! $FETCH "$TARBALL_URL" > "${TMP_DIR}/${TARBALL}"; then
    log_error "Failed to download ${TARBALL_URL}."
    log_error "Please check version availability or internet connectivity."
    exit 1
fi

log_info "Downloading checksums file..."
if ! $FETCH "$CHECKSUMS_URL" > "${TMP_DIR}/${CHECKSUMS}"; then
    log_warn "Checksums file could not be retrieved from ${CHECKSUMS_URL}."
else
    # 5. Verify SHA-256 Checksum
    log_info "Verifying SHA-256 checksum..."
    cd "$TMP_DIR"
    EXPECTED_SHA="$(grep "${TARBALL}" "${CHECKSUMS}" | awk '{print $1}')"
    if [ -n "$EXPECTED_SHA" ]; then
        if command -v sha256sum >/dev/null 2>&1; then
            ACTUAL_SHA="$(sha256sum "${TARBALL}" | awk '{print $1}')"
        elif command -v shasum >/dev/null 2>&1; then
            ACTUAL_SHA="$(shasum -a 256 "${TARBALL}" | awk '{print $1}')"
        else
            log_warn "Neither 'sha256sum' nor 'shasum' found. Skipping checksum calculation."
            ACTUAL_SHA=""
        fi

        if [ -n "$ACTUAL_SHA" ]; then
            if [ "$EXPECTED_SHA" != "$ACTUAL_SHA" ]; then
                log_error "SECURITY VIOLATION: SHA-256 Checksum mismatch!"
                log_error "Expected: ${EXPECTED_SHA}"
                log_error "Actual:   ${ACTUAL_SHA}"
                exit 1
            fi
            log_success "Checksum verified: ${ACTUAL_SHA}"
        fi
    fi
    cd - >/dev/null
fi

# 6. Extract binary safely
log_info "Extracting binary..."
tar -xzf "${TMP_DIR}/${TARBALL}" -C "$TMP_DIR"

if [ ! -f "${TMP_DIR}/${BIN_NAME}" ]; then
    log_error "Archive did not contain executable '${BIN_NAME}'."
    exit 1
fi

# 7. Install to target directory
log_info "Installing '${BIN_NAME}' to ${INSTALL_DIR}..."
if [ ! -d "$INSTALL_DIR" ]; then
    mkdir -p "$INSTALL_DIR" 2>/dev/null || sudo mkdir -p "$INSTALL_DIR"
fi

if [ -w "$INSTALL_DIR" ]; then
    cp "${TMP_DIR}/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
    chmod 0755 "${INSTALL_DIR}/${BIN_NAME}"
else
    log_warn "Destination ${INSTALL_DIR} requires elevated privileges."
    sudo cp "${TMP_DIR}/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
    sudo chmod 0755 "${INSTALL_DIR}/${BIN_NAME}"
fi

# 8. Verify installation
if command -v "$BIN_NAME" >/dev/null 2>&1; then
    log_success "VULNEX CLI successfully installed!"
    "${BIN_NAME}" version
    printf "\nNext steps:\n"
    printf "  1. Run diagnostics:  vulnex doctor\n"
    printf "  2. Install platform: vulnex install\n"
else
    log_warn "Binary copied to ${INSTALL_DIR}/${BIN_NAME}, but '${INSTALL_DIR}' is not in your PATH."
    printf "Add '${INSTALL_DIR}' to your PATH to run 'vulnex' directly.\n"
fi
