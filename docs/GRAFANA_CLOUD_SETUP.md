# Grafana Cloud Setup Guide

Step-by-step guide to configure Farkle observability with **your own** Grafana Cloud account.

## Prerequisites

- Grafana Cloud account (free tier available at https://grafana.com/auth/sign-up)
- Farkle game running locally
- Grafana Alloy installed (`brew install grafana/grafana/alloy` on macOS)

---

## Step 1: Get Your Grafana Cloud Credentials

### 1.1 Log into Grafana Cloud

Go to https://grafana.com and sign in to your account.

### 1.2 Navigate to Your Stack

1. Click on your organization name (top left)
2. Select "Stacks"
3. Click on your stack (or create one if you don't have any)

### 1.3 Collect Prometheus Credentials

1. In your stack, click **"Prometheus"** → **"Details"** → **"Remote Write Endpoint"**
2. Note down:
   - **URL:** `https://prometheus-YOUR-REGION.grafana.net/api/prom/push`
     - Example: `https://prometheus-prod-us-east-0.grafana.net/api/prom/push`
   - **Instance ID / Username:** A number like `1234567`
   - Your region code (e.g., `prod-us-east-0`, `prod-eu-west-0`)

### 1.4 Collect Tempo Credentials (for traces)

1. Click **"Tempo"** → **"Details"**
2. Note down:
   - **URL:** `tempo-YOUR-REGION.grafana.net:443`
     - Example: `tempo-prod-us-east-0.grafana.net:443`
   - **Instance ID / Username:** A number like `7654321`

### 1.5 Collect Loki Credentials (for logs)

1. Click **"Loki"** → **"Details"**
2. Note down:
   - **URL:** `https://logs-YOUR-REGION.grafana.net/loki/api/v1/push`
     - Example: `https://logs-prod-006.grafana.net/loki/api/v1/push`
   - **Instance ID / Username:** A number like `9876543`

### 1.6 Create API Token

1. Go to **"Settings"** → **"API Keys"** (or **"Access Policies"** in newer UI)
2. Click **"Create API Key"**
3. Name: `farkle-observability`
4. Role: **MetricsPublisher** (or **Admin** for full access)
5. Click **"Create"**
6. **IMPORTANT:** Copy the API key immediately! It looks like:
   ```
   glc_eyJvIjoiMTIzNDU2IiwibiI6ImZhcmtsZS1rZXkiLC...
   ```
7. Store it safely - you won't be able to see it again

---

## Step 2: Configure Alloy

### 2.1 Copy the Example Configuration

```bash
cd /path/to/farkle-fun
cp alloy-config.alloy.example alloy-config.alloy
```

### 2.2 Edit Configuration

Open `alloy-config.alloy` and replace the placeholders:

```alloy
// Prometheus section
prometheus.remote_write "grafana_cloud" {
  endpoint {
    url = "https://prometheus-prod-us-east-0.grafana.net/api/prom/push"  // ← Your URL

    basic_auth {
      username = "1234567"  // ← Your Prometheus Instance ID
      password = "glc_eyJvIjoiMTIzNDU2..."  // ← Your API Token
    }
  }
}

// Tempo section
otelcol.exporter.otlp "grafana_cloud" {
  client {
    endpoint = "tempo-prod-us-east-0.grafana.net:443"  // ← Your Tempo URL

    auth = otelcol.auth.basic.grafana_cloud.handler
  }
}

otelcol.auth.basic "grafana_cloud" {
  username = "7654321"  // ← Your Tempo Instance ID
  password = "glc_eyJvIjoiMTIzNDU2..."  // ← Same API Token
}

// Loki section
loki.write "grafana_cloud" {
  endpoint {
    url = "https://logs-prod-006.grafana.net/loki/api/v1/push"  // ← Your Loki URL

    basic_auth {
      username = "9876543"  // ← Your Loki Instance ID
      password = "glc_eyJvIjoiMTIzNDU2..."  // ← Same API Token
    }
  }
}
```

### 2.3 Optional: Customize Service Name

If you want to use a different service name instead of `farkle-game`:

1. Find all instances of `service = "farkle-game"` in the config
2. Replace with your desired name (e.g., `my-farkle-app`)
3. Update dashboard TraceQL/LogQL queries accordingly

---

## Step 3: Start the Observability Stack

### 3.1 Start Everything

```bash
./scripts/start-all.sh
```

Or manually:

```bash
# Terminal 1: Start Alloy
alloy run alloy-config.alloy

# Terminal 2: Start Farkle Server
export JAEGER_ENDPOINT="localhost:4318"
go run cmd/server/main.go

# Terminal 3: Start Traffic Generation (optional)
./scripts/continuous-traffic.sh start
```

### 3.2 Verify Data Flow

**Check Alloy is collecting:**
```bash
curl http://localhost:12345/metrics | grep prometheus_remote_write_wal_samples_appended_total
```

**Check server is generating metrics:**
```bash
curl http://localhost:8080/metrics | grep farkle_game_rolls_total
```

**Check traces are being sent:**
```bash
curl http://localhost:12345/metrics | grep otelcol_exporter_sent_spans_total
```

---

## Step 4: Import Dashboards

### 4.1 Import to Grafana Cloud

1. Go to your Grafana Cloud dashboard
2. Click **"Dashboards"** → **"New"** → **"Import"**
3. Click **"Upload JSON file"**
4. Select `grafana-dashboards/farkle-simple-dashboard.json` or `farkle-complete-dashboard.json`
5. When prompted, configure data sources:
   - **Prometheus:** Select your Prometheus data source (usually auto-detected)
   - **Tempo:** Select your Tempo data source (complete dashboard only)
   - **Loki:** Select your Loki data source (complete dashboard only)
6. Click **"Import"**

### 4.2 Verify Data

Wait 1-2 minutes for data to appear, then check:

**Metrics (should appear immediately):**
- Go to dashboard
- Panels should show request rates, game stats, etc.

**Traces (may take 30-60 seconds):**
- Go to **Explore** → Select **Tempo**
- Query: `{resource.service.name="farkle-game"}`
- You should see traces with proper timing

**Logs (may take 30-60 seconds):**
- Go to **Explore** → Select **Loki**
- Query: `{service="farkle-game"} | json`
- You should see structured JSON logs

---

## Troubleshooting

### No Metrics Appearing

**Check Alloy logs:**
```bash
tail -f logs/alloy.log | grep -i error
```

**Common issues:**
- Wrong credentials → You'll see 401/403 errors
- Wrong URLs → You'll see connection errors
- Firewall blocking → Check if you can `curl` the Grafana Cloud URLs

**Fix:**
1. Verify credentials in Grafana Cloud portal
2. Check URLs match your region exactly
3. Ensure API token has correct permissions

### No Traces Appearing

**Check trace export:**
```bash
curl http://localhost:12345/metrics | grep otelcol_exporter_send_failed_spans_total
```

If failures > 0:
- Check Tempo endpoint and credentials
- Verify Tempo instance ID is correct
- Wait 1-2 minutes for indexing

### No Logs Appearing

**Check log file exists:**
```bash
ls -la /tmp/farkle-app.log
```

**Check Alloy is reading logs:**
```bash
curl http://localhost:12345/metrics | grep loki_source
```

**Common issues:**
- Log file doesn't exist → Start server first
- Wrong path → Check `loki.source.file` config
- Loki credentials wrong → Check `loki.write` config

### WAL Errors

If you see `"Segments: segments are not sequential"`:

```bash
./scripts/stop-all.sh
rm -rf data-alloy/prometheus.remote_write.grafana_cloud/wal
./scripts/start-all.sh
```

---

## Security Best Practices

### 1. Protect Your Credentials

**Never commit:**
- `alloy-config.alloy` (contains your API token)
- `.env` files with credentials

**Always use:**
- `.gitignore` to exclude sensitive files (already configured)
- Environment variables for CI/CD
- Secret management tools in production

### 2. API Token Permissions

Create separate tokens for different environments:
- **Development:** `MetricsPublisher` role
- **Production:** `MetricsPublisher` role (not Admin)
- **CI/CD:** Dedicated token with minimal permissions

### 3. Rotate Tokens

If your token is compromised:
1. Revoke old token in Grafana Cloud
2. Create new token
3. Update `alloy-config.alloy`
4. Restart Alloy

---

## Cost Management (Free Tier Limits)

Grafana Cloud free tier includes:
- **Metrics:** 10,000 series
- **Traces:** 50GB per month
- **Logs:** 50GB per month

### Monitor Usage

Check your usage at: https://grafana.com/orgs/YOUR_ORG/billing

### Reduce Usage

**If approaching limits:**

1. **Reduce scrape frequency:**
   ```alloy
   scrape_interval = "30s"  // Instead of 15s
   ```

2. **Sample traces:**
   ```go
   // In internal/observability/tracing.go
   sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.5))  // 50% sampling
   ```

3. **Filter logs:**
   ```alloy
   // Only send ERROR and WARN logs
   stage.match {
     selector = "{level=~\"ERROR|WARN\"}"
   }
   ```

4. **Stop continuous traffic:**
   ```bash
   ./scripts/continuous-traffic.sh stop
   ```

---

## Next Steps

- ✅ Setup complete? Explore your dashboards!
- 📊 Want custom metrics? See `docs/OBSERVABILITY.md`
- 🔍 Deep dive into traces? Check `docs/CONTINUOUS_TRAFFIC.md`
- 🤝 Share your setup? Contribute improvements!

---

## Support

Having issues? Check:
- [Main README](../README.md)
- [Grafana Cloud Docs](https://grafana.com/docs/)
- [Alloy Documentation](https://grafana.com/docs/alloy/)
- [GitHub Issues](https://github.com/ariatron/go_farkle_fun/issues)
