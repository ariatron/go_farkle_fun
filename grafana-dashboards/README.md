# Grafana Dashboards for Farkle Game

This directory contains pre-built Grafana dashboards for monitoring your Farkle game with complete observability.

## Available Dashboards

### 1. **farkle-simple-dashboard.json** (Recommended for beginners)
- **Panels:** 7
- **Data Sources:** Prometheus only
- **Best for:** Quick setup, basic monitoring
- **Metrics:** HTTP requests, latency, game stats

### 2. **farkle-complete-dashboard.json** (Full observability)
- **Panels:** 27
- **Data Sources:** Prometheus, Tempo (traces), Loki (logs)
- **Best for:** Production monitoring, troubleshooting, demos
- **Includes:** Metrics, traces, logs, and correlations

---

## Quick Start

### For Grafana Cloud Users

1. **Import Dashboard:**
   - Go to your Grafana Cloud instance
   - Navigate to **Dashboards** → **New** → **Import**
   - Click **Upload JSON file**
   - Select `farkle-simple-dashboard.json` or `farkle-complete-dashboard.json`

2. **Configure Data Sources:**
   - When prompted, select your data sources:
     - **Prometheus:** Your Prometheus data source (usually named "prometheus" or "grafanacloud-*-prom")
     - **Tempo:** Your Tempo data source (for complete dashboard)
     - **Loki:** Your Loki data source (for complete dashboard)

3. **Done!** The dashboard will auto-populate with your Farkle game metrics.

---

## Dashboard Configuration

### Data Source Requirements

**Both dashboards use generic data source types** without hardcoded UIDs, so they work with any Grafana setup:

| Dashboard | Required Data Sources |
|-----------|----------------------|
| Simple | Prometheus |
| Complete | Prometheus, Tempo, Loki |

**Note:** If you don't have Tempo or Loki configured, use the simple dashboard or remove the trace/log panels from the complete dashboard.

---

## Customization

### Panel Queries

All panels use standardized metric names from the Farkle application:

**HTTP Metrics:**
- `farkle_http_requests_total` - Request counter by endpoint, method, status
- `farkle_http_request_duration_seconds_bucket` - Latency histogram
- `farkle_http_response_size_bytes_bucket` - Response size distribution

**Game Metrics:**
- `farkle_game_rolls_total` - Total dice rolls
- `farkle_game_banks_total` - Total banks
- `farkle_game_farkles_total` - Total farkles
- `farkle_game_wins_total` - Total wins
- `farkle_active_games` - Current active games
- `farkle_points_distribution_bucket` - Points distribution histogram

**Trace Queries (TraceQL):**
```traceql
{resource.service.name="farkle-game"}
```

**Log Queries (LogQL):**
```logql
{service="farkle-game"} | json
```

### Changing Service Name

If you changed the service name from `farkle-game` to something else:

1. **Find and Replace in JSON:**
   ```bash
   # macOS/Linux
   sed -i 's/farkle-game/your-service-name/g' farkle-complete-dashboard.json

   # Or manually edit the JSON file
   ```

2. **Update trace/log queries:**
   - Open the dashboard JSON
   - Find all instances of `"farkle-game"`
   - Replace with your service name

---

## Dashboard Variables (Advanced)

To make dashboards more dynamic, you can add template variables:

### Example: Add Service Name Variable

1. **In Dashboard Settings** → **Variables** → **Add Variable**:
   - **Name:** `service`
   - **Type:** Constant
   - **Value:** `farkle-game`

2. **Update Panel Queries:**
   ```promql
   # Before:
   {resource.service.name="farkle-game"}

   # After:
   {resource.service.name="$service"}
   ```

### Example: Add Environment Variable

```promql
# Add variable for environment (production, staging, dev)
farkle_http_requests_total{environment="$environment"}
```

---

## Troubleshooting

### "No data" in panels

**Cause:** Data source not configured or no metrics being sent

**Fix:**
1. Verify Alloy/Grafana Agent is running: `ps aux | grep alloy`
2. Check metrics endpoint: `curl http://localhost:8080/metrics`
3. Verify data source in Grafana matches your setup
4. Check time range (default is "Last 15 minutes")

### "Data source not found"

**Cause:** Hardcoded data source UID in dashboard

**Fix:**
1. Edit the dashboard JSON
2. Find all `"datasource": {"uid": "..."}` entries
3. Remove `"uid"` field, keep only `"type"`
4. Example fix:
   ```json
   // Before:
   "datasource": {
     "type": "prometheus",
     "uid": "some-specific-uid"
   }

   // After:
   "datasource": {
     "type": "prometheus"
   }
   ```

### Traces not appearing (Complete Dashboard)

**Cause:** Tempo not configured or traces not being exported

**Fix:**
1. Verify Farkle server is sending traces:
   ```bash
   curl http://localhost:12345/metrics | grep otelcol_receiver_accepted_spans_total
   ```
2. Check Alloy logs: `tail -f logs/alloy.log | grep -i tempo`
3. Verify `JAEGER_ENDPOINT=localhost:4318` is set when starting server
4. Wait 1-2 minutes for trace indexing in Grafana Cloud

### Logs not appearing (Complete Dashboard)

**Cause:** Loki not configured or logs not being collected

**Fix:**
1. Verify log file exists: `ls -la /tmp/farkle-app.log`
2. Check Alloy is reading logs:
   ```bash
   curl http://localhost:12345/metrics | grep loki_source
   ```
3. Verify Loki configuration in `alloy-config.alloy`

---

## Creating Custom Dashboards

### Export Your Configuration

After customizing a dashboard in Grafana:

1. Click **Dashboard Settings** (gear icon)
2. Go to **JSON Model**
3. Copy the JSON
4. Save as `farkle-custom-dashboard.json`
5. Remove sensitive data (credentials, specific UIDs)

### Share Your Dashboard

To share with others:

1. Remove `"uid"` fields from datasource configs
2. Keep only `"type"` fields (e.g., `"type": "prometheus"`)
3. Add comments documenting any custom queries
4. Include setup instructions

---

## Dashboard Features

### Simple Dashboard

**System Health:**
- Total Requests
- Request Rate (req/s)
- P95 Latency

**Game Analytics:**
- Game Event Rates (rolls/banks/farkles)
- Farkle Rate %
- Total Rolls
- Total Wins

### Complete Dashboard

All Simple Dashboard features, plus:

**HTTP Performance:**
- Request Rate by Endpoint
- Response Time Percentiles (p50, p95, p99)
- HTTP Status Codes (pie chart)
- Response Size Distribution

**Game Analytics (Extended):**
- Active Games
- Win Rate
- Points Distribution Heatmaps (rolls and banks)

**Distributed Tracing:**
- Recent Traces (table view)
- Trace Latency by Endpoint
- Trace-to-log correlation

**Application Logs:**
- Log Volume by Level
- Error Log Counter
- Logs per Second
- Recent Application Logs (scrollable view)
- Game Event Logs (filtered)

---

## Best Practices

1. **Start Simple:** Use `farkle-simple-dashboard.json` first to verify data flow
2. **Add Complexity:** Once metrics work, add traces and logs with complete dashboard
3. **Customize:** Modify queries and panels to match your specific needs
4. **Time Ranges:** Adjust default time range based on your traffic patterns
5. **Alerts:** Add alert rules to panels for proactive monitoring
6. **Refresh Rate:** Default is 10s, increase for lower cardinality data

---

## Support

If you encounter issues:

1. Check the main `README.md` for observability setup
2. Review `docs/GRAFANA_CLOUD.md` for Grafana Cloud specific help
3. Verify your data sources are configured correctly
4. Check Alloy/Agent logs for export errors

---

## Contributing

Found a way to improve the dashboards? Submit a PR with:
- Clear description of the improvement
- Screenshots showing before/after
- Updated JSON file
- Documentation updates

---

## License

MIT - Same as the main Farkle project
