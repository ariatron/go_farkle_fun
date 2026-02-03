# Go-Farkle

A lightweight, web-based implementation of the classic dice game **Farkle**, built with a **Go backend** and a **vanilla JavaScript/CSS frontend**.


## Features

- **Complete Scoring Engine**
  - Standard **1s** and **5s**
  - **Three-of-a-Kind**
  - **Four-of-a-Kind Bonus** (+1000 points house rule)
  - **Straights (1–6)**
  - **Three Pairs**
- **House Rules**
  - **Minimum First Bank**: Must score at least 500 points before first bank
  - **4-of-a-Kind Bonus**: Earn +1000 bonus points on top of normal scoring
- **Hot Dice Logic**
  - Score with all 6 dice to get a fresh set and keep your turn alive
- **Interactive UI**
  - Click dice to select which ones to keep before rolling again
  - Set your player name for a personalized game
- **Win Condition**
  - First player to bank **10,000 points** wins
- **RESTful API**
  - Clean separation between game logic and the web interface
- **Comprehensive Test Coverage**
  - 22+ unit tests covering game logic, scoring, house rules, and edge cases

### Prerequisites
- **Go 1.18+** (recommended)

### Installation & Setup

Initialize the module (if not already initialized):

```bash
go mod init farkle-app
go mod tidy
```

Run the server:
```bash
go run cmd/server/main.go
```

Open your browser to `http://localhost:8080`

## API Documentation

### Endpoints

#### Roll Dice
```
POST /api/roll
Content-Type: application/json

{
  "dice_to_keep": [1, 5, 5]
}
```
Processes the kept dice from the previous roll and generates a new roll.

#### Bank Points
```
GET /api/bank
```
Banks the current turn's accumulated points, resets the turn, and checks for win condition.

#### Set Player Name
```
POST /api/set-player-name
Content-Type: application/json

{
  "player_name": "Alice"
}
```
Sets the player's name for the current game session.

#### Reset Game
```
GET /api/reset
```
Resets all game state for a fresh start.

### Response Structure
All endpoints return the current game state:
```json
{
  "player_name": "Alice",
  "turn": {
    "accumulated_score": 500,
    "dice_remaining": 3,
    "is_game_over": false
  },
  "last_roll": [1, 2, 3, 4, 5, 6],
  "total_bank": 1500,
  "winner": false,
  "history": [500, 1000]
}
```

## Testing

Run all tests:
```bash
go test ./internal/game -v
```

Expected output:
```bash
=== RUN   TestHotDiceLogic
--- PASS: TestHotDiceLogic (0.00s)
=== RUN   TestRoll
--- PASS: TestRoll (0.00s)
=== RUN   TestProcessRollSuccessfulScore
--- PASS: TestProcessRollSuccessfulScore (0.00s)
=== RUN   TestProcessRollFarkle
--- PASS: TestProcessRollFarkle (0.00s)
=== RUN   TestProcessRollHotDice
--- PASS: TestProcessRollHotDice (0.00s)
=== RUN   TestProcessRollPartialScore
--- PASS: TestProcessRollPartialScore (0.00s)
=== RUN   TestProcessRollAccumulatedScore
--- PASS: TestProcessRollAccumulatedScore (0.00s)
=== RUN   TestHotDiceReset
--- PASS: TestHotDiceReset (0.00s)
=== RUN   TestNewTurnInitialization
--- PASS: TestNewTurnInitialization (0.00s)
=== RUN   TestCalculateScore
=== RUN   TestCalculateScore/Single_1_and_5
=== RUN   TestCalculateScore/Three_of_a_Kind_(2s)
=== RUN   TestCalculateScore/Three_of_a_Kind_(1s)
=== RUN   TestCalculateScore/Straight_(1-6)
=== RUN   TestCalculateScore/Three_Pairs
=== RUN   TestCalculateScore/Four_of_a_Kind_(as_Two_Pairs)
=== RUN   TestCalculateScore/Farkle_Roll
=== RUN   TestCalculateScore/Mixed_Scoring
=== RUN   TestCalculateScore/Four_of_a_Kind_(1s)
=== RUN   TestCalculateScore/Four_of_a_Kind_(3s)
=== RUN   TestCalculateScore/Four_of_a_Kind_(5s)
=== RUN   TestCalculateScore/Five_of_a_Kind_(2s)
=== RUN   TestCalculateScore/Four_6s_with_extras
--- PASS: TestCalculateScore (0.00s)
    --- PASS: TestCalculateScore/Single_1_and_5 (0.00s)
    --- PASS: TestCalculateScore/Three_of_a_Kind_(2s) (0.00s)
    --- PASS: TestCalculateScore/Three_of_a_Kind_(1s) (0.00s)
    --- PASS: TestCalculateScore/Straight_(1-6) (0.00s)
    --- PASS: TestCalculateScore/Three_Pairs (0.00s)
    --- PASS: TestCalculateScore/Four_of_a_Kind_(as_Two_Pairs) (0.00s)
    --- PASS: TestCalculateScore/Farkle_Roll (0.00s)
    --- PASS: TestCalculateScore/Mixed_Scoring (0.00s)
    --- PASS: TestCalculateScore/Four_of_a_Kind_(1s) (0.00s)
    --- PASS: TestCalculateScore/Four_of_a_Kind_(3s) (0.00s)
    --- PASS: TestCalculateScore/Four_of_a_Kind_(5s) (0.00s)
    --- PASS: TestCalculateScore/Five_of_a_Kind_(2s) (0.00s)
    --- PASS: TestCalculateScore/Four_6s_with_extras (0.00s)
PASS
ok      farkle-app/internal/game        0.365s
```

### Test Coverage

The test suite covers:
- **Dice Generation**: Validates proper dice generation and range validation
- **Scoring Logic**: Tests all scoring scenarios (1s, 5s, three-of-a-kind, straights, three pairs)
- **House Rules**: Four-of-a-kind bonus scoring (1s, 3s, 5s, with extras)
- **Farkle Detection**: Ensures farkle (no scoring dice) is properly detected
- **Hot Dice Mechanics**: Tests the reset-to-6-dice logic when all dice score
- **State Management**: Verifies accumulated scores and dice remaining are tracked correctly
- **Edge Cases**: Tests multi-roll scenarios and complex game states

## Game Rules

### Scoring
- **1s**: 100 points each
- **5s**: 50 points each
- **Three of a Kind**:
  - Three 1s = 1,000 points
  - Three 2s = 200 points
  - Three 3s = 300 points
  - etc.
- **Four of a Kind** (House Rule): Base three-of-a-kind score **+ 1,000 bonus points**
  - Four 1s = 1,000 + 1,000 + 100 = 2,100 points
  - Four 3s = 300 + 1,000 = 1,300 points
  - Four 5s = 500 + 1,000 + 50 = 1,550 points
- **Straight (1-2-3-4-5-6)**: 1,500 points
- **Three Pairs**: 1,500 points

### House Rules
- **Minimum First Bank**: You must accumulate at least **500 points** before you can bank for the first time. After your first successful bank, you can bank any amount.
- **Four-of-a-Kind Bonus**: Rolling 4 or more of the same die awards a **+1,000 point bonus** on top of the normal scoring.

### Hot Dice
If all 6 dice score in a single roll, you get all 6 dice back and continue rolling! This is your chance to rack up points.

### How to Play
1. Set your name (optional)
2. Click "Roll Dice" to start
3. Click dice to select which ones to keep (they highlight in yellow)
4. Click "Roll Dice" again to roll the remaining dice
5. Keep rolling until you decide to bank your points or you Farkle
6. If you Farkle (no scoring dice), you lose all accumulated points for that turn
7. Click "Bank Points" to save your turn's accumulated score (remember: 500 point minimum for first bank!)
8. First to 10,000 points wins!

## Architecture

```
├── cmd/
│   └── server/
│       └── main.go              # Server initialization
├── internal/
│   ├── api/
│   │   └── handlers.go          # HTTP handlers & game state
│   └── game/
│       ├── engine.go            # Turn and roll logic
│       ├── engine_test.go        # Turn logic tests
│       ├── scoring.go            # Scoring calculations
│       └── scoring_test.go       # Scoring tests
└── static/
    └── index.html               # UI (HTML, CSS, JS)
```

### Design Patterns

- **Separation of Concerns**: Game logic (`internal/game`) is isolated from API handlers (`internal/api`)
- **Thread-Safe State**: Uses `sync.Mutex` to protect concurrent access to game state
- **RESTful API**: Clean HTTP endpoints for all game operations
- **JSON Serialization**: Automatic struct tagging for API responses

## Performance

- Lightweight and fast - no database or external services
- In-memory game state - instant feedback
- Minimal dependencies - only Go standard library

## Observability

The application includes comprehensive monitoring, tracing, and load testing:

### Features

- **📊 Structured Logging**: JSON logs with trace correlation using Go's built-in `log/slog`
- **📈 Prometheus Metrics**: HTTP and game-specific metrics exposed at `/metrics`
- **🔍 OpenTelemetry Tracing**: Distributed traces exported to Jaeger
- **⚡ K6 Load Tests**: 5 test scenarios (smoke, load, stress, spike, game scenario)

### Quick Start

```bash
# Start the application
go run cmd/server/main.go

# Access endpoints
# Game: http://localhost:8080
# Metrics: http://localhost:8080/metrics
# Health: http://localhost:8080/health
```

### Start Observability Stack

```bash
# Start Jaeger, Prometheus, and Grafana
docker-compose -f docker-compose.observability.yml up -d

# Access UIs
# Jaeger: http://localhost:16686
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3000 (admin/admin)
```

### Run Load Tests

```bash
# Install K6: https://k6.io/docs/getting-started/installation/

# Run tests locally (console output only)
k6 run tests/k6/smoke-test.js
k6 run tests/k6/load-test.js
k6 run tests/k6/stress-test.js
k6 run tests/k6/spike-test.js
k6 run tests/k6/game-scenario.js

# Run tests with Grafana Cloud reporting (sends metrics to Grafana Cloud in real-time!)
./scripts/run-k6-with-cloud.sh                           # Default: load test
./scripts/run-k6-with-cloud.sh tests/k6/smoke-test.js   # Specific test
./scripts/run-k6-with-cloud.sh tests/k6/stress-test.js  # Stress test
```

**K6 + Grafana Cloud Integration:**
- K6 tests can send metrics directly to Grafana Cloud using Prometheus remote write
- View test results in real-time as the test runs
- Import the K6 dashboard: `grafana-dashboards/k6-dashboard.json`
- See detailed guide: [`tests/k6/README.md`](tests/k6/README.md)
- Correlate K6 load with application metrics, traces, and logs

### Metrics Available

**HTTP Metrics:**
- Request count, duration, response size by endpoint

**Game Metrics:**
- Total rolls, banks, farkles, wins
- Active games count
- Points distribution

**Example Prometheus Queries:**
```promql
# Request rate
rate(farkle_http_requests_total[1m])

# P95 latency
histogram_quantile(0.95, rate(farkle_http_request_duration_seconds_bucket[5m]))

# Farkle rate
rate(farkle_game_farkles_total[5m]) / rate(farkle_game_rolls_total[5m])
```

### Grafana Cloud Integration ☁️

Send metrics, traces, and logs to **Grafana Cloud** for managed observability using **Grafana Alloy** (the successor to Grafana Agent):

```bash
# Install Grafana Alloy (macOS)
brew install grafana/grafana/alloy

# Configuration is already set up in alloy-config.alloy!

# Start Grafana Alloy
alloy run alloy-config.alloy

# In a new terminal, start the Farkle server
export JAEGER_ENDPOINT="localhost:4318"  # Send traces to Alloy
go run cmd/server/main.go

# View Alloy UI
open http://localhost:12345
```

**Benefits:**
- ✅ Managed infrastructure (no local Prometheus/Jaeger/Grafana)
- ✅ Long-term data retention
- ✅ Built-in alerting and notifications
- ✅ Free tier available (10k series, 50GB traces/logs)
- ✅ Pre-built dashboard included

### 🎭 Complete Observability Demo

**One-command demo** that showcases metrics, traces, and logs:

```bash
./scripts/demo-full-observability.sh
```

This automated demo:
1. Starts Grafana Alloy (metrics + traces + logs collector)
2. Starts the Farkle game server
3. Runs a 10-minute K6 load test
4. Generates rich observability data

**Import the complete dashboard:**
- File: `grafana-dashboards/farkle-complete-dashboard.json`
- Includes: HTTP metrics, game analytics, traces, and logs
- See: [`DASHBOARD_DEMO_GUIDE.md`](DASHBOARD_DEMO_GUIDE.md) for full demo script

**What you'll see:**
- 📊 Real-time metrics (request rates, latency, errors)
- 🎮 Game-specific analytics (rolls, farkles, wins)
- 🔍 Distributed traces showing request flows
- 📝 Structured logs with trace correlation
- ⚡ K6 load test results

### Documentation

For complete observability documentation, see:
- **[Local Observability](docs/OBSERVABILITY.md)** - Run Jaeger/Prometheus/Grafana locally
- **[Grafana Cloud Setup](docs/GRAFANA_CLOUD_SETUP.md)** - Cloud-based managed observability
- **[K6 Metrics Guide](docs/K6_METRICS_GUIDE.md)** - K6 load testing with Grafana Cloud
- **[K6 Tests README](tests/k6/README.md)** - K6 test scenarios and usage

## Future Enhancements

- Multiplayer support with player turns
- Database persistence for game history
- Difficulty levels or variants
- Leaderboard system
- Mobile app version

## License

MIT