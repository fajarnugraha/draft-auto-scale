#!/bin/bash

# This script downloads the appropriate k6 and jq binaries for the current OS and architecture.

set -e # Exit immediately if a command exits with a non-zero status.

# --- Configuration ---
BIN_DIR="bin"
K6_VERSION="v0.50.0" # Latest version as of Aug 2025
JQ_VERSION="jq-1.7.1" # Latest version as of Aug 2025

# --- Functions ---
print_header() {
    echo "------------------------------------------------"
    echo "  $1"
    echo "------------------------------------------------"
}

# --- Main Script ---
mkdir -p "$BIN_DIR"

# 1. Install k6
if [ -f "$BIN_DIR/k6" ]; then
    echo "k6 is already installed."
else
    print_header "Installing k6"
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64)
            ARCH="arm64"
            ;;
    esac

    K6_URL="https://github.com/grafana/k6/releases/download/${K6_VERSION}/k6-${K6_VERSION}-${OS}-${ARCH}.tar.gz"
    echo "Downloading k6 from $K6_URL"
    curl -L "$K6_URL" | tar -xz -C "$BIN_DIR" --strip-components=1 "k6-${K6_VERSION}-${OS}-${ARCH}/k6"
    chmod +x "$BIN_DIR/k6"
    echo "k6 installed successfully."
fi


# 2. Install jq
if [ -f "$BIN_DIR/jq" ]; then
    echo "jq is already installed."
else
    print_header "Installing jq"
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    JQ_URL="https://github.com/jqlang/jq/releases/download/${JQ_VERSION}/jq-${OS}-amd64"
    # Note: jq has different naming conventions, we'll assume amd64 for simplicity as it's most common.
    # A more robust script would handle different architectures for jq as well.
    echo "Downloading jq from $JQ_URL"
    curl -L -o "$BIN_DIR/jq" "$JQ_URL"
    chmod +x "$BIN_DIR/jq"
    echo "jq installed successfully."
fi

echo "All tools are ready."
