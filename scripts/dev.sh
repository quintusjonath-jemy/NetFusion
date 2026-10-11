#!/usr/bin/env bash
# ==============================================================================
# NetFusion — Concurrent Development Launcher
# Starts Go backend (embedded SQLite + simulation mode) and Vite frontend concurrently.
# ==============================================================================

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "=========================================================="
echo " Starting NetFusion Local Development Stack"
echo "=========================================================="

cleanup() {
    echo ""
    echo "Shutting down NetFusion processes..."
    kill 0
    exit
}

trap cleanup SIGINT SIGTERM EXIT

# 1. Start Go Backend
echo "[+] Starting NetFusion Backend on http://127.0.0.1:8080..."
(
    cd "$ROOT_DIR/backend"
    PORT=8080 HOST=127.0.0.1 DB_DRIVER=sqlite SIMULATION_MODE=true LOG_LEVEL=INFO go run ./cmd/server/main.go
) &

# Wait briefly for backend to initialize
sleep 2

# 2. Start Frontend Dev Server
echo "[+] Starting NetFusion React Dashboard on http://localhost:5173..."
(
    cd "$ROOT_DIR/frontend"
    npm run dev
) &

wait
