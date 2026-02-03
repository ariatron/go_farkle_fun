# Grafana Alloy Setup Guide

**Grafana Alloy** is the successor to Grafana Agent and the recommended way to collect and send telemetry data to Grafana Cloud.

## What is Alloy?

Grafana Alloy is a vendor-neutral distribution of the OpenTelemetry Collector that:
- Collects metrics, logs, and traces
- Processes and transforms telemetry data
- Sends data to Grafana Cloud or other backends
- Uses a powerful configuration language (Alloy configuration syntax)
- Replaces the deprecated Grafana Agent

## Installation

### macOS

```bash
brew install grafana/grafana/alloy
```

### Linux (Ubuntu/Debian)

```bash
# Add Grafana repository
sudo mkdir -p /etc/apt/keyrings/
wget -q -O - https://apt.grafana.com/gpg.key | gpg --dearmor | sudo tee /etc/apt/keyrings/grafana.gpg > /dev/null
echo "deb [signed-by=/etc/apt/keyrings/grafana.gpg] https://apt.grafana.com stable main" | sudo tee /etc/apt/sources.list.d/grafana.list

# Install Alloy
sudo apt-get update
sudo apt-get install -y alloy
```

### Linux (RHEL/CentOS/Fedora)

```bash
# Add Grafana repository
cat <<EOF | sudo tee /etc/yum.repos.d/grafana.repo
[grafana]
name=grafana
baseurl=https://rpm.grafana.com
repo_gpgcheck=1
enabled=1
gpgcheck=1
gpgkey=https://rpm.grafana.com/gpg.key
sslverify=1
sslcacert=/etc/pki/tls/certs/ca-bundle.crt
EOF

# Install Alloy
sudo yum install -y alloy
```

### Docker

```bash
docker run -v $(pwd)/alloy-config.alloy:/etc/alloy/config.alloy \
  -p 4318:4318 -p 4317:4317 -p 12345:12345 \
  grafana/alloy:latest run \
  --server.http.listen-addr=0.0.0.0:12345 \
  --storage.path=/var/lib/alloy/data \
  /etc/alloy/config.alloy
```

### Windows

Download from: https://github.com/grafana/alloy/releases

## Configuration

Your Farkle game is pre-configured with Alloy. The configuration is in [`alloy-config.alloy`](/Users/ariatron/farkle-fun/alloy-config.alloy).

### Configuration Structure

```alloy
// Scrape Prometheus metrics from Farkle app
prometheus.scrape "farkle_metrics" {
  targets = [{
    __address__ = "localhost:8080",
    service     = "farkle-game",
  }]
  forward_to = [prometheus.remote_write.grafana_cloud.receiver]
}

// Send metrics to Grafana Cloud
prometheus.remote_write "grafana_cloud" {
  endpoint {
    url = "https://prometheus-REGION.grafana.net/api/prom/push"
    basic_auth {
      username = "YOUR_INSTANCE_ID"
      password = "YOUR_API_KEY"
    }
  }
}

// Receive OTLP traces from Farkle app
otelcol.receiver.otlp "farkle_traces" {
  http { endpoint = "0.0.0.0:4318" }
  grpc { endpoint = "0.0.0.0:4317" }
  output {
    traces = [otelcol.processor.batch.default.input]
  }
}

// Send traces to Grafana Cloud Tempo
otelcol.exporter.otlp "grafana_cloud" {
  client {
    endpoint = "tempo-REGION.grafana.net:443"
    auth = otelcol.auth.basic.grafana_cloud.handler
  }
}
```

## Running Alloy

### Run Mode (Foreground)

```bash
cd /Users/ariatron/farkle-fun
alloy run alloy-config.alloy
```

You should see:
```
ts=2026-02-02T16:13:00Z level=info msg="starting alloy"
ts=2026-02-02T16:13:00Z level=info component=prometheus.scrape.farkle_metrics msg="scrape target discovered"
ts=2026-02-02T16:13:00Z level=info component=otelcol.receiver.otlp.farkle_traces msg="starting OTLP receiver"
```

### Service Mode (Background)

#### macOS (via Homebrew)

```bash
# Start Alloy service
brew services start alloy

# Check status
brew services info alloy

# Stop Alloy service
brew services stop alloy

# View logs
tail -f /opt/homebrew/var/log/alloy.log
```

#### Linux (systemd)

```bash
# Start Alloy service
sudo systemctl start alloy

# Enable on boot
sudo systemctl enable alloy

# Check status
sudo systemctl status alloy

# View logs
sudo journalctl -u alloy -f
```

**Note**: When running as a service, copy your config to the default location:
```bash
sudo cp alloy-config.alloy /etc/alloy/config.alloy
sudo systemctl restart alloy
```

### Docker Compose

Add to your `docker-compose.yml`:

```yaml
services:
  alloy:
    image: grafana/alloy:latest
    container_name: farkle-alloy
    ports:
      - "4318:4318"  # OTLP HTTP
      - "4317:4317"  # OTLP gRPC
      - "12345:12345"  # Alloy UI
    volumes:
      - ./alloy-config.alloy:/etc/alloy/config.alloy:ro
    command:
      - run
      - --server.http.listen-addr=0.0.0.0:12345
      - --storage.path=/var/lib/alloy/data
      - /etc/alloy/config.alloy
    restart: unless-stopped
```

## Verification

### 1. Check Alloy UI

Open http://localhost:12345 in your browser to see:
- Component status
- Metrics being scraped
- Traces being received
- Connection status to Grafana Cloud

### 2. Check Metrics Collection

```bash
# View metrics that Alloy is collecting
curl http://localhost:12345/metrics | grep farkle_
```

### 3. Check Grafana Cloud

1. Go to your Grafana Cloud instance
2. **Explore** → **Prometheus**
3. Query: `farkle_http_requests_total`

If you see data, it's working! 🎉

## Differences from Grafana Agent

| Feature | Grafana Agent | Grafana Alloy |
|---------|--------------|---------------|
| **Config Format** | YAML | Alloy Config Language |
| **Status** | Deprecated | Active Development |
| **Components** | Static | Dynamic Pipelines |
| **Performance** | Good | Better |
| **Debugging** | Logs only | Built-in UI + Logs |
| **Flexibility** | Limited | Highly Flexible |

## Migration from Grafana Agent

If you were using `grafana-agent-config.yaml`, Alloy provides automatic conversion:

```bash
# Convert old config to Alloy format
alloy convert --source-format=static \
  --output=alloy-config.alloy \
  grafana-agent-config.yaml
```

**Note**: Your Farkle game already has an Alloy config, so no conversion needed!

## Troubleshooting

### Alloy won't start

**Check the config syntax:**
```bash
alloy fmt alloy-config.alloy
```

**Validate the config:**
```bash
alloy run --check alloy-config.alloy
```

### No metrics in Grafana Cloud

**Check Alloy logs:**
```bash
# If running in foreground, check the output
# Look for "remote_write" errors
```

**Check Alloy UI:**
- Go to http://localhost:12345
- Look at the `prometheus.scrape.farkle_metrics` component
- Check for errors or warnings

**Test Farkle metrics endpoint:**
```bash
curl http://localhost:8080/metrics | head -20
```

### No traces in Grafana Cloud

**Check OTLP receiver:**
```bash
# Check Alloy UI at http://localhost:12345
# Look at otelcol.receiver.otlp.farkle_traces component
```

**Test trace endpoint:**
```bash
# Make a request to generate a trace
curl -X POST http://localhost:8080/api/roll \
  -H "Content-Type: application/json" \
  -d '{"dice_to_keep":[]}'
```

**Check Farkle app is sending to correct endpoint:**
```bash
# Farkle should send traces to localhost:4318
export JAEGER_ENDPOINT="localhost:4318"
go run cmd/server/main.go
```

### Authentication errors

**Verify credentials:**
```bash
# Check your alloy-config.alloy file
grep "username\|password" alloy-config.alloy

# Test Prometheus endpoint
curl -v -X POST \
  -u "YOUR_INSTANCE_ID:YOUR_API_KEY" \
  https://prometheus-prod-us-east-0.grafana.net/api/prom/push
```

## Advanced Configuration

### Adding Service Discovery

```alloy
// Discover services via Docker
discovery.docker "services" {
  host = "unix:///var/run/docker.sock"
}

prometheus.scrape "discovered" {
  targets    = discovery.docker.services.targets
  forward_to = [prometheus.remote_write.grafana_cloud.receiver]
}
```

### Adding Processors

```alloy
// Add resource attributes to traces
otelcol.processor.resource "add_environment" {
  attributes {
    environment = "production"
    team        = "platform"
  }

  output {
    traces = [otelcol.processor.batch.default.input]
  }
}
```

### Filtering Metrics

```alloy
prometheus.relabel "filter_metrics" {
  forward_to = [prometheus.remote_write.grafana_cloud.receiver]

  rule {
    source_labels = ["__name__"]
    regex         = "farkle_.*"
    action        = "keep"
  }
}
```

## Resources

- **Alloy Documentation**: https://grafana.com/docs/alloy/
- **Alloy GitHub**: https://github.com/grafana/alloy
- **Configuration Reference**: https://grafana.com/docs/alloy/latest/reference/
- **Migration Guide**: https://grafana.com/docs/alloy/latest/migration-guide/
- **Community**: https://community.grafana.com/

## Quick Reference Commands

```bash
# Install Alloy (macOS)
brew install grafana/grafana/alloy

# Run Alloy
alloy run alloy-config.alloy

# Validate config
alloy run --check alloy-config.alloy

# Format config
alloy fmt alloy-config.alloy

# Convert from Grafana Agent
alloy convert --source-format=static grafana-agent-config.yaml

# Run as service (macOS)
brew services start alloy

# Run as service (Linux)
sudo systemctl start alloy

# View Alloy UI
open http://localhost:12345
```
