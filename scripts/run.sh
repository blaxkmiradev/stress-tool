#!/bin/bash
# scripts/run.sh

TARGET=${1:-"http://localhost:8080"}
THREADS=${2:-100}
DURATION=${3:-30}
ATTACK=${4:-"flood"}

echo "[+] Running Layer7 stress test"
echo "Target: $TARGET"
echo "Threads: $THREADS"
echo "Duration: $DURATION"
echo "Attack type: $ATTACK"

go run cmd/ddos/main.go \
    -target "$TARGET" \
    -threads "$THREADS" \
    -duration "$DURATION" \
    -attack "$ATTACK"
