#!/bin/bash

APP_NAME="dicesong"
BUILD_DIR="build"
BIN_DIR="$HOME/.local/bin"
SRC_DIR="."


build() {
    echo "🔧 Building $APP_NAME..."
    mkdir -p "$BUILD_DIR"
    go build -o "$BUILD_DIR/$APP_NAME" "$SRC_DIR"
    echo "✅ Build complete."
}

install() {
    build
    echo "📦 Installing $APP_NAME to $BIN_DIR..."
    mkdir -p "$BIN_DIR"
    cp "$BUILD_DIR/$APP_NAME" "$BIN_DIR/"
    echo "✅ Installed $APP_NAME to $BIN_DIR"
}

uninstall() {
    echo "🗑️ Uninstalling $APP_NAME from $BIN_DIR..."
    rm -f "$BIN_DIR/$APP_NAME"
    echo "✅ Uninstalled $APP_NAME."
}

clean() {
    echo "🧹 Cleaning build files..."
    rm -rf "$BUILD_DIR"
    echo "✅ Clean complete."
}

usage() {
    echo "Usage: $0 {build|install|uninstall|clean}"
    echo "  build      - Compile the application."
    echo "  install    - Build and install the application to $BIN_DIR."
    echo "  uninstall  - Remove the application from $BIN_DIR."
    echo "  clean      - Remove build artifacts."
}

case "$1" in
    build)
        build
        ;;
    install)
        install
        ;;
    uninstall)
        uninstall
        ;;
    clean)
        clean
        ;;
    *)
        usage
        ;;
esac
