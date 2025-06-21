#!/bin/bash

# Quick build test script for local development

set -e

echo "🔨 Testing local build..."

# Build for current platform
go build -ldflags "-s -w -X main.Version=dev-test -X main.GitCommit=$(git rev-parse --short HEAD 2>/dev/null || echo unknown) -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o flint-vault-test ./cmd

echo "✅ Build successful!"

# Test version
echo ""
echo "📋 Version info:"
./flint-vault-test version

# Test help
echo ""
echo "📖 Help output:"
./flint-vault-test --help | head -10

# Clean up
rm -f flint-vault-test

echo ""
echo "🎉 Local build test completed successfully!" 