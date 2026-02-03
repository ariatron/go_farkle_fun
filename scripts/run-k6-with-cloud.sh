#!/bin/bash
# Run K6 tests with Grafana Cloud Prometheus remote write output
#
# This script runs K6 tests and sends metrics directly to Grafana Cloud
# using the experimental Prometheus remote write output.
#
# Usage: ./scripts/run-k6-with-cloud.sh [test-file]
# Example: ./scripts/run-k6-with-cloud.sh tests/k6/load-test.js

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if config file exists
if [ ! -f "k6-grafana-cloud.json" ]; then
    echo -e "${YELLOW}Warning: k6-grafana-cloud.json not found${NC}"
    echo "Please create it from the example:"
    echo "  cp k6-grafana-cloud.json.example k6-grafana-cloud.json"
    echo "Then edit it with your Grafana Cloud credentials"
    exit 1
fi

# Load configuration
export $(cat k6-grafana-cloud.json | jq -r 'to_entries | .[] | "\(.key)=\(.value)"')

# Default test file
TEST_FILE="${1:-tests/k6/load-test.js}"

if [ ! -f "$TEST_FILE" ]; then
    echo -e "${YELLOW}Error: Test file not found: $TEST_FILE${NC}"
    echo "Available tests:"
    ls -1 tests/k6/*.js | grep -v utils.js
    exit 1
fi

echo -e "${BLUE}==================================${NC}"
echo -e "${BLUE}K6 Test with Grafana Cloud${NC}"
echo -e "${BLUE}==================================${NC}"
echo -e "${GREEN}Test file:${NC} $TEST_FILE"
echo -e "${GREEN}Prometheus endpoint:${NC} $K6_PROMETHEUS_RW_SERVER_URL"
echo -e "${GREEN}Username:${NC} $K6_PROMETHEUS_RW_USERNAME"
echo ""

# Run K6 with experimental Prometheus remote write output
echo -e "${BLUE}Starting K6 test...${NC}"
k6 run \
  --out experimental-prometheus-rw \
  "$TEST_FILE"

echo ""
echo -e "${GREEN}✓ Test complete!${NC}"
echo ""
echo -e "${BLUE}View results in Grafana Cloud:${NC}"
echo "  1. Go to your Grafana Cloud instance"
echo "  2. Navigate to Explore"
echo "  3. Select Prometheus data source"
echo "  4. Query k6 metrics, e.g.:"
echo "     - k6_http_req_duration"
echo "     - k6_http_reqs_total"
echo "     - k6_checks"
echo "     - k6_vus"
echo ""
