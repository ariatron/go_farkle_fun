# Grafana Cloud Quick Start with Alloy

Get your Farkle game sending data to Grafana Cloud in 5 minutes using **Grafana Alloy**!

> **Note**: Grafana Alloy is the successor to Grafana Agent (which is deprecated). We recommend using Alloy for all new deployments.

## Prerequisites

- Grafana Cloud account (free tier: https://grafana.com/auth/sign-up/create-user)
- Grafana Alloy installed locally

## Step 1: Install Grafana Alloy

### macOS
```bash
brew install grafana/grafana/alloy
```

### Linux (Ubuntu/Debian)
```bash
sudo mkdir -p /etc/apt/keyrings/
wget -q -O - https://apt.grafana.com/gpg.key | gpg --dearmor | sudo tee /etc/apt/keyrings/grafana.gpg > /dev/null
echo "deb [signed-by=/etc/apt/keyrings/grafana.gpg] https://apt.grafana.com stable main" | sudo tee /etc/apt/sources.list.d/grafana.list
sudo apt-get update
sudo apt-get install -y alloy
```

### Docker
```bash
docker pull grafana/alloy:latest
```

## Step 2: Configure Alloy (Already Done!)

Your Farkle game already has an Alloy configuration file: [`alloy-config.alloy`](/Users/ariatron/farkle-fun/alloy-config.alloy)

The configuration is already set up with your Grafana Cloud credentials:
- **Region**: `prod-us-east-0`
- **Instance ID**: `1086929`
- **Metrics**: Scrapes from `localhost:8080/metrics` every 15s
- **Traces**: Receives OTLP on ports 4318 (HTTP) and 4317 (gRPC)

## Step 3: Start Grafana Alloy

```bash
cd /Users/ariatron/farkle-fun
alloy run alloy-config.alloy
```

You should see:
```
ts=2026-02-02T16:13:00Z level=info msg="starting alloy"
ts=2026-02-02T16:13:00Z component=prometheus.scrape.farkle_metrics msg="scrape target discovered"
ts=2026-02-02T16:13:00Z component=otelcol.receiver.otlp.farkle_traces msg="starting OTLP receiver"
```

**Tip**: Leave this terminal open, or run as a service (see below).

### Alternative: Run as Service

#### macOS
```bash
# Copy config to default location
sudo mkdir -p /etc/alloy
sudo cp alloy-config.alloy /etc/alloy/config.alloy

# Start service
brew services start alloy

# Check status
brew services info alloy
```

#### Linux
```bash
# Copy config to default location
sudo cp alloy-config.alloy /etc/alloy/config.alloy

# Start service
sudo systemctl start alloy
sudo systemctl enable alloy  # Start on boot

# Check status
sudo systemctl status alloy
```

## Step 4: Start Farkle Server

In a **new terminal**:

```bash
cd /Users/ariatron/farkle-fun

# Set OTLP endpoint to send traces to Alloy
export JAEGER_ENDPOINT="localhost:4318"

# Start server
go run cmd/server/main.go
```

You should see:
```
🎲 Farkle Server started at http://localhost:8080
📊 Metrics available at http://localhost:8080/metrics
🔍 Traces will be sent to Jaeger (if running)
```

## Step 5: Generate Some Data

### Option 1: Play Manually
Open http://localhost:8080 in your browser and play!

### Option 2: Run K6 Tests
In a **third terminal**:
```bash
cd /Users/ariatron/farkle-fun
k6 run tests/k6/smoke-test.js
```

This will generate realistic traffic and create metrics/traces.

## Step 6: View Data in Grafana Cloud

### Open Alloy UI (Local Verification)
Open http://localhost:12345 to see:
- Metrics being collected
- Traces being received
- Connection status to Grafana Cloud

### View in Grafana Cloud
1. Go to: https://yourusername.grafana.net
2. Click **Explore** in the sidebar

### View Metrics
1. Select **Prometheus** as data source
2. Try these queries:

```promql
# Request rate
rate(farkle_http_requests_total[5m])

# Game rolls per second
rate(farkle_game_rolls_total[1m])

# Farkle rate
rate(farkle_game_farkles_total[5m]) / rate(farkle_game_rolls_total[5m])

# P95 latency
histogram_quantile(0.95, rate(farkle_http_request_duration_seconds_bucket[5m]))
```

### View Traces
1. Select **Tempo** as data source
2. Click **Search**
3. Filter by:
   - Service: `farkle-game`
   - Operation: `POST /api/roll`
4. Click on a trace to see the full span details

## Step 7: Import Dashboard

1. In Grafana Cloud, go to **Dashboards** → **New** → **Import**
2. Click **Upload JSON file**
3. Select: `grafana-dashboards/farkle-dashboard.json`
4. Click **Import**

Your dashboard is now live! 🎉

## Architecture

```
┌─────────────┐
│ Farkle App  │
│  :8080      │
└──────┬──────┘
       │ metrics (scraped)
       │ traces (pushed to :4318)
       ▼
┌─────────────────┐
│ Grafana Alloy   │
│  :12345 (UI)    │
│  :4318 (OTLP)   │
└──────┬──────────┘
       │
       │ remote_write
       ▼
┌─────────────────┐
│ Grafana Cloud   │
│ - Prometheus    │
│ - Tempo         │
│ - Grafana       │
└─────────────────┘
```

## Troubleshooting

### No Metrics Appearing

**1. Check Alloy is running:**
```bash
curl http://localhost:12345
```

**2. Check Alloy UI:**
- Open http://localhost:12345
- Look at `prometheus.scrape.farkle_metrics` component
- Should show "targets discovered"

**3. Check Farkle metrics endpoint:**
```bash
curl http://localhost:8080/metrics | grep farkle_
```

**4. Check Alloy logs:**
```bash
# If running in foreground, check terminal output
# If running as service:
brew services info alloy  # macOS
sudo journalctl -u alloy -f  # Linux
```

### No Traces Appearing

**1. Check Farkle is sending to correct endpoint:**
```bash
# Should be localhost:4318 when using Alloy
echo $JAEGER_ENDPOINT
```

**2. Check Alloy OTLP receiver:**
- Open http://localhost:12345
- Look at `otelcol.receiver.otlp.farkle_traces`
- Should show "receiver started"

**3. Generate a test trace:**
```bash
curl -X POST http://localhost:8080/api/roll \
  -H "Content-Type: application/json" \
  -d '{"dice_to_keep":[]}'
```

**4. Check Alloy can reach Tempo:**
```bash
# Test connection
curl -v https://tempo-prod-us-east-0.grafana.net:443
```

### Alloy Won't Start

**1. Validate config:**
```bash
alloy run --check alloy-config.alloy
```

**2. Format config:**
```bash
alloy fmt alloy-config.alloy
```

**3. Check for port conflicts:**
```bash
lsof -i :12345  # Alloy UI
lsof -i :4318   # OTLP HTTP
lsof -i :4317   # OTLP gRPC
```

## Next Steps

1. **Set up alerts** in Grafana Cloud:
   - High error rate: `sum(rate(farkle_http_requests_total{status=~"5.."}[5m])) / sum(rate(farkle_http_requests_total[5m])) > 0.05`
   - High latency: `histogram_quantile(0.95, rate(farkle_http_request_duration_seconds_bucket[5m])) > 1`
   - Service down: `up{job="farkle-app"} == 0`

2. **Customize the dashboard**:
   - Add custom panels
   - Set up annotations
   - Create dashboard variables

3. **Enable Loki logs** (optional):
   - Uncomment the Loki section in `alloy-config.alloy`
   - Configure log file paths
   - Restart Alloy

4. **Production deployment**:
   - Set up Alloy as a system service
   - Configure automatic restarts
   - Set up log rotation
   - Monitor Alloy itself

## Resources

- **Alloy Setup Guide**: [docs/ALLOY_SETUP.md](docs/ALLOY_SETUP.md)
- **Grafana Cloud Guide**: [docs/GRAFANA_CLOUD.md](docs/GRAFANA_CLOUD.md)
- **Local Observability**: [docs/OBSERVABILITY.md](docs/OBSERVABILITY.md)
- **Alloy Documentation**: https://grafana.com/docs/alloy/
- **Grafana Cloud Docs**: https://grafana.com/docs/grafana-cloud/

## Support

- **Alloy Community**: https://community.grafana.com/c/alloy/
- **GitHub Issues**: https://github.com/grafana/alloy/issues
- **Grafana Support**: Available for paid plans
