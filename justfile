# preflag — command runner recipes

version := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
commit := `git rev-parse --short HEAD 2>/dev/null || echo "none"`
date := `date -u +"%Y-%m-%dT%H:%M:%SZ"`

ldflags := "-s -w -X main.version=" + version + " -X main.commit=" + commit + " -X main.date=" + date

_default:
    @just --list

# Build the binary.
build:
    go build -ldflags '{{ ldflags }}' -o preflag

# Run tests.
test:
    go test -v ./...

# Clean build artifacts.
clean:
    rm -f preflag
    rm -rf dist/

# Install the binary.
install:
    go install -ldflags '{{ ldflags }}'

# Build all cross-compilation targets.
build-all: build-linux build-darwin build-windows

# Build Linux targets (glibc + musl).
build-linux:
    GOOS=linux GOARCH=amd64 go build -ldflags '{{ ldflags }}' -o dist/preflag-x86_64-unknown-linux-gnu
    GOOS=linux GOARCH=arm64 go build -ldflags '{{ ldflags }}' -o dist/preflag-aarch64-unknown-linux-gnu
    GOOS=linux GOARCH=arm GOARM=6 go build -ldflags '{{ ldflags }}' -o dist/preflag-arm-unknown-linux-gnueabihf
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '{{ ldflags }}' -o dist/preflag-x86_64-unknown-linux-musl
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags '{{ ldflags }}' -o dist/preflag-aarch64-unknown-linux-musl

# Build macOS targets.
build-darwin:
    GOOS=darwin GOARCH=amd64 go build -ldflags '{{ ldflags }}' -o dist/preflag-x86_64-apple-darwin
    GOOS=darwin GOARCH=arm64 go build -ldflags '{{ ldflags }}' -o dist/preflag-aarch64-apple-darwin

# Build Windows targets.
build-windows:
    GOOS=windows GOARCH=amd64 go build -ldflags '{{ ldflags }}' -o dist/preflag-x86_64-pc-windows-gnu.exe

# Bump version, tag, and push (triggers GitHub Actions release). Usage: just release major|minor|patch
release bump:
    #!/usr/bin/env bash
    set -euo pipefail
    next=$(svu {{ bump }})
    echo "Releasing ${next}"
    git tag -a "${next}" -m "Release ${next}"
    git push origin "${next}"
