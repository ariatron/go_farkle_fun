# K6 Load Tests with Grafana Cloud

This directory contains K6 load tests for the Farkle game, with integrated Grafana Cloud reporting.

## Quick Start

### Run Tests Locally (Console Output Only)

```bash
# From project root
k6 run tests/k6/smoke-test.js
k6 run tests/k6/load-test.js
k6 run tests/k6/stress-test.js
k6 run tests/k6/spike-test.js
```

### Run Tests with Grafana Cloud Reporting

```bash
# Run with default test (load-test.js)
./scripts/run-k6-with-cloud.sh

# Run specific test
./scripts/run-k6-with-cloud.sh tests/k6/smoke-test.js
./scripts/run-k6-with-cloud.sh tests/k6/stress-test.js
./scripts/run-k6-with-cloud.sh tests/k6/spike-test.js
```

**Note:** Tests run with Grafana Cloud reporting send metrics in real-time to your Grafana Cloud Prometheus instance. You can monitor the test progress live in Grafana dashboards!

## Test Scenarios

### 1. Smoke Test (`smoke-test.js`)
- **Purpose:** Quick sanity check
- **VUs:** 1-2 users
- **Duration:** 1 minute
- **Use case:** Verify basic functionality before larger tests

### 2. Load Test (`load-test.js`)
- **Purpose:** Expected production traffic
- **VUs:** 10 users
- **Duration:** 5 minutes (1min ramp-up, 3min sustained, 1min ramp-down)
- **Use case:** Validate system under normal load

### 3. Stress Test (`stress-test.js`)
- **Purpose:** Find system limits
- **VUs:** Ramps to 50 users
- **Duration:** 10 minutes
- **Use case:** Identify breaking points and degradation

### 4. Spike Test (`spike-test.js`)
- **Purpose:** Test sudden traffic surges
- **VUs:** Spikes from 10 to 100 users
- **Duration:** 5 minutes
- **Use case:** Verify system handles sudden traffic increases

### 5. Game Scenario (`game-scenario.js`)
- **Purpose:** Realistic game play
- **VUs:** 5 users
- **Duration:** 3 minutes
- **Use case:** Simulate real player behavior

### 6. Continuous Traffic (`continuous-traffic.js`)
- **Purpose:** Long-running background traffic
- **VUs:** 3 users with periodic bursts to 10
- **Duration:** 24 hours
- **Use case:** Generate continuous data for observability demos

## Grafana Cloud Integration

### Setup

1. **Configuration file already created** - `k6-grafana-cloud.json` contains your credentials
2. **Run tests using the helper script** - Automatically loads credentials and enables reporting

### View Results in Grafana Cloud

After running tests with Grafana Cloud reporting:

1. Go to your Grafana Cloud instance
2. Navigate to **Explore**
3. Select your **Prometheus** data source
4. Query K6 metrics:

```promql
# HTTP request duration (response times)
k6_http_req_duration

# HTTP request rate
rate(k6_http_reqs_total[1m])

# Active virtual users
k6_vus

# Check pass/fail rate
rate(k6_checks_total[1m])

# HTTP request failures
rate(k6_http_req_failed_total[1m])

# Data sent/received
k6_data_sent
k6_data_received
```

### K6 Dashboard

Create a Grafana dashboard with panels for:

- **Response Time Trends:** Track p50, p95, p99 latencies
- **Request Rate:** Requests per second over time
- **Active VUs:** Number of simulated users
- **Error Rate:** Failed requests percentage
- **Throughput:** Data sent/received
- **Check Success Rate:** Percentage of passed checks

Example queries for dashboard panels:

```promql
# P95 response time by endpoint
histogram_quantile(0.95,
  rate(k6_http_req_duration_bucket[5m])
)

# Request rate by status
sum by (status) (
  rate(k6_http_reqs_total[1m])
)

# Error rate percentage
sum(rate(k6_http_req_failed_total[1m]))
/
sum(rate(k6_http_reqs_total[1m])) * 100
```

## Test Thresholds

Each test has defined SLOs (Service Level Objectives):

- **Response time:**
  - p95 < 200-250ms (varies by test)
  - p99 < 500-750ms (varies by test)
- **Error rate:** < 1%
- **Check pass rate:** > 95%

Tests will **fail** if thresholds are not met, making them suitable for CI/CD pipelines.

## Realistic Game Behavior

All tests simulate realistic player behavior:
- Rolling dice (GET `/api/roll`)
- Banking points (POST `/api/bank`)
- Resetting games (POST `/api/reset`)
- Setting player names (POST `/api/set-player-name`)

Think times between actions: 1-5 seconds (simulating human decision-making)

## Utilities (`utils.js`)

Common functions used across all tests:
- `resetGame()` - Start a new game
- `setPlayerName(name)` - Set player name
- `rollDice()` - Roll dice
- `bankPoints()` - Bank current score
- `playRealisticTurn()` - Simulate a complete turn with decisions
- `getCommonThresholds()` - Shared performance thresholds

## Troubleshooting

### No metrics in Grafana Cloud

1. **Verify credentials:** Check `k6-grafana-cloud.json` has correct values
2. **Check K6 version:** `k6 version` (need v0.47.0+ for Prometheus remote write)
3. **Test manually:**
   ```bash
   export K6_PROMETHEUS_RW_SERVER_URL="https://prometheus-..."
   export K6_PROMETHEUS_RW_USERNAME="your-user-id"
   export K6_PROMETHEUS_RW_PASSWORD="your-api-token"
   k6 run --out experimental-prometheus-rw tests/k6/smoke-test.js
   ```
4. **Check Grafana Cloud:** Verify Prometheus endpoint is accessible
5. **Review K6 output:** Look for connection errors in console

### Tests failing

- **Server not running:** Start Farkle server on port 8080
- **High response times:** Check server logs for errors
- **Connection refused:** Verify server is listening on `localhost:8080`

### Certificate errors

If you see SSL/TLS errors:
```bash
# Try disabling TLS verification (development only!)
export K6_INSECURE_SKIP_TLS_VERIFY=true
```

## For New Users (Different Grafana Cloud Instance)

If you're setting up K6 with your own Grafana Cloud:

1. Copy the example config:
   ```bash
   cp k6-grafana-cloud.json.example k6-grafana-cloud.json
   ```

2. Edit `k6-grafana-cloud.json` with your credentials:
   - `K6_PROMETHEUS_RW_SERVER_URL`: Your Prometheus push endpoint
   - `K6_PROMETHEUS_RW_USERNAME`: Your Prometheus user ID
   - `K6_PROMETHEUS_RW_PASSWORD`: Your Grafana Cloud API token

3. See `docs/GRAFANA_CLOUD_SETUP.md` for detailed setup instructions

## CI/CD Integration

Run K6 tests in your CI pipeline:

```yaml
# GitHub Actions example
- name: Run K6 Load Test
  env:
    K6_PROMETHEUS_RW_SERVER_URL: ${{ secrets.GRAFANA_PROMETHEUS_URL }}
    K6_PROMETHEUS_RW_USERNAME: ${{ secrets.GRAFANA_PROMETHEUS_USER }}
    K6_PROMETHEUS_RW_PASSWORD: ${{ secrets.GRAFANA_API_TOKEN }}
  run: |
    k6 run --out experimental-prometheus-rw tests/k6/load-test.js
```

Tests will fail the pipeline if thresholds aren't met!

## References

- [K6 Documentation](https://k6.io/docs/)
- [K6 Prometheus Remote Write](https://k6.io/docs/results-output/real-time/prometheus-remote-write/)
- [Grafana Cloud K6](https://grafana.com/docs/grafana-cloud/k6/)
- [K6 Thresholds](https://k6.io/docs/using-k6/thresholds/)
