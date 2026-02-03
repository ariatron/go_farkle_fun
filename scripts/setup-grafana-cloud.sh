#!/bin/bash
# Setup script for Grafana Cloud integration

set -e

YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "╔════════════════════════════════════════╗"
echo "║  Farkle Game - Grafana Cloud Setup    ║"
echo "╔════════════════════════════════════════╗"
echo -e "${NC}"
echo ""

# Check if .env exists
if [ -f .env ]; then
    echo -e "${YELLOW}⚠️  .env file already exists${NC}"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Setup cancelled."
        exit 0
    fi
fi

echo -e "${GREEN}Let's configure your Grafana Cloud credentials${NC}"
echo ""
echo "You'll need the following from your Grafana Cloud portal:"
echo "1. Stack name (e.g., mystack.grafana.net)"
echo "2. Instance ID (numerical, e.g., 123456)"
echo "3. API Key (starts with glc_)"
echo "4. Region (e.g., us-central1, prod-us-east-0)"
echo ""
echo "Get these from: https://grafana.com/auth/sign-in/"
echo "Then go to: My Account → Stack → Details"
echo ""

# Prompt for credentials
read -p "Stack name (without .grafana.net): " STACK_NAME
read -p "Region (e.g., us-central1): " REGION
read -p "Instance ID: " INSTANCE_ID
read -p "API Key: " API_KEY

# Validate inputs
if [ -z "$STACK_NAME" ] || [ -z "$REGION" ] || [ -z "$INSTANCE_ID" ] || [ -z "$API_KEY" ]; then
    echo -e "${YELLOW}Error: All fields are required${NC}"
    exit 1
fi

# Create .env file
cat > .env <<EOF
# Grafana Cloud Configuration
GRAFANA_CLOUD_ENABLED=true

# Prometheus (Metrics)
PROMETHEUS_REMOTE_WRITE_URL=https://prometheus-${REGION}.grafana.net/api/prom/push
PROMETHEUS_REMOTE_WRITE_USERNAME=${INSTANCE_ID}
PROMETHEUS_REMOTE_WRITE_PASSWORD=${API_KEY}

# Tempo (Traces)
TEMPO_ENDPOINT=tempo-${REGION}.grafana.net:443
TEMPO_USERNAME=${INSTANCE_ID}
TEMPO_PASSWORD=${API_KEY}

# Loki (Logs) - Optional
LOKI_ENABLED=false
LOKI_URL=https://logs-${REGION}.grafana.net/loki/api/v1/push
LOKI_USERNAME=${INSTANCE_ID}
LOKI_PASSWORD=${API_KEY}

# Application Settings
SERVICE_NAME=farkle-game
ENVIRONMENT=production
PORT=8080
LOG_LEVEL=info
EOF

echo ""
echo -e "${GREEN}✓ .env file created${NC}"

# Create Grafana Agent config
echo ""
echo "Creating Grafana Agent configuration..."

# Create grafana-agent-config.yaml with actual credentials
cat > grafana-agent-config.yaml <<EOF
server:
  log_level: info
  http_listen_port: 12345

metrics:
  global:
    scrape_interval: 15s
    remote_write:
      - url: https://prometheus-${REGION}.grafana.net/api/prom/push
        basic_auth:
          username: ${INSTANCE_ID}
          password: ${API_KEY}

  configs:
    - name: farkle-game
      scrape_configs:
        - job_name: 'farkle-app'
          static_configs:
            - targets: ['localhost:8080']
              labels:
                service: 'farkle-game'
                environment: 'production'
          metrics_path: '/metrics'
          scrape_interval: 15s

traces:
  configs:
    - name: farkle-traces
      receivers:
        otlp:
          protocols:
            http:
              endpoint: "0.0.0.0:4318"
            grpc:
              endpoint: "0.0.0.0:4317"

      remote_write:
        - endpoint: tempo-${REGION}.grafana.net:443
          basic_auth:
            username: ${INSTANCE_ID}
            password: ${API_KEY}

      batch:
        timeout: 5s
        send_batch_size: 100
EOF

echo -e "${GREEN}✓ Grafana Agent config created${NC}"

echo ""
echo -e "${GREEN}Setup complete!${NC}"
echo ""
echo "Next steps:"
echo ""
echo "1. Install Grafana Agent (if not already installed):"
echo "   macOS: brew install grafana-agent"
echo "   Linux: See https://grafana.com/docs/agent/latest/set-up/install-agent-linux/"
echo ""
echo "2. Start Grafana Agent:"
echo "   grafana-agent -config.file=grafana-agent-config.yaml"
echo ""
echo "3. Start your Farkle server:"
echo "   export JAEGER_ENDPOINT=\"tempo-${REGION}.grafana.net:443\""
echo "   go run cmd/server/main.go"
echo ""
echo "4. View your data in Grafana Cloud:"
echo "   https://${STACK_NAME}.grafana.net"
echo ""
echo "5. Import the dashboard:"
echo "   Dashboard → Import → Upload grafana-dashboards/farkle-dashboard.json"
echo ""
