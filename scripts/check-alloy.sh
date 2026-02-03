#!/bin/bash
# Check Alloy status and recent logs

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}═══════════════════════════════════${NC}"
echo -e "${BLUE}   Grafana Alloy Status Check${NC}"
echo -e "${BLUE}═══════════════════════════════════${NC}"
echo ""

# Check if Alloy is running
if pgrep -f "alloy run" > /dev/null; then
    PID=$(pgrep -f "alloy run")
    echo -e "${GREEN}✓ Alloy is running${NC}"
    echo "  PID: $PID"
    echo ""
else
    echo -e "${RED}✗ Alloy is NOT running${NC}"
    echo ""
    echo "Start with: ./scripts/start-alloy.sh"
    echo "Or manually: alloy run alloy-config.alloy"
    exit 1
fi

# Check endpoints
echo -e "${YELLOW}Checking endpoints...${NC}"

# Check UI
if curl -s http://localhost:12345 > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Alloy UI (12345)${NC} - http://localhost:12345"
else
    echo -e "${RED}✗ Alloy UI (12345)${NC} - Not accessible"
fi

# Check OTLP HTTP
if curl -s http://localhost:4318 > /dev/null 2>&1; then
    echo -e "${GREEN}✓ OTLP HTTP (4318)${NC} - Listening"
else
    echo -e "${RED}✗ OTLP HTTP (4318)${NC} - Not listening"
fi

# Check OTLP gRPC (can't easily test, just show expected)
echo -e "${YELLOW}• OTLP gRPC (4317)${NC} - Expected (can't verify without client)"

echo ""

# Show recent logs
echo -e "${YELLOW}Recent logs (last 15 lines):${NC}"
echo "────────────────────────────────────"
if [ -f /tmp/alloy.log ]; then
    tail -15 /tmp/alloy.log
    echo ""
    echo "View all logs: tail -f /tmp/alloy.log"
else
    echo "No log file found at /tmp/alloy.log"
fi

echo ""
echo -e "${BLUE}═══════════════════════════════════${NC}"
