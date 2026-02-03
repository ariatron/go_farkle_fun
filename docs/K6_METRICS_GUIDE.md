# K6 Metrics in Grafana Cloud

This guide shows you how to visualize K6 load test metrics in Grafana Cloud.

## Quick Start

After running K6 tests with Grafana Cloud reporting:

```bash
./scripts/run-k6-with-cloud.sh tests/k6/load-test.js
```

Metrics are available immediately in your Grafana Cloud Prometheus data source.

## Key K6 Metrics

### HTTP Performance Metrics

| Metric | Description | Example Query |
|--------|-------------|---------------|
| `k6_http_req_duration` | HTTP request duration histogram | `histogram_quantile(0.95, rate(k6_http_req_duration_bucket[5m]))` |
| `k6_http_reqs_total` | Total HTTP requests counter | `rate(k6_http_reqs_total[1m])` |
| `k6_http_req_failed_total` | Failed HTTP requests | `rate(k6_http_req_failed_total[1m])` |
| `k6_data_sent` | Data sent in bytes | `rate(k6_data_sent[1m])` |
| `k6_data_received` | Data received in bytes | `rate(k6_data_received[1m])` |

### Load Testing Metrics

| Metric | Description | Example Query |
|--------|-------------|---------------|
| `k6_vus` | Active virtual users | `k6_vus` |
| `k6_vus_max` | Maximum VUs | `k6_vus_max` |
| `k6_iterations_total` | Completed iterations | `rate(k6_iterations_total[1m])` |
| `k6_iteration_duration` | Iteration duration histogram | `histogram_quantile(0.95, rate(k6_iteration_duration_bucket[5m]))` |

### Check Metrics

| Metric | Description | Example Query |
|--------|-------------|---------------|
| `k6_checks_total` | Total checks (pass + fail) | `rate(k6_checks_total[1m])` |
| `k6_check_passes` | Passed checks | `sum(rate(k6_checks_total{outcome="pass"}[1m]))` |
| `k6_check_fails` | Failed checks | `sum(rate(k6_checks_total{outcome="fail"}[1m]))` |

## Example Queries for Grafana Dashboards

### 1. Response Time Percentiles

```promql
# P50, P95, P99 response times
histogram_quantile(0.50, rate(k6_http_req_duration_bucket[5m])) or
histogram_quantile(0.95, rate(k6_http_req_duration_bucket[5m])) or
histogram_quantile(0.99, rate(k6_http_req_duration_bucket[5m]))
```

**Panel type:** Time series
**Legend:** `p{{le}}`

### 2. Request Rate

```promql
# Requests per second
sum(rate(k6_http_reqs_total[1m]))
```

**Panel type:** Time series
**Unit:** requests/sec

### 3. Error Rate Percentage

```promql
# Error rate as percentage
(
  sum(rate(k6_http_req_failed_total[1m]))
  /
  sum(rate(k6_http_reqs_total[1m]))
) * 100
```

**Panel type:** Gauge
**Unit:** percent
**Thresholds:**
- Green: 0-1%
- Yellow: 1-5%
- Red: >5%

### 4. Active Virtual Users

```promql
# Current VUs
k6_vus
```

**Panel type:** Time series
**Legend:** Active VUs

### 5. Check Success Rate

```promql
# Percentage of passed checks
(
  sum(rate(k6_checks_total{outcome="pass"}[1m]))
  /
  sum(rate(k6_checks_total[1m]))
) * 100
```

**Panel type:** Gauge
**Unit:** percent
**Thresholds:**
- Red: <95%
- Yellow: 95-99%
- Green: >99%

### 6. Throughput (Data Transfer)

```promql
# Data sent per second
sum(rate(k6_data_sent[1m]))

# Data received per second
sum(rate(k6_data_received[1m]))
```

**Panel type:** Time series
**Unit:** bytes/sec

### 7. Request Distribution by Status Code

```promql
# Requests by HTTP status
sum by (status) (
  rate(k6_http_reqs_total[1m])
)
```

**Panel type:** Pie chart
**Legend:** Status {{status}}

### 8. Iteration Duration

```promql
# P95 iteration duration
histogram_quantile(0.95,
  rate(k6_iteration_duration_bucket[5m])
)
```

**Panel type:** Time series
**Legend:** P95 Iteration Duration

## Creating a K6 Dashboard

### Option 1: Import Pre-built Dashboard

1. Go to **Dashboards** → **Import**
2. Use dashboard JSON from `grafana-dashboards/k6-dashboard.json`
3. Select your Prometheus data source
4. Click **Import**

### Option 2: Build Your Own

1. **Create new dashboard**
   - Click **+** → **Dashboard**
   - Click **Add visualization**

2. **Add Response Time Panel**
   - Query: `histogram_quantile(0.95, rate(k6_http_req_duration_bucket[5m]))`
   - Title: "Response Time (P95)"
   - Panel type: Time series
   - Unit: milliseconds

3. **Add Request Rate Panel**
   - Query: `sum(rate(k6_http_reqs_total[1m]))`
   - Title: "Request Rate"
   - Panel type: Time series
   - Unit: reqps (requests per second)

4. **Add Error Rate Panel**
   - Query: `(sum(rate(k6_http_req_failed_total[1m])) / sum(rate(k6_http_reqs_total[1m]))) * 100`
   - Title: "Error Rate"
   - Panel type: Gauge
   - Unit: percent
   - Thresholds: Green (0-1%), Yellow (1-5%), Red (>5%)

5. **Add Virtual Users Panel**
   - Query: `k6_vus`
   - Title: "Active Virtual Users"
   - Panel type: Time series

6. **Add Check Success Rate Panel**
   - Query: `(sum(rate(k6_checks_total{outcome="pass"}[1m])) / sum(rate(k6_checks_total[1m]))) * 100`
   - Title: "Check Success Rate"
   - Panel type: Gauge
   - Unit: percent
   - Thresholds: Red (<95%), Yellow (95-99%), Green (>99%)

7. **Save dashboard**
   - Click **Save** (disk icon)
   - Name: "K6 Load Testing"
   - Folder: General

## Correlating K6 Tests with Application Metrics

One of the most powerful features is correlating K6 load test metrics with your Farkle application metrics and traces.

### Example: Compare Load vs Response Time

Create a dashboard with two panels:

**Panel 1: K6 Request Rate**
```promql
sum(rate(k6_http_reqs_total[1m]))
```

**Panel 2: Farkle App Response Time**
```promql
rate(farkle_request_duration_seconds_sum[1m])
/
rate(farkle_request_duration_seconds_count[1m])
```

This shows how application performance changes under different load levels.

### Example: Error Correlation

**Panel 1: K6 Error Rate**
```promql
sum(rate(k6_http_req_failed_total[1m]))
```

**Panel 2: Farkle App Error Logs**
Switch to **Loki** data source:
```logql
{service="farkle-game"} |= "error" | json
```

This correlates load test failures with application error logs.

### Example: Trace Analysis During Load Tests

1. Run a load test: `./scripts/run-k6-with-cloud.sh tests/k6/load-test.js`
2. Go to **Grafana Cloud** → **Explore**
3. Select **Tempo** data source
4. Query traces during the load test time range:
   ```
   { service.name="farkle-game" }
   ```
5. Sort by duration to find slowest requests
6. Inspect traces to see where time is spent

## Alerting on K6 Metrics

Set up alerts for K6 test failures:

### Alert: High Error Rate

**Query:**
```promql
(sum(rate(k6_http_req_failed_total[1m])) / sum(rate(k6_http_reqs_total[1m]))) * 100 > 5
```

**Condition:** Error rate above 5% for 2 minutes

### Alert: High Response Time

**Query:**
```promql
histogram_quantile(0.95, rate(k6_http_req_duration_bucket[5m])) > 0.250
```

**Condition:** P95 response time above 250ms for 5 minutes

### Alert: Check Failures

**Query:**
```promql
(sum(rate(k6_checks_total{outcome="fail"}[1m])) / sum(rate(k6_checks_total[1m]))) * 100 > 5
```

**Condition:** Check failure rate above 5% for 2 minutes

## Real-Time Monitoring During Tests

When running K6 tests with Grafana Cloud, you can watch metrics update in real-time:

1. **Open Grafana Cloud** in your browser
2. **Navigate to Explore** (compass icon)
3. **Select Prometheus** data source
4. **Start a K6 test:**
   ```bash
   ./scripts/run-k6-with-cloud.sh tests/k6/stress-test.js
   ```
5. **Query metrics** as the test runs:
   ```promql
   k6_http_req_duration
   ```
6. **Enable auto-refresh** (top right) - set to 5s or 10s
7. **Watch the graphs update** as load increases/decreases

This is particularly useful for:
- Stress tests - see when the system starts degrading
- Spike tests - observe recovery after traffic spikes
- Endurance tests - monitor for memory leaks or gradual degradation

## Tips and Best Practices

### 1. Tag Your Tests

Add custom tags to K6 tests for easier filtering:

```javascript
export const options = {
  tags: {
    test_type: 'load_test',
    environment: 'production',
    version: 'v1.2.3',
  },
};
```

Query in Grafana:
```promql
rate(k6_http_reqs_total{test_type="load_test"}[1m])
```

### 2. Use Time Range Variables

Create dashboard variables for flexible time ranges:
- `$__rate_interval` - Auto-adjusts rate interval
- `$__range` - Current dashboard time range

### 3. Combine with Annotations

Add annotations to mark when tests run:
1. Dashboard settings → Annotations
2. Add annotation query from Loki:
   ```logql
   {service="farkle-game"} |= "K6 test started"
   ```

### 4. Export Test Results

Save K6 summary data alongside real-time metrics:

```javascript
export function handleSummary(data) {
  return {
    'summary.json': JSON.stringify(data),
    'stdout': textSummary(data),
  };
}
```

### 5. Baseline Comparison

Run baseline tests regularly and compare:
1. Run baseline: `./scripts/run-k6-with-cloud.sh tests/k6/load-test.js`
2. Note timestamp in Grafana
3. After code changes, run same test
4. Compare time ranges in Grafana Explore

## Troubleshooting

### No metrics appearing

1. **Check K6 output** - Look for Prometheus remote write errors
2. **Verify credentials** - Check `k6-grafana-cloud.json`
3. **Test manually:**
   ```bash
   export $(cat k6-grafana-cloud.json | jq -r 'to_entries | .[] | "\(.key)=\(.value)"')
   k6 run --out experimental-prometheus-rw tests/k6/smoke-test.js
   ```
4. **Check Prometheus endpoint** - Ensure URL is accessible

### Metrics delayed

- **Push interval** - K6 sends metrics every 5 seconds (configurable)
- **Grafana refresh** - Set dashboard auto-refresh to 5-10 seconds
- **Query time range** - Use `now-5m` to `now` for recent data

### High cardinality warnings

K6 generates many unique label combinations. If you see warnings:
- Reduce test duration
- Limit number of VUs
- Aggregate metrics before querying

### Certificate errors

If you see TLS/SSL errors:
```bash
# Development only!
export K6_INSECURE_SKIP_TLS_VERIFY=true
```

## Resources

- [K6 Documentation](https://k6.io/docs/)
- [K6 Metrics Reference](https://k6.io/docs/using-k6/metrics/)
- [Prometheus Remote Write](https://k6.io/docs/results-output/real-time/prometheus-remote-write/)
- [Grafana Dashboards](https://grafana.com/docs/grafana/latest/dashboards/)
- [PromQL Basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)
