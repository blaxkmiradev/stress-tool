#!/bin/bash
# scripts/build.sh

set -e

echo "[+] Building Layer7 Flood Tool"
echo "[+] Downloading dependencies..."
go mod download
go mod verify

echo "[+] Building main binary..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/layer7-flood-linux-amd64 cmd/ddos/main.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/layer7-flood-darwin-amd64 cmd/ddos/main.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/layer7-flood-windows-amd64.exe cmd/ddos/main.go

echo "[+] Build complete. Binaries in ./bin/"
ls -lh bin/
