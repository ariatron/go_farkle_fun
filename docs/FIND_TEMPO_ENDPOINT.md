# How to Find Your Tempo Endpoint

The 403 errors you're seeing mean the Tempo endpoint or credentials are incorrect. Here's how to find the right ones:

## Step 1: Log into Grafana Cloud

Go to: https://grafana.com/auth/sign-in/

## Step 2: Navigate to Tempo Settings

1. Click on your **Stack** name in the top navigation
2. In the left sidebar, look for **Tempo** or **Traces**
3. Click on **Details** or **Configuration**

## Step 3: Find OTLP Endpoint

Look for one of these:

### Option A: OTLP/HTTP Endpoint
```
Endpoint: https://tempo-XXXXX.grafana.net/otlp
OR
Endpoint: https://otlp-gateway-prod-us-east-0.grafana.net/otlp
```

### Option B: OTLP/gRPC Endpoint
```
Endpoint: tempo-XXXXX.grafana.net:443
```

### Option C: Direct Tempo Endpoint
```
Endpoint: tempo-prod-XX-prod-us-east-0.grafana.net:443
```

## Step 4: Find Instance ID for Tempo

This might be different from your Prometheus instance ID!

Look for:
- **Instance ID**: (a number like 123456)
- **User**: (might be the same as instance ID)

## Step 5: Update Alloy Config

Once you have the correct endpoint and credentials, update `alloy-config.alloy`:

### If using OTLP/HTTP endpoint:
```alloy
otelcol.exporter.otlp "grafana_cloud" {
  client {
    endpoint = "otlp-gateway-prod-us-east-0.grafana.net:443"  // No /otlp suffix for gRPC

    auth = otelcol.auth.basic.grafana_cloud.handler
  }
}

otelcol.auth.basic "grafana_cloud" {
  username = "YOUR_TEMPO_INSTANCE_ID"  // Might be different from Prometheus ID
  password = "YOUR_API_KEY"
}
```

### If Tempo has a different auth format:
Some Tempo instances use header-based auth:

```alloy
otelcol.exporter.otlp "grafana_cloud" {
  client {
    endpoint = "tempo-XXXXX.grafana.net:443"

    headers = {
      "Authorization" = "Bearer YOUR_API_KEY"
    }
  }
}
```

## Common Issues

### Issue 1: Wrong Region
Your Prometheus uses `prod-13-prod-us-east-0`, but Tempo might use a different pod/region like:
- `prod-03-prod-us-east-0`
- `prod-us-east-0` (no number)

### Issue 2: Different Instance ID
Tempo and Prometheus can have different instance IDs in Grafana Cloud.

### Issue 3: Need Different API Key
Sometimes you need to create a separate API key with "Traces" permissions.

## Quick Test

After updating the config, restart Alloy and check logs:

```bash
pkill -f "alloy run"
./scripts/start-alloy.sh

# Watch for errors
tail -f /tmp/alloy.log | grep -i "tempo\|otlp\|403"
```

If you still see 403 errors, the credentials are wrong.
If you see connection errors, the endpoint is wrong.

## Alternative: Use Grafana Agent Auto-Config Script

You mentioned having this command from Grafana Cloud:

```bash
GCLOUD_HOSTED_METRICS_URL="..." ... /bin/sh -c "$(curl -fsSL ...install-macos-homebrew.sh)"
```

That script might have configured Tempo correctly. Look for a generated config file at:
- `/opt/homebrew/etc/alloy/config.alloy`
- Or wherever the script put it

Check that file for the correct Tempo configuration!
