# Grafana Cloud Integration Guide

This guide explains how to configure your Farkle game to send observability data to Grafana Cloud.

## Overview

Grafana Cloud provides managed observability services:
- **Metrics** - Prometheus-compatible metrics storage and querying
- **Logs** - Loki for log aggregation
- **Traces** - Tempo for distributed tracing
- **Dashboards** - Grafana for visualization

## Prerequisites

1. Grafana Cloud account (free tier available at https://grafana.com/auth/sign-up/create-user)
2. Your Grafana Cloud stack details:
   - Stack name/URL
   - Prometheus endpoint
   - Tempo endpoint
   - Loki endpoint
   - API key/token

## Configuration

### 1. Get Your Grafana Cloud Credentials

From your Grafana Cloud portal:

1. Go to **My Account** → **Stack**
2. Note your stack name (e.g., `mystack.grafana.net`)
3. Click **Details** on your Prometheus instance
   - Note the **Remote Write Endpoint**
   - Note the **Username** (usually your stack ID)
4. Click **Details** on your Tempo instance
   - Note the **OTLP/HTTP endpoint**
5. Generate an **API Key**:
   - Go to **Security** → **API Keys**
   - Create a new key with **MetricsPublisher** role
   - Save this key securely

### 2. Environment Variables

Create a `.env` file in your project root:

```bash
# Grafana Cloud Configuration
GRAFANA_CLOUD_ENABLED=true

# Prometheus Remote Write
PROMETHEUS_REMOTE_WRITE_URL=https://prometheus-us-central1.grafana.net/api/prom/push
PROMETHEUS_REMOTE_WRITE_USERNAME=123456  # Your Grafana Cloud instance ID
PROMETHEUS_REMOTE_WRITE_PASSWORD=glc_xxxxxxxxxxxxx  # Your API key

# Tempo (Traces)
TEMPO_ENDPOINT=tempo-us-central1.grafana.net:443
TEMPO_HEADERS=Authorization=Basic <base64-encoded-username:apikey>

# Loki (Logs) - Optional
LOKI_ENABLED=false
LOKI_URL=https://logs-us-central1.grafana.net/loki/api/v1/push
LOKI_USERNAME=123456
LOKI_PASSWORD=glc_xxxxxxxxxxxxx

# Application Settings
SERVICE_NAME=farkle-game
ENVIRONMENT=production
```

**Note**: Replace the URLs and credentials with your actual Grafana Cloud stack details.

### 3. Install Prometheus Pushgateway Client (Optional)

For better Grafana Cloud integration, you can use remote write:

```bash
cd /Users/ariatron/farkle-fun
go get github.com/prometheus/client_golang/prometheus/push
```

Or use the Grafana Agent (recommended for production).

### 4. Configure Tempo Endpoint

Update your server to use Grafana Cloud Tempo instead of local Jaeger:

```bash
# For Grafana Cloud Tempo
export JAEGER_ENDPOINT="tempo-us-central1.grafana.net:443"
export OTEL_EXPORTER_OTLP_ENDPOINT="https://tempo-us-central1.grafana.net:443"
export OTEL_EXPORTER_OTLP_HEADERS="Authorization=Basic $(echo -n '123456:glc_xxxxx' | base64)"

go run cmd/server/main.go
```

### 5. Run with Grafana Cloud

```bash
# Set your Grafana Cloud credentials
export JAEGER_ENDPOINT="tempo-us-central1.grafana.net:443"

# Start the server
go run cmd/server/main.go
```

The application will:
- ✅ Expose metrics at `:8080/metrics` for scraping
- ✅ Send traces to Grafana Cloud Tempo
- ✅ Output structured JSON logs (can be sent to Loki)

## Viewing Your Data

### Metrics

1. Log in to your Grafana Cloud portal
2. Go to **Explore**
3. Select your Prometheus data source
4. Query your metrics:

```promql
# Request rate
rate(farkle_http_requests_total[5m])

# P95 latency
histogram_quantile(0.95, rate(farkle_http_request_duration_seconds_bucket[5m]))

# Game metrics
farkle_game_rolls_total
farkle_game_wins_total
```

### Traces

1. Go to **Explore**
2. Select your Tempo data source
3. Search for traces by:
   - Service name: `farkle-game`
   - Operation name: `POST /api/roll`, `GET /api/bank`, etc.
   - Time range

### Logs (if using Loki)

1. Go to **Explore**
2. Select your Loki data source
3. Query logs:

```logql
{service_name="farkle-game"} | json
{service_name="farkle-game"} |= "farkled"
{service_name="farkle-game"} | json | level="ERROR"
```

## Metrics Scraping

You have two options for sending metrics to Grafana Cloud:

### Option 1: Grafana Agent (Recommended)

Install and configure the Grafana Agent on your server:

```bash
# Download Grafana Agent
brew install grafana-agent  # macOS
# or use your package manager

# Create agent config
cat > agent-config.yaml <<EOF
server:
  log_level: info

metrics:
  global:
    scrape_interval: 15s
    remote_write:
      - url: https://prometheus-us-central1.grafana.net/api/prom/push
        basic_auth:
          username: 123456
          password: glc_xxxxxxxxxxxxx
  configs:
    - name: farkle
      scrape_configs:
        - job_name: 'farkle-game'
          static_configs:
            - targets: ['localhost:8080']
          metrics_path: '/metrics'

traces:
  configs:
    - name: farkle
      receivers:
        otlp:
          protocols:
            http:
              endpoint: "0.0.0.0:4318"
      remote_write:
        - endpoint: tempo-us-central1.grafana.net:443
          basic_auth:
            username: 123456
            password: glc_xxxxxxxxxxxxx
      batch:
        timeout: 5s
        send_batch_size: 100
EOF

# Run the agent
grafana-agent -config.file=agent-config.yaml
```

### Option 2: Direct Scraping

Configure Grafana Cloud to scrape your metrics endpoint directly:

1. In Grafana Cloud, go to **Integrations**
2. Add a **Prometheus scrape job**
3. Configure:
   - Job name: `farkle-game`
   - Target: Your server's public URL
   - Metrics path: `/metrics`
   - Scrape interval: `15s`

## Creating Dashboards

### Pre-built Dashboard Template

Import this dashboard JSON in Grafana Cloud:

1. Go to **Dashboards** → **New** → **Import**
2. Use the dashboard template from `grafana-dashboards/farkle-dashboard.json`

Or create custom panels with these queries:

**Request Rate Panel:**
```promql
sum(rate(farkle_http_requests_total[5m])) by (endpoint)
```

**Latency Panel:**
```promql
histogram_quantile(0.95,
  sum(rate(farkle_http_request_duration_seconds_bucket[5m])) by (le, endpoint)
)
```

**Game Stats Panel:**
```promql
# Rolls per second
rate(farkle_game_rolls_total[1m])

# Banks per second
rate(farkle_game_banks_total[1m])

# Farkle rate
rate(farkle_game_farkles_total[1m]) / rate(farkle_game_rolls_total[1m])

# Total wins
farkle_game_wins_total
```

**Active Games Panel:**
```promql
farkle_active_games
```

## Alerting

Create alerts in Grafana Cloud:

### High Error Rate Alert

```promql
sum(rate(farkle_http_requests_total{status=~"5.."}[5m]))
/
sum(rate(farkle_http_requests_total[5m]))
> 0.05
```

### High Latency Alert

```promql
histogram_quantile(0.95,
  rate(farkle_http_request_duration_seconds_bucket[5m])
) > 1
```

### Service Down Alert

```promql
up{job="farkle-game"} == 0
```

## Production Deployment Checklist

- [ ] Configure Grafana Cloud credentials
- [ ] Install and configure Grafana Agent (or alternative scraper)
- [ ] Verify metrics are being received in Grafana Cloud
- [ ] Verify traces are being received in Tempo
- [ ] Set up dashboards
- [ ] Configure alerts
- [ ] Set up notification channels (Slack, email, PagerDuty, etc.)
- [ ] Test alerting with K6 load tests
- [ ] Document runbooks for common issues

## Cost Optimization

Grafana Cloud free tier includes:
- 10,000 series for Prometheus metrics
- 50 GB for Loki logs
- 50 GB for Tempo traces

Tips to stay within limits:
1. **Reduce metric cardinality**: Avoid high-cardinality labels
2. **Adjust retention**: Use shorter retention for less critical data
3. **Sample traces**: Use probabilistic sampling (e.g., 10% of traces)
4. **Filter logs**: Only send important logs to Loki

## Troubleshooting

### Metrics not appearing

**Check connection:**
```bash
# Test remote write endpoint
curl -v -X POST \
  -u "123456:glc_xxxxx" \
  https://prometheus-us-central1.grafana.net/api/prom/push
```

**Check agent logs:**
```bash
# If using Grafana Agent
tail -f /var/log/grafana-agent.log
```

### Traces not appearing

**Verify endpoint:**
```bash
# Test Tempo endpoint
curl -v https://tempo-us-central1.grafana.net:443
```

**Check app logs:**
```bash
go run cmd/server/main.go | grep tracing
```

### High cardinality warnings

If you receive cardinality warnings from Grafana Cloud:

1. Review your metric labels
2. Remove or aggregate high-cardinality labels
3. Use recording rules to pre-aggregate data

## Resources

- [Grafana Cloud Documentation](https://grafana.com/docs/grafana-cloud/)
- [Prometheus Remote Write](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write)
- [Grafana Agent Documentation](https://grafana.com/docs/agent/)
- [Tempo Documentation](https://grafana.com/docs/tempo/)

## Support

For Grafana Cloud support:
- Documentation: https://grafana.com/docs/
- Community Forum: https://community.grafana.com/
- Support: Available for paid plans
