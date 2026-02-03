# Migrating from Grafana Agent to Alloy

Grafana Agent is deprecated in favor of **Grafana Alloy**. This guide helps you understand the differences and migrate your setup.

## Why Migrate?

| Aspect | Grafana Agent | Grafana Alloy |
|--------|--------------|---------------|
| **Status** | ❌ Deprecated (EOL planned) | ✅ Active Development |
| **Config Format** | YAML | Alloy Config Language (HCL-like) |
| **Performance** | Good | Better (optimized pipelines) |
| **Debugging** | Logs only | Built-in UI + Logs |
| **Flexibility** | Static pipelines | Dynamic, composable components |
| **Learning Curve** | Easy | Moderate |
| **Documentation** | Legacy | Modern & Comprehensive |

## Key Differences

### 1. Configuration Language

**Grafana Agent (YAML):**
```yaml
metrics:
  configs:
    - name: farkle
      scrape_configs:
        - job_name: 'farkle-app'
          static_configs:
            - targets: ['localhost:8080']
```

**Grafana Alloy (Alloy Config):**
```alloy
prometheus.scrape "farkle_metrics" {
  targets = [{
    __address__ = "localhost:8080",
  }]
  forward_to = [prometheus.remote_write.grafana_cloud.receiver]
}
```

### 2. Component Model

**Agent**: Static configuration, reload required for changes

**Alloy**: Dynamic components that can be added/removed at runtime

### 3. Debugging

**Agent**: Check logs and test endpoints manually

**Alloy**: Built-in UI at `http://localhost:12345` shows:
- Component status
- Data flow
- Metrics being collected
- Error messages

## Migration Steps

### Step 1: Install Alloy

**macOS:**
```bash
brew install grafana/grafana/alloy
```

**Linux:**
```bash
sudo apt-get install alloy
```

### Step 2: Convert Configuration

**Automatic Conversion:**
```bash
# Alloy can convert Agent YAML to Alloy format
alloy convert --source-format=static \
  --output=alloy-config.alloy \
  grafana-agent-config.yaml
```

**Manual Conversion:**

Your Farkle game already has an Alloy config at [`alloy-config.alloy`](/Users/ariatron/farkle-fun/alloy-config.alloy), so no conversion needed!

### Step 3: Test the Configuration

```bash
# Validate syntax
alloy run --check alloy-config.alloy

# Format the config
alloy fmt alloy-config.alloy

# Run in foreground to test
alloy run alloy-config.alloy
```

### Step 4: Update Your Farkle Server

**Before (with Agent):**
```bash
export JAEGER_ENDPOINT="tempo-prod-us-east-0.grafana.net:443"
go run cmd/server/main.go
```

**After (with Alloy):**
```bash
# Traces now go through Alloy's local OTLP receiver
export JAEGER_ENDPOINT="localhost:4318"
go run cmd/server/main.go
```

### Step 5: Run as Service (Optional)

**Stop Agent service:**
```bash
# macOS
brew services stop grafana-agent

# Linux
sudo systemctl stop grafana-agent
sudo systemctl disable grafana-agent
```

**Start Alloy service:**
```bash
# Copy config to default location
sudo mkdir -p /etc/alloy
sudo cp alloy-config.alloy /etc/alloy/config.alloy

# macOS
brew services start alloy

# Linux
sudo systemctl start alloy
sudo systemctl enable alloy
```

### Step 6: Verify

1. **Check Alloy UI**: http://localhost:12345
2. **Check metrics**: `curl http://localhost:8080/metrics`
3. **Generate traffic**: Play the game or run K6 tests
4. **Check Grafana Cloud**: Verify data is arriving

## Configuration Mapping

### Prometheus Scraping

**Agent YAML:**
```yaml
metrics:
  global:
    scrape_interval: 15s
  configs:
    - name: farkle
      scrape_configs:
        - job_name: 'farkle-app'
          static_configs:
            - targets: ['localhost:8080']
          metrics_path: '/metrics'
      remote_write:
        - url: https://prometheus-REGION.grafana.net/api/prom/push
          basic_auth:
            username: 123456
            password: glc_xxxxx
```

**Alloy Config:**
```alloy
prometheus.scrape "farkle_metrics" {
  targets = [{
    __address__ = "localhost:8080",
  }]
  forward_to = [prometheus.remote_write.grafana_cloud.receiver]
  scrape_interval = "15s"
  metrics_path = "/metrics"
}

prometheus.remote_write "grafana_cloud" {
  endpoint {
    url = "https://prometheus-REGION.grafana.net/api/prom/push"
    basic_auth {
      username = "123456"
      password = "glc_xxxxx"
    }
  }
}
```

### OTLP Traces

**Agent YAML:**
```yaml
traces:
  configs:
    - name: default
      receivers:
        otlp:
          protocols:
            http:
              endpoint: "0.0.0.0:4318"
      remote_write:
        - endpoint: tempo-REGION.grafana.net:443
          basic_auth:
            username: 123456
            password: glc_xxxxx
```

**Alloy Config:**
```alloy
otelcol.receiver.otlp "farkle_traces" {
  http { endpoint = "0.0.0.0:4318" }
  grpc { endpoint = "0.0.0.0:4317" }

  output {
    traces = [otelcol.exporter.otlp.grafana_cloud.input]
  }
}

otelcol.exporter.otlp "grafana_cloud" {
  client {
    endpoint = "tempo-REGION.grafana.net:443"
    auth = otelcol.auth.basic.grafana_cloud.handler
  }
}

otelcol.auth.basic "grafana_cloud" {
  username = "123456"
  password = "glc_xxxxx"
}
```

## Common Issues

### Port Conflicts

**Problem**: Alloy uses different default ports than Agent

**Solution**:
- Alloy UI: `12345` (Agent used `12345` too)
- OTLP HTTP: `4318` (same as Agent)
- OTLP gRPC: `4317` (same as Agent)

### Configuration Errors

**Problem**: YAML config doesn't work with Alloy

**Solution**: Use the conversion tool:
```bash
alloy convert --source-format=static grafana-agent-config.yaml
```

### Service Not Starting

**Problem**: Alloy service fails to start

**Solution**: Check config location and permissions:
```bash
# Verify config exists
ls -la /etc/alloy/config.alloy

# Validate config
alloy run --check /etc/alloy/config.alloy

# Check logs
sudo journalctl -u alloy -n 50
```

## Benefits of Alloy

### 1. Built-in UI

Access http://localhost:12345 to see:
- Real-time component status
- Data flow visualization
- Metrics being collected
- Trace pipeline status
- Error messages and warnings

### 2. Better Performance

- More efficient data processing
- Lower memory usage
- Faster startup times
- Optimized remote write

### 3. Easier Debugging

**Before (Agent):**
```bash
# Check if metrics are being scraped
curl http://localhost:8080/metrics

# Check Agent logs
tail -f /var/log/grafana-agent.log

# Guess at what's wrong
```

**After (Alloy):**
```bash
# Check Alloy UI
open http://localhost:12345

# See exactly what's happening:
# - Are targets discovered?
# - Is data being sent?
# - Any errors?
```

### 4. Dynamic Configuration

Components can be added/removed without full restart:
```alloy
// Add new scrape target dynamically
prometheus.scrape "new_service" {
  targets = discovery.kubernetes.pods.targets
  forward_to = [prometheus.remote_write.grafana_cloud.receiver]
}
```

### 5. Better Documentation

- Comprehensive guides: https://grafana.com/docs/alloy/
- Active community: https://community.grafana.com/c/alloy/
- Regular updates and improvements

## Timeline

- **Grafana Agent**: Deprecated, maintenance mode only
- **Grafana Alloy**: Active development, recommended for all new deployments
- **Your Farkle Game**: Already configured for Alloy! 🎉

## Next Steps

1. ✅ Install Alloy
2. ✅ Use the provided `alloy-config.alloy`
3. ✅ Start Alloy: `alloy run alloy-config.alloy`
4. ✅ Update Farkle: `export JAEGER_ENDPOINT="localhost:4318"`
5. ✅ Check Alloy UI: http://localhost:12345
6. ✅ Verify data in Grafana Cloud

## Resources

- **Alloy Setup**: [docs/ALLOY_SETUP.md](ALLOY_SETUP.md)
- **Quick Start**: [GRAFANA_CLOUD_QUICKSTART.md](../GRAFANA_CLOUD_QUICKSTART.md)
- **Alloy Docs**: https://grafana.com/docs/alloy/
- **Migration Guide**: https://grafana.com/docs/alloy/latest/migration-guide/
