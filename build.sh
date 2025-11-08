#!/bin/bash

# RWMod Monitor Build Script for Linux/macOS

set -e

APP_NAME="rwmod-monitor"
VERSION=${VERSION:-"1.0.0"}

echo "Building RWMod Monitor..."
echo "=========================="

# Build for current platform
echo "Building for current platform..."
go build -ldflags "-s -w" -o ${APP_NAME}
echo "✓ Built: ${APP_NAME}"

# Build for Linux (amd64)
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ${APP_NAME}-linux-amd64
echo "✓ Built: ${APP_NAME}-linux-amd64"

# Build for Linux (arm64)
echo "Building for Linux (arm64)..."
GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o ${APP_NAME}-linux-arm64
echo "✓ Built: ${APP_NAME}-linux-arm64"

# Build for macOS (amd64)
echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o ${APP_NAME}-darwin-amd64
echo "✓ Built: ${APP_NAME}-darwin-amd64"

# Build for macOS (arm64)
echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o ${APP_NAME}-darwin-arm64
echo "✓ Built: ${APP_NAME}-darwin-arm64"

# Build for Windows (amd64)
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o ${APP_NAME}-windows-amd64.exe
echo "✓ Built: ${APP_NAME}-windows-amd64.exe"

# Build for Windows (arm64)
echo "Building for Windows (arm64)..."
GOOS=windows GOARCH=arm64 go build -ldflags "-s -w" -o ${APP_NAME}-windows-arm64.exe
echo "✓ Built: ${APP_NAME}-windows-arm64.exe"

echo ""
echo "=========================="
echo "Build completed successfully!"
echo "=========================="
echo ""
echo "Output files:"
ls -lh ${APP_NAME}* 2>/dev/null | grep -v ".go" || true
