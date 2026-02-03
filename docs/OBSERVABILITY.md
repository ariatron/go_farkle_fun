# Observability Guide

Complete guide to monitoring, tracing, and testing the Farkle game application.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Logs](#logs)
- [Metrics](#metrics)
- [Tracing](#tracing)
- [Load Testing](#load-testing)
- [Dashboards](#dashboards)
- [Troubleshooting](#troubleshooting)

## Overview

The Farkle game includes comprehensive observability through three pillars:

1. **Structured Logging** - JSON logs with trace correlation
2. **Prometheus Metrics** - HTTP and game-specific metrics
3. **OpenTelemetry Tracing** - Distributed traces with Jaeger

## Architecture

```
┌─────────────┐
│   Browser   │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────┐
│      Farkle Server :8080        │
│  ┌───────────────────────────┐  │
│  │  Observability Middleware │  │
│  │  - Logging                │  │
│  │  - Metrics                │  │
│  │  - Tracing                │  │
│  └───────────────────────────┘  │
└────┬────────────┬───────────┬───┘
     │            │           │
     ▼            ▼           ▼
  Logs      /metrics      Jaeger
  stdout    :8080/metrics  :4318
```

## Quick Start

### 1. Start the Application

```bash
cd /Users/ariatron/farkle-fun
go run cmd/server/main.go
```

The server will start with:
- **Application**: http://localhost:8080
- **Metrics**: http://localhost:8080/metrics
- **Health**: http://localhost:8080/health

### 2. Start Observability Stack (Optional)

```bash
docker-compose -f docker-compose.observability.yml up -d
```

This starts:
- **Jaeger UI**: http://localhost:16686
- **Prometheus UI**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

### 3. Run Load Tests

```bash
# Smoke test (verify functionality)
k6 run tests/k6/smoke-test.js

# Load test (expected traffic)
k6 run tests/k6/load-test.js

# Stress test (find limits)
k6 run tests/k6/stress-test.js

# Spike test (sudden surge)
k6 run tests/k6/spike-test.js

# Game scenario (realistic gameplay)
k6 run tests/k6/game-scenario.js
```

## Logs

### Format

All logs are structured JSON with the following fields:

```json
{
  "time": "2026-02-02T15:47:00Z",
  "level": "INFO",
  "msg": "HTTP request",
  "method": "POST",
  "path": "/api/roll",
  "status": 200,
  "duration_ms": 2.34,
  "trace_id": "abc123..."
}
```

### Log Levels

- **INFO**: Normal operations (HTTP requests, game events)
- **WARN**: Warnings (tracing unavailable, validation issues)
- **ERROR**: Errors (startup failures, panics)
- **DEBUG**: Detailed debugging (set via environment)

### Game Events

The application logs these game-specific events:

```json
// Dice rolled
{"msg": "Dice rolled", "dice": [1,2,3,4,5,6], "possible_score": 150}

// Points banked
{"msg": "Points banked", "banked_score": 500, "total_bank": 1500}

// Farkle
{"msg": "Player farkled", "accumulated_score_lost": 300}

// Win
{"msg": "Player won the game!", "final_score": 10050}

// Game reset
{"msg": "Game reset", "previous_score": 5000}
```

### Viewing Logs

```bash
# View in real-time
go run cmd/server/main.go

# Filter by level
go run cmd/server/main.go | grep '"level":"ERROR"'

# Pretty print with jq
go run cmd/server/main.go | jq .

# Filter game events
go run cmd/server/main.go | jq 'select(.msg | contains("rolled"))'
```

## Metrics

### Available Metrics

#### HTTP Metrics

**`farkle_http_requests_total`** - Counter
- Total HTTP requests
- Labels: `method`, `endpoint`, `status`

**`farkle_http_request_duration_seconds`** - Histogram
- Request latency in seconds
- Labels: `method`, `endpoint`
- Buckets: 1ms, 5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s

**`farkle_http_response_size_bytes`** - Histogram
- Response size in bytes
- Labels: `method`, `endpoint`
- Buckets: 100, 200, 500, 1000, 2000, 5000, 10000

#### Game Metrics

**`farkle_game_rolls_total`** - Counter
- Total dice rolls

**`farkle_game_banks_total`** - Counter
- Total points banked

**`farkle_game_farkles_total`** - Counter
- Total farkles

**`farkle_game_wins_total`** - Counter
- Total games won

**`farkle_active_games`** - Gauge
- Number of active games

**`farkle_points_distribution`** - Histogram
- Distribution of points scored
- Labels: `type` (roll or bank)
- Buckets: 0, 50, 100, 150, 200, 300, 500, 1000, 1500, 2000, 3000

### Querying Metrics

#### Direct Access

```bash
# View all metrics
curl http://localhost:8080/metrics

# Filter specific metric
curl http://localhost:8080/metrics | grep farkle_game_rolls_total
```

#### Prometheus Queries

Access Prometheus UI at http://localhost:9090

```promql
# Request rate per endpoint
rate(farkle_http_requests_total[1m])

# P95 latency
histogram_quantile(0.95, rate(farkle_http_request_duration_seconds_bucket[5m]))

# Farkle rate
rate(farkle_game_farkles_total[5m]) / rate(farkle_game_rolls_total[5m])

# Total games won
farkle_game_wins_total

# Average points per bank
rate(farkle_points_distribution_sum{type="bank"}[5m]) /
rate(farkle_points_distribution_count{type="bank"}[5m])
```

## Tracing

### Configuration

Traces are exported to Jaeger via OTLP HTTP on port 4318.

```bash
# Default endpoint
JAEGER_ENDPOINT=localhost:4318 go run cmd/server/main.go

# Custom endpoint
JAEGER_ENDPOINT=my-jaeger:4318 go run cmd/server/main.go
```

### Trace Structure

Each HTTP request creates a trace with spans:

```
HTTP Request (POST /api/roll)
├── RollHandler
│   ├── Process kept dice
│   ├── Roll new dice
│   └── Calculate score
└── Response
```

### Span Attributes

**HTTP Spans:**
- `http.method`: Request method
- `http.url`: Full URL
- `http.status_code`: Response status
- `http.duration_ms`: Request duration

**Game Spans:**
- `kept_dice`: Dice kept from previous roll
- `score`: Points scored
- `rolled_dice`: Newly rolled dice
- `dice_count`: Number of dice
- `banked_score`: Points banked
- `total_bank`: Total banked points

### Viewing Traces

1. Open Jaeger UI: http://localhost:16686
2. Select service: `farkle-game`
3. Click "Find Traces"
4. Click on a trace to view details

### Trace Events

Special events logged within spans:

- `farkle`: Player farkled
- `farkle_on_roll`: Farkled immediately on roll
- `game_won`: Player won the game
- `game_reset`: Game was reset

## Load Testing

### Test Suite

Five K6 test scenarios are included:

#### 1. Smoke Test (`smoke-test.js`)
- **Purpose**: Verify basic functionality
- **VUs**: 1
- **Duration**: 30 seconds
- **Thresholds**: p95 < 200ms, 99% < 500ms

```bash
k6 run tests/k6/smoke-test.js
```

#### 2. Load Test (`load-test.js`)
- **Purpose**: Expected production load
- **VUs**: 10
- **Duration**: 5 minutes
- **Thresholds**: p95 < 250ms, p99 < 750ms

```bash
k6 run tests/k6/load-test.js
```

#### 3. Stress Test (`stress-test.js`)
- **Purpose**: Find breaking point
- **VUs**: Up to 50
- **Duration**: 9 minutes
- **Thresholds**: p95 < 1000ms, < 5% errors

```bash
k6 run tests/k6/stress-test.js
```

#### 4. Spike Test (`spike-test.js`)
- **Purpose**: Sudden traffic surge
- **VUs**: 2 → 50 → 2
- **Duration**: 5 minutes
- **Thresholds**: p95 < 1500ms, < 10% errors

```bash
k6 run tests/k6/spike-test.js
```

#### 5. Game Scenario (`game-scenario.js`)
- **Purpose**: Realistic gameplay
- **VUs**: 5
- **Duration**: 10 minutes
- **Thresholds**: p95 < 300ms, > 10 games completed

```bash
k6 run tests/k6/game-scenario.js
```

### Custom Test Configuration

Override settings via environment variables:

```bash
# Custom base URL
BASE_URL=http://production:8080 k6 run tests/k6/load-test.js

# Custom VUs and duration
k6 run --vus 20 --duration 10m tests/k6/load-test.js
```

## Dashboards

### Grafana (Optional)

1. Access Grafana: http://localhost:3000 (admin/admin)
2. Go to "Dashboards" → "New" → "New Dashboard"
3. Add panels with Prometheus queries

**Example Panels:**

**Request Rate:**
```promql
sum(rate(farkle_http_requests_total[1m])) by (endpoint)
```

**Latency:**
```promql
histogram_quantile(0.95,
  sum(rate(farkle_http_request_duration_seconds_bucket[5m])) by (le, endpoint)
)
```

**Game Metrics:**
```promql
# Rolls per second
rate(farkle_game_rolls_total[1m])

# Farkle rate
rate(farkle_game_farkles_total[1m]) / rate(farkle_game_rolls_total[1m])

# Wins
farkle_game_wins_total
```

## Troubleshooting

### Logs Not Appearing

**Issue**: No logs in stdout

**Solutions:**
```bash
# Ensure logger is initialized
grep "Logger initialized" logs

# Check log level
LOG_LEVEL=debug go run cmd/server/main.go
```

### Metrics Endpoint Returns 404

**Issue**: `/metrics` returns 404

**Solutions:**
```bash
# Verify server is running
curl http://localhost:8080/health

# Check metrics endpoint
curl http://localhost:8080/metrics

# Ensure Prometheus client is imported
grep "prometheus" go.mod
```

### Traces Not Appearing in Jaeger

**Issue**: No traces in Jaeger UI

**Solutions:**
```bash
# Check if Jaeger is running
docker ps | grep jaeger

# Verify OTLP port is accessible
curl http://localhost:4318/v1/traces

# Check server logs for tracing warnings
go run cmd/server/main.go | grep tracing

# Test with manual span
# Tracing will warn but continue if Jaeger unavailable
```

### K6 Tests Failing

**Issue**: K6 tests show high error rates

**Solutions:**
```bash
# Ensure server is running
curl http://localhost:8080/health

# Check server logs during test
go run cmd/server/main.go &
k6 run tests/k6/smoke-test.js

# Reduce VUs if resource constrained
k6 run --vus 5 tests/k6/load-test.js

# Increase timeout thresholds
# Edit test file and adjust threshold values
```

### High Memory Usage

**Issue**: Application consuming too much memory

**Solutions:**
```bash
# Monitor memory
go run cmd/server/main.go &
watch -n 1 'ps aux | grep main'

# Profile memory
go tool pprof http://localhost:8080/debug/pprof/heap

# Check for goroutine leaks
curl http://localhost:8080/debug/pprof/goroutine
```

## Best Practices

### Logs
- Use structured logging with consistent fields
- Include trace IDs for correlation
- Log business events, not just HTTP requests
- Avoid logging sensitive data (passwords, tokens)

### Metrics
- Use counters for events that only increase
- Use gauges for values that go up and down
- Use histograms for distributions (latency, sizes)
- Keep label cardinality low (< 100 unique values per label)

### Tracing
- Create spans for significant operations
- Add relevant attributes to spans
- Use events for important moments within spans
- Keep span names consistent and descriptive

### Testing
- Run smoke tests after every deploy
- Run load tests before releasing
- Establish baseline metrics for comparison
- Monitor tests in production-like environments

## Additional Resources

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [K6 Documentation](https://k6.io/docs/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
