# 🎲 Complete Observability Dashboard Demo Guide

This guide will help you demonstrate the full power of Grafana Cloud's observability platform using your Farkle game.

## 🎯 What You'll Demonstrate

This demo showcases:
- **📊 Metrics** - Real-time Prometheus metrics from the game
- **🔍 Traces** - Distributed tracing with Tempo showing request flows
- **📝 Logs** - Structured logs in Loki with correlation
- **⚡ Load Testing** - K6 test results integrated with observability
- **🎨 Custom Dashboard** - Comprehensive dashboard showing all pillars

## 🚀 Quick Start (Automated)

### Run the Complete Demo

```bash
cd /Users/ariatron/farkle-fun
./scripts/demo-full-observability.sh
```

This script will:
1. ✅ Start Grafana Alloy
2. ✅ Start Farkle game server
3. ✅ Verify all endpoints
4. ✅ Run 10-minute K6 load test
5. ✅ Generate rich observability data

## 📊 Import the Dashboard

### Step 1: Access Grafana Cloud
Go to your Grafana Cloud instance: `https://your-stack.grafana.net`

### Step 2: Import Dashboard
1. Click **Dashboards** in the left sidebar
2. Click **New** → **Import**
3. Click **Upload JSON file**
4. Select: `grafana-dashboards/farkle-complete-dashboard.json`
5. Click **Import**

### Step 3: Configure Data Sources
If prompted, map data sources:
- **Prometheus**: Your Prometheus data source
- **Tempo**: Your Tempo data source
- **Loki**: Your Loki data source

## 🎨 Dashboard Sections

The dashboard is organized into sections:

### 1. 📊 System Overview
**What it shows:**
- Total requests processed
- Current request rate (req/s)
- P95 latency
- Error rate percentage
- Game statistics (rolls, banks, farkles, wins)

**Demo talking points:**
- "Here we can see real-time system health at a glance"
- "Notice the game-specific metrics alongside traditional HTTP metrics"
- "Error rates are tracked and alerted on automatically"

### 2. 🌐 HTTP Performance
**What it shows:**
- Request rate by endpoint over time
- Response time percentiles (p50, p95, p99)
- HTTP status code distribution
- Response size distribution

**Demo talking points:**
- "We can see which endpoints are most active"
- "The `/api/roll` endpoint is our busiest"
- "P95 latency stays under 100ms even under load"
- "All responses are 200 OK - system is healthy"

### 3. 🎮 Game Analytics
**What it shows:**
- Game event rates (rolls, banks, farkles, wins)
- Farkle rate percentage gauge
- Win statistics
- Active games counter
- Points distribution heatmaps

**Demo talking points:**
- "These are custom business metrics specific to our game"
- "Farkle rate shows game balance - around 15-20% is normal"
- "We can see points distribution patterns in the heatmaps"
- "This helps us understand player behavior"

### 4. 🔍 Distributed Tracing
**What it shows:**
- Recent traces from Tempo
- Trace latency by endpoint
- Request flow visualization

**Demo talking points:**
- "Click on any trace to see the full request flow"
- "Each span shows timing for different operations"
- "We can trace requests across services (future: microservices)"
- "Traces are automatically correlated with logs via trace ID"

### 5. 📝 Application Logs
**What it shows:**
- Log volume by level (INFO, WARN, ERROR)
- Error log counter
- Logs per second
- Recent application logs
- Game event logs (rolls, banks, wins)

**Demo talking points:**
- "Structured JSON logs make them easy to query"
- "We can filter by log level to find errors quickly"
- "Game events are logged with context for debugging"
- "Trace IDs in logs link directly to corresponding traces"

## 🎭 Demo Script

Here's a suggested flow for your demo:

### Introduction (2 minutes)
```bash
# Show the game running
open http://localhost:8080

# Show Alloy UI
open http://localhost:12345
```

**Say:**
"This is a simple dice game called Farkle. But behind it, we have a complete observability stack using Grafana Cloud."

### Show Metrics (3 minutes)
```bash
# Go to Grafana Cloud
# Navigate to your dashboard
```

**Say:**
"Let me show you the dashboard. At the top, we have system overview - you can see request rates, latency, and error rates in real-time."

**Scroll to HTTP Performance:**
"Here's detailed HTTP performance. We track request rates by endpoint, response times at different percentiles, and status codes."

### Show Game Metrics (2 minutes)
**Scroll to Game Analytics:**
"These are custom business metrics. We're tracking game-specific events like dice rolls, banks, and farkles. The farkle rate gauge shows us game balance - if this gets too high, the game is too difficult."

### Show Traces (3 minutes)
```bash
# Scroll to Distributed Tracing section
```

**Say:**
"Now let's look at distributed tracing. Click on any recent trace..."

**Click a trace:**
"This shows the complete request flow. You can see how long each operation took. This is crucial for finding performance bottlenecks."

### Show Logs (2 minutes)
**Scroll to Application Logs:**
"Here are our application logs, stored in Loki. Notice they're structured JSON, so we can query specific fields."

**Click on a log with a trace_id:**
"See this trace_id field? I can click it to jump directly to the corresponding trace. This correlation between logs and traces is incredibly powerful for debugging."

### Run Load Test (3 minutes)
```bash
# In terminal
k6 run tests/k6/load-test.js
```

**Say:**
"Let me run a load test to show how the system behaves under stress. Watch the dashboard update in real-time as requests come in."

**Watch the dashboard:**
"See the request rate spike? And look - latency stays stable. Error rate remains at zero. The system is handling the load well."

### Show Correlation (2 minutes)
**Pick a specific time range with interesting data:**

**Say:**
"Let me show you something powerful. Let's say I notice high latency at this time..."

**Hover over spike in latency graph:**
"I can see exactly when it happened. Now let me check the logs for that same timeframe..."

**Adjust time range in logs panel:**
"Here are the logs from that period. And I can see the traces too. This correlation makes debugging incredibly fast."

## 🎓 Advanced Demo Tips

### Show Alerting
1. Go to **Alerting** → **Alert Rules**
2. Show how you'd create an alert:
   ```promql
   (rate(farkle_http_requests_total{status=~"5.."}[5m]) / rate(farkle_http_requests_total[5m])) > 0.05
   ```
3. Explain: "This would alert if error rate exceeds 5%"

### Show Query Building
1. Go to **Explore**
2. Show building a PromQL query step by step
3. Demonstrate log queries with LogQL

### Show Different K6 Tests
```bash
# Stress test to show system limits
k6 run tests/k6/stress-test.js

# Spike test to show recovery
k6 run tests/k6/spike-test.js
```

## 📈 Example Queries to Show

### Metrics (Prometheus)
```promql
# Request rate
rate(farkle_http_requests_total[5m])

# P95 latency by endpoint
histogram_quantile(0.95, rate(farkle_http_request_duration_seconds_bucket[5m]) by (le, endpoint))

# Farkle rate
(rate(farkle_game_farkles_total[5m]) / rate(farkle_game_rolls_total[5m])) * 100

# Top endpoints by request count
topk(5, sum by (endpoint) (farkle_http_requests_total))

# Error rate
(sum(rate(farkle_http_requests_total{status=~"5.."}[5m])) / sum(rate(farkle_http_requests_total[5m]))) * 100
```

### Traces (Tempo/TraceQL)
```traceql
# All traces for the service
{resource.service.name="farkle-game"}

# Slow traces (over 100ms)
{resource.service.name="farkle-game" && duration > 100ms}

# Traces with errors
{resource.service.name="farkle-game" && status = error}

# Traces for specific endpoint
{resource.service.name="farkle-game" && name =~ "POST /api/roll"}
```

### Logs (Loki/LogQL)
```logql
# All logs from service
{service="farkle-game"} | json

# Error logs only
{service="farkle-game", level="ERROR"} | json

# Game events
{service="farkle-game"} |~ "rolled|banked|farkled|won" | json

# Logs for specific trace
{service="farkle-game"} | json | trace_id="abc123..."

# Logs with high latency
{service="farkle-game"} | json | duration_ms > 100

# Rate of errors over time
sum(rate({service="farkle-game", level="ERROR"} [1m]))
```

## 🎪 Impressive Demonstrations

### 1. Trace to Logs Correlation
1. Find a trace in the Traces panel
2. Copy the trace ID
3. Query logs: `{service="farkle-game"} | json | trace_id="<paste-id>"`
4. Show all logs for that specific request

### 2. Incident Investigation
1. Create a scenario: "User reports slow game"
2. Check latency metrics - spot the spike
3. Look at traces during that time - find slow operations
4. Check logs - see what was happening
5. Demonstrate root cause analysis

### 3. Performance Optimization
1. Show baseline metrics
2. Make a "code change" (restart with different settings)
3. Run K6 test
4. Compare before/after metrics
5. Show improvement

### 4. Real-time Monitoring
1. Have dashboard on screen
2. Open game in browser
3. Play manually while watching dashboard update
4. Show metrics/traces/logs appear in real-time

## 🛠️ Troubleshooting During Demo

### Dashboard not showing data?
```bash
# Check Alloy status
./scripts/check-alloy.sh

# Check recent logs
tail -20 /tmp/alloy.log | grep -i error
```

### No traces appearing?
```bash
# Verify JAEGER_ENDPOINT
echo $JAEGER_ENDPOINT  # Should be localhost:4318

# Generate a test trace
curl -X POST http://localhost:8080/api/roll \
  -H "Content-Type: application/json" \
  -d '{"dice_to_keep":[]}'
```

### No logs in Loki?
```bash
# Check if log file exists and has content
ls -lh /tmp/farkle-app.log
tail /tmp/farkle-app.log
```

## 📚 Resources

- **Dashboard JSON**: `grafana-dashboards/farkle-complete-dashboard.json`
- **K6 Tests**: `tests/k6/`
- **Alloy Config**: `alloy-config.alloy`
- **Full Documentation**: `docs/OBSERVABILITY.md`

## 🎬 Post-Demo

After your demo:

1. **Stop everything cleanly:**
   ```bash
   pkill -f "alloy run"
   pkill -f "go run cmd/server/main.go"
   ```

2. **Save any custom queries** you created during the demo

3. **Export dashboard** if you made changes:
   - Dashboard settings → JSON Model → Copy
   - Save to file

## 💡 Key Talking Points

1. **Complete Observability**: "This isn't just monitoring - we have metrics, traces, AND logs all correlated"

2. **Real-time**: "Everything updates in real-time - no batch processing, no delays"

3. **Custom Metrics**: "We track business-specific metrics alongside technical ones"

4. **Correlation**: "The power is in connecting metrics → traces → logs with one click"

5. **Scalability**: "This same approach works for microservices, distributed systems, any scale"

6. **Developer Experience**: "Developers can debug production issues in minutes, not hours"

## 🎉 Success!

You now have a complete observability demo that showcases:
- ✅ Real-time metrics collection and visualization
- ✅ Distributed tracing with Tempo
- ✅ Structured logging with Loki
- ✅ Load testing with K6
- ✅ Custom business metrics
- ✅ Full correlation across all pillars
- ✅ Beautiful dashboard in Grafana Cloud

**Happy demoing! 🎲📊🔍**
