#!/bin/bash

# Start All Services for Farkle Observability
# Starts: Alloy, Farkle Server, and Continuous Traffic

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if alloy is installed
if ! command -v alloy &> /dev/null; then
    print_error "Grafana Alloy is not installed"
    echo "Install: brew install grafana/grafana/alloy"
    exit 1
fi

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    print_warning "K6 is not installed (traffic generation will be skipped)"
    echo "Install: brew install k6"
fi

# Check if alloy-config.alloy exists
if [ ! -f "$PROJECT_ROOT/alloy-config.alloy" ]; then
    print_error "alloy-config.alloy not found"
    echo "Copy alloy-config.alloy.example to alloy-config.alloy and add your credentials"
    exit 1
fi

echo "================================================"
echo "🎲 Starting Farkle Observability Stack"
echo "================================================"
echo ""

# 1. Start Alloy
print_info "Starting Grafana Alloy..."
cd "$PROJECT_ROOT"
nohup alloy run alloy-config.alloy > logs/alloy.log 2>&1 &
ALLOY_PID=$!
echo "$ALLOY_PID" > .alloy.pid

sleep 3

if ps -p "$ALLOY_PID" > /dev/null 2>&1; then
    print_status "Grafana Alloy started (PID: $ALLOY_PID)"
    print_info "Alloy UI: http://localhost:12345"
else
    print_error "Failed to start Alloy"
    exit 1
fi

echo ""

# 2. Start Farkle Server
print_info "Starting Farkle Server..."
export JAEGER_ENDPOINT="localhost:4318"

nohup go run cmd/server/main.go > logs/server.log 2>&1 &
SERVER_PID=$!
echo "$SERVER_PID" > .server.pid

sleep 3

if ps -p "$SERVER_PID" > /dev/null 2>&1; then
    print_status "Farkle Server started (PID: $SERVER_PID)"
    print_info "Game UI: http://localhost:8080"
    print_info "Metrics: http://localhost:8080/metrics"
    print_info "Health: http://localhost:8080/health"
else
    print_error "Failed to start Farkle Server"
    kill "$ALLOY_PID" 2>/dev/null || true
    exit 1
fi

echo ""

# 3. Start Continuous Traffic (optional)
if command -v k6 &> /dev/null; then
    read -p "Start continuous traffic generation? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Starting continuous traffic..."
        ./scripts/continuous-traffic.sh start
        echo ""
    fi
else
    print_warning "K6 not installed, skipping traffic generation"
    echo ""
fi

# Summary
echo "================================================"
echo "✅ All services started successfully!"
echo "================================================"
echo ""
echo "📊 Access Points:"
echo "  • Game UI:        http://localhost:8080"
echo "  • Prometheus:     http://localhost:8080/metrics"
echo "  • Alloy UI:       http://localhost:12345"
echo "  • Grafana Cloud:  https://grafana.com"
echo ""
echo "📝 Logs:"
echo "  • Alloy:          tail -f logs/alloy.log"
echo "  • Server:         tail -f logs/server.log"
echo "  • Traffic:        tail -f continuous-traffic.log"
echo ""
echo "🛑 Stop all services:"
echo "  ./scripts/stop-all.sh"
echo ""
