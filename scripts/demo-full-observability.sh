#!/bin/bash
# Complete Observability Demo Script
# Demonstrates metrics, traces, and logs flowing to Grafana Cloud

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m' # No Color

echo -e "${BOLD}${BLUE}"
echo "╔═══════════════════════════════════════════╗"
echo "║  🎲 Farkle Game Observability Demo  🎲   ║"
echo "║  Metrics + Traces + Logs + K6 Tests      ║"
echo "╚═══════════════════════════════════════════╝"
echo -e "${NC}"
echo ""

# Check if everything is installed
echo -e "${CYAN}Checking prerequisites...${NC}"

if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go is not installed${NC}"
    exit 1
fi

if ! command -v alloy &> /dev/null; then
    echo -e "${RED}✗ Grafana Alloy is not installed${NC}"
    echo "  Install with: brew install grafana/grafana/alloy"
    exit 1
fi

if ! command -v k6 &> /dev/null; then
    echo -e "${RED}✗ K6 is not installed${NC}"
    echo "  Install with: brew install k6"
    exit 1
fi

echo -e "${GREEN}✓ All prerequisites installed${NC}"
echo ""

# Stop any running instances
echo -e "${YELLOW}Cleaning up any running instances...${NC}"
pkill -f "alloy run" 2>/dev/null || true
pkill -f "go run cmd/server/main.go" 2>/dev/null || true
sleep 2
echo -e "${GREEN}✓ Clean${NC}"
echo ""

# Start Alloy
echo -e "${BOLD}${CYAN}Step 1: Starting Grafana Alloy${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cd "$(dirname "$0")/.."
nohup alloy run alloy-config.alloy > /tmp/alloy.log 2>&1 &
ALLOY_PID=$!
sleep 3

if ps -p $ALLOY_PID > /dev/null; then
    echo -e "${GREEN}✓ Alloy started (PID: $ALLOY_PID)${NC}"
    echo "  Logs: /tmp/alloy.log"
    echo "  UI: http://localhost:12345"
else
    echo -e "${RED}✗ Failed to start Alloy${NC}"
    cat /tmp/alloy.log
    exit 1
fi
echo ""

# Start Farkle Server
echo -e "${BOLD}${CYAN}Step 2: Starting Farkle Game Server${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
export JAEGER_ENDPOINT="localhost:4318"
nohup go run cmd/server/main.go > /tmp/farkle-server.log 2>&1 &
SERVER_PID=$!
sleep 4

if ps -p $SERVER_PID > /dev/null; then
    echo -e "${GREEN}✓ Farkle server started (PID: $SERVER_PID)${NC}"
    echo "  Game: http://localhost:8080"
    echo "  Metrics: http://localhost:8080/metrics"
    echo "  Logs: /tmp/farkle-app.log"
else
    echo -e "${RED}✗ Failed to start Farkle server${NC}"
    cat /tmp/farkle-server.log
    exit 1
fi
echo ""

# Verify endpoints
echo -e "${BOLD}${CYAN}Step 3: Verifying Observability Stack${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check Alloy UI
if curl -s http://localhost:12345 > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Alloy UI${NC} - http://localhost:12345"
else
    echo -e "${RED}✗ Alloy UI${NC}"
fi

# Check OTLP receiver
if curl -s http://localhost:4318 > /dev/null 2>&1; then
    echo -e "${GREEN}✓ OTLP HTTP (traces)${NC} - localhost:4318"
else
    echo -e "${RED}✗ OTLP HTTP${NC}"
fi

# Check Farkle metrics
if curl -s http://localhost:8080/metrics | grep -q "farkle_"; then
    echo -e "${GREEN}✓ Farkle Metrics${NC} - http://localhost:8080/metrics"
else
    echo -e "${RED}✗ Farkle Metrics${NC}"
fi

# Check Farkle health
if curl -s http://localhost:8080/health | grep -q "OK"; then
    echo -e "${GREEN}✓ Farkle Health${NC} - http://localhost:8080/health"
else
    echo -e "${RED}✗ Farkle Health${NC}"
fi

echo ""
sleep 2

# Run K6 demo test
echo -e "${BOLD}${CYAN}Step 4: Running K6 Load Test (Dashboard Demo)${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${YELLOW}This will generate varied traffic patterns for the dashboard...${NC}"
echo -e "${YELLOW}Duration: ~10 minutes${NC}"
echo ""

sleep 3

k6 run tests/k6/dashboard-demo.js

echo ""
echo -e "${GREEN}✓ K6 test completed${NC}"
echo ""

# Show results
echo -e "${BOLD}${CYAN}Step 5: Demo Complete!${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${BOLD}${GREEN}Your observability stack is running! 🎉${NC}"
echo ""
echo -e "${BOLD}Local Endpoints:${NC}"
echo "  • Game:         http://localhost:8080"
echo "  • Alloy UI:     http://localhost:12345"
echo "  • Metrics:      http://localhost:8080/metrics"
echo "  • Health:       http://localhost:8080/health"
echo ""
echo -e "${BOLD}Grafana Cloud:${NC}"
echo "  • Dashboard:    Import grafana-dashboards/farkle-complete-dashboard.json"
echo "  • Metrics:      Explore → Prometheus"
echo "  • Traces:       Explore → Tempo → Service: farkle-game"
echo "  • Logs:         Explore → Loki → Service: farkle-game"
echo ""
echo -e "${BOLD}View Logs:${NC}"
echo "  • Alloy:        tail -f /tmp/alloy.log"
echo "  • Server:       tail -f /tmp/farkle-server.log"
echo "  • App Logs:     tail -f /tmp/farkle-app.log"
echo ""
echo -e "${BOLD}Example Queries:${NC}"
echo -e "${CYAN}Metrics (Prometheus):${NC}"
echo "  rate(farkle_http_requests_total[5m])"
echo "  histogram_quantile(0.95, rate(farkle_http_request_duration_seconds_bucket[5m]))"
echo "  farkle_game_rolls_total"
echo ""
echo -e "${CYAN}Traces (Tempo):${NC}"
echo "  {resource.service.name=\"farkle-game\"}"
echo ""
echo -e "${CYAN}Logs (Loki):${NC}"
echo "  {service=\"farkle-game\"} | json"
echo "  {service=\"farkle-game\",level=\"ERROR\"}"
echo "  {service=\"farkle-game\"} |~ \"rolled|banked|farkled\""
echo ""
echo -e "${BOLD}Stop Everything:${NC}"
echo "  pkill -f \"alloy run\""
echo "  pkill -f \"go run cmd/server/main.go\""
echo ""
echo -e "${BOLD}Run More Tests:${NC}"
echo "  k6 run tests/k6/load-test.js"
echo "  k6 run tests/k6/stress-test.js"
echo "  k6 run tests/k6/spike-test.js"
echo ""
echo -e "${GREEN}${BOLD}Happy demoing! 🎲📊🔍${NC}"
echo ""
