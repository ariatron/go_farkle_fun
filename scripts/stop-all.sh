#!/bin/bash

# Stop All Services for Farkle Observability

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

echo "🛑 Stopping Farkle Observability Stack..."
echo ""

# Stop continuous traffic
if [ -f "$PROJECT_ROOT/.continuous-traffic.pid" ]; then
    PID=$(cat "$PROJECT_ROOT/.continuous-traffic.pid")
    if ps -p "$PID" > /dev/null 2>&1; then
        kill "$PID" 2>/dev/null || true
        print_status "Stopped continuous traffic (PID: $PID)"
    fi
    rm "$PROJECT_ROOT/.continuous-traffic.pid"
fi

# Stop Farkle server
if [ -f "$PROJECT_ROOT/.server.pid" ]; then
    PID=$(cat "$PROJECT_ROOT/.server.pid")
    if ps -p "$PID" > /dev/null 2>&1; then
        kill "$PID" 2>/dev/null || true
        print_status "Stopped Farkle Server (PID: $PID)"
    fi
    rm "$PROJECT_ROOT/.server.pid"
fi

# Stop Alloy
if [ -f "$PROJECT_ROOT/.alloy.pid" ]; then
    PID=$(cat "$PROJECT_ROOT/.alloy.pid")
    if ps -p "$PID" > /dev/null 2>&1; then
        kill "$PID" 2>/dev/null || true
        print_status "Stopped Grafana Alloy (PID: $PID)"
    fi
    rm "$PROJECT_ROOT/.alloy.pid"
fi

echo ""
print_status "All services stopped"
