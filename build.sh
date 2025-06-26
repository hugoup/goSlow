#!/bin/sh
# Build goSlow for Linux, macOS, and Windows (amd64)
# Usage: ./build.sh
set -e

APP=goSlow
OUTDIR=build
mkdir -p $OUTDIR

echo "Building for Linux (amd64)..."
GOOS=linux   GOARCH=amd64 go build -o $OUTDIR/${APP}-linux-amd64 main.go

echo "Building for macOS (amd64)..."
GOOS=darwin  GOARCH=amd64 go build -o $OUTDIR/${APP}-darwin-amd64 main.go

echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o $OUTDIR/${APP}-windows-amd64.exe main.go

echo "All builds complete. Binaries are in $OUTDIR/"
