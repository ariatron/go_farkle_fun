#!/bin/bash
# Helper script to start Grafana Alloy in the background

cd "$(dirname "$0")/.."

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting Grafana Alloy...${NC}"

# Check if Alloy is already running
if pgrep -f "alloy run" > /dev/null; then
    echo -e "${YELLOW}⚠️  Alloy is already running${NC}"
    echo "Stop it with: pkill -f 'alloy run'"
    exit 1
fi

# Start Alloy in background
nohup alloy run alloy-config.alloy > /tmp/alloy.log 2>&1 &
ALLOY_PID=$!

# Wait a moment for it to start
sleep 2

# Check if it's running
if ps -p $ALLOY_PID > /dev/null; then
    echo -e "${GREEN}✓ Alloy started successfully!${NC}"
    echo ""
    echo "PID: $ALLOY_PID"
    echo "Logs: /tmp/alloy.log"
    echo ""
    echo "Endpoints:"
    echo "  - Alloy UI: http://localhost:12345"
    echo "  - OTLP HTTP: http://localhost:4318"
    echo "  - OTLP gRPC: http://localhost:4317"
    echo ""
    echo "View logs: tail -f /tmp/alloy.log"
    echo "Stop: pkill -f 'alloy run'"
else
    echo -e "${RED}❌ Failed to start Alloy${NC}"
    echo "Check logs: cat /tmp/alloy.log"
    exit 1
fi
