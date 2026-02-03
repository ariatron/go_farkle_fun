# Continuous Traffic Generation for Observability

Generate continuous background traffic to your Farkle game for realistic observability data in Grafana Cloud.

## Quick Start

### 1. Start Everything

```bash
# Terminal 1: Start Grafana Alloy (collector)
alloy run alloy-config.alloy

# Terminal 2: Start Farkle Server
export JAEGER_ENDPOINT="localhost:4318"
go run cmd/server/main.go

# Terminal 3: Start continuous traffic
./scripts/continuous-traffic.sh start
```

### 2. Monitor Traffic

```bash
# Check status
./scripts/continuous-traffic.sh status

# View live logs
./scripts/continuous-traffic.sh logs
# or
tail -f continuous-traffic.log
```

### 3. Stop Traffic

```bash
./scripts/continuous-traffic.sh stop
```

---

## Traffic Patterns

The continuous traffic generator creates **realistic game behavior**:

### Background Traffic (24/7)
- **3 concurrent users** playing constantly
- Simulates steady baseline usage
- Generates traces, metrics, and logs continuously

### Periodic Bursts
- Every 12 minutes: traffic spikes to **10 users**
- Simulates peak usage patterns
- Tests system under variable load

### Realistic Gameplay
- Random player names (Alice, Bob, Charlie, etc.)
- 3-8 turns per game
- 2-6 rolls per turn
- Strategic banking (waits for 500+ points, banks when > 1000)
- Proper handling of farkles and wins

---

## Configuration

### Adjust Traffic Levels

Edit `tests/k6/continuous-traffic.js`:

```javascript
export const options = {
  scenarios: {
    background: {
      vus: 3,          // Change number of concurrent users
      duration: '24h', // Change how long it runs
    },
    periodic_bursts: {
      stages: [
        { duration: '5m', target: 0 },
        { duration: '2m', target: 10 },  // Peak burst size
        { duration: '3m', target: 10 },
        { duration: '2m', target: 0 },
      ],
    },
  },
};
```

### Run Duration Options

**Short test (1 hour):**
```javascript
duration: '1h'
```

**Medium test (8 hours - workday):**
```javascript
duration: '8h'
```

**Long test (24 hours - full day):**
```javascript
duration: '24h'
```

**Infinite (until stopped manually):**
```javascript
duration: '999h'  // Effectively infinite
```

---

## Auto-Start on System Boot (macOS)

To automatically start traffic generation when your Mac starts:

### 1. Create logs directory
```bash
mkdir -p /Users/ariatron/farkle-fun/logs
```

### 2. Install launchd service
```bash
cp scripts/com.farkle.traffic.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.farkle.traffic.plist
```

### 3. Manage the service
```bash
# Start
launchctl start com.farkle.traffic

# Stop
launchctl stop com.farkle.traffic

# Uninstall
launchctl unload ~/Library/LaunchAgents/com.farkle.traffic.plist
rm ~/Library/LaunchAgents/com.farkle.traffic.plist
```

---

## Cron Job (Run at Specific Times)

To run traffic during specific hours (e.g., business hours):

### Edit crontab
```bash
crontab -e
```

### Add schedules
```bash
# Start traffic at 9 AM on weekdays
0 9 * * 1-5 cd /Users/ariatron/farkle-fun && ./scripts/continuous-traffic.sh start

# Stop traffic at 6 PM on weekdays
0 18 * * 1-5 cd /Users/ariatron/farkle-fun && ./scripts/continuous-traffic.sh stop

# Restart traffic every Sunday at 2 AM
0 2 * * 0 cd /Users/ariatron/farkle-fun && ./scripts/continuous-traffic.sh restart
```

---

## Monitoring in Grafana Cloud

### What You'll See

**Metrics:**
- Request rate: Steady ~5-15 req/s baseline, peaks to ~50 req/s
- Latency: P95 latency under load
- Game metrics: Rolls, banks, farkles per minute

**Traces:**
- Complete request traces for every API call
- Distributed tracing through handlers
- Performance bottleneck identification

**Logs:**
- Structured JSON logs with trace correlation
- Game events (rolls, banks, farkles, wins)
- HTTP requests with status codes and durations

### Recommended Dashboard Queries

**Request Rate:**
```promql
sum(rate(farkle_http_requests_total[1m]))
```

**Active Games:**
```promql
farkle_active_games
```

**Farkle Rate:**
```promql
rate(farkle_game_farkles_total[5m]) / rate(farkle_game_rolls_total[5m])
```

**Recent Traces:**
```traceql
{resource.service.name="farkle-game"}
```

**Game Event Logs:**
```logql
{service="farkle-game"} |~ "rolled|banked|farkled|won" | json
```

---

## Troubleshooting

### Traffic not generating

**Check if K6 is running:**
```bash
./scripts/continuous-traffic.sh status
```

**Check if server is running:**
```bash
curl http://localhost:8080/health
```

**View K6 logs:**
```bash
tail -f continuous-traffic.log
```

### No data in Grafana Cloud

**Check Alloy is running:**
```bash
# Should show UI at http://localhost:12345
open http://localhost:12345
```

**Check Alloy is receiving data:**
```bash
# Look for successful remote_write metrics
curl -s http://localhost:12345/metrics | grep prometheus_remote_write
```

**Check application logs:**
```bash
tail -f /tmp/farkle-app.log
```

### High CPU/Memory usage

**Reduce concurrent users:**
```javascript
// In tests/k6/continuous-traffic.js
vus: 1,  // Reduce from 3 to 1
```

**Increase sleep intervals:**
```javascript
// In continuous-traffic.js
sleep(randomIntBetween(5, 10));  // Longer waits between actions
```

---

## Alternative: One-Time Load Tests

If you don't want continuous traffic, run periodic tests instead:

```bash
# Run every few hours
watch -n 14400 'k6 run tests/k6/game-scenario.js'

# Or use a simple loop
while true; do
  k6 run tests/k6/game-scenario.js
  sleep 3600  # Wait 1 hour
done
```

---

## Cost Considerations

**Grafana Cloud Free Tier:**
- 10,000 series (metrics)
- 50GB traces
- 50GB logs

**Estimated usage with continuous traffic:**
- Metrics: ~200-500 series
- Traces: ~1-2GB per day
- Logs: ~500MB per day

The continuous traffic generator is designed to stay well within free tier limits while providing rich observability data.
