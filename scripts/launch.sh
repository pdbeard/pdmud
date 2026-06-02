#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$SCRIPT_DIR/.."

cd "$ROOT"

# Create log files so widgets don't error on first start
mkdir -p logs
touch logs/global.log logs/local.log logs/tells.log logs/auction.log
touch logs/affects.state logs/party.state logs/party.state.tmp
touch logs/gmcp.log logs/msdp.log

# Build widgets if binaries are missing
if [[ ! -f bin/chat || ! -f bin/affects ]]; then
    echo "Binaries not found, building..."
    make build
fi

# Launch zellij with our layout
exec zellij --layout layouts/mud.kdl
