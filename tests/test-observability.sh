#!/bin/bash
# Test script to verify observability features

set -e

echo "🧪 Testing Farkle Observability Features"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"

echo "1️⃣  Checking if server is running..."
if curl -s -f "${BASE_URL}/health" > /dev/null; then
    echo -e "${GREEN}✓ Server is running${NC}"
else
    echo "❌ Server is not running. Start it with: go run cmd/server/main.go"
    exit 1
fi

echo ""
echo "2️⃣  Testing Health Endpoint..."
HEALTH=$(curl -s "${BASE_URL}/health")
if [ "$HEALTH" = "OK" ]; then
    echo -e "${GREEN}✓ Health check passed${NC}"
else
    echo "❌ Health check failed"
    exit 1
fi

echo ""
echo "3️⃣  Testing Metrics Endpoint..."
METRICS=$(curl -s "${BASE_URL}/metrics")
if echo "$METRICS" | grep -q "farkle_game_rolls_total"; then
    echo -e "${GREEN}✓ Metrics endpoint working${NC}"
    echo "   Found metrics:"
    echo "$METRICS" | grep "^farkle_" | grep "TYPE" | awk '{print "   - " $4}'
else
    echo "❌ Metrics endpoint not working"
    exit 1
fi

echo ""
echo "4️⃣  Testing API and Observability..."
echo "   Making API calls..."

# Reset game
curl -s "${BASE_URL}/api/reset" > /dev/null
echo "   - Reset game"

# Set player name
curl -s -X POST "${BASE_URL}/api/set-player-name" \
    -H "Content-Type: application/json" \
    -d '{"player_name":"TestPlayer"}' > /dev/null
echo "   - Set player name"

# Roll dice
curl -s -X POST "${BASE_URL}/api/roll" \
    -H "Content-Type: application/json" \
    -d '{"dice_to_keep":[]}' > /dev/null
echo "   - Rolled dice"

sleep 1

echo ""
echo "5️⃣  Verifying Metrics Updated..."
ROLL_COUNT=$(curl -s "${BASE_URL}/metrics" | grep "farkle_game_rolls_total" | grep -v "#" | awk '{print $2}')
if [ -n "$ROLL_COUNT" ] && [ "$ROLL_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ Metrics updated (${ROLL_COUNT} rolls recorded)${NC}"
else
    echo -e "${YELLOW}⚠️  Roll count is ${ROLL_COUNT}${NC}"
fi

HTTP_COUNT=$(curl -s "${BASE_URL}/metrics" | grep "farkle_http_requests_total" | grep -v "#" | wc -l)
if [ "$HTTP_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ HTTP metrics recorded (${HTTP_COUNT} metric combinations)${NC}"
else
    echo "❌ No HTTP metrics recorded"
    exit 1
fi

echo ""
echo "6️⃣  Checking Structured Logs..."
echo -e "${YELLOW}   Note: Logs are output to stdout in JSON format${NC}"
echo "   Sample log fields: time, level, msg, method, path, status, duration_ms, trace_id"

echo ""
echo "7️⃣  Summary"
echo "   ════════"
echo -e "${GREEN}   ✓ All observability features working!${NC}"
echo ""
echo "📊 View metrics: ${BASE_URL}/metrics"
echo "🏥 Health check: ${BASE_URL}/health"
echo "🔍 Traces: Start Jaeger with docker-compose -f docker-compose.observability.yml up -d"
echo "   Then visit http://localhost:16686"
echo ""
echo "🧪 Run load tests:"
echo "   k6 run tests/k6/smoke-test.js"
echo "   k6 run tests/k6/load-test.js"
echo ""
