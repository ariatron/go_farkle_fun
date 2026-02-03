# Go-Farkle

A lightweight, web-based implementation of the classic dice game **Farkle**, built with a **Go backend** and a **vanilla JavaScript/CSS frontend**.


## Features

- **Complete Scoring Engine**
  - Standard **1s** and **5s**
  - **Three-of-a-Kind** (House Rule)
  - **Straights (1–6)**
  - **Three Pairs**
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
  - 17+ unit tests covering game logic, scoring, and edge cases

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
--- PASS: TestCalculateScore (0.00s)
    --- PASS: TestCalculateScore/Single_1_and_5 (0.00s)
    --- PASS: TestCalculateScore/Three_of_a_Kind_(2s) (0.00s)
    --- PASS: TestCalculateScore/Three_of_a_Kind_(1s) (0.00s)
    --- PASS: TestCalculateScore/Straight_(1-6) (0.00s)
    --- PASS: TestCalculateScore/Three_Pairs (0.00s)
    --- PASS: TestCalculateScore/Four_of_a_Kind_(as_Two_Pairs) (0.00s)
    --- PASS: TestCalculateScore/Farkle_Roll (0.00s)
    --- PASS: TestCalculateScore/Mixed_Scoring (0.00s)
PASS
ok      farkle-app/internal/game        0.365s
```

### Test Coverage

The test suite covers:
- **Dice Generation**: Validates proper dice generation and range validation
- **Scoring Logic**: Tests all scoring scenarios (1s, 5s, three-of-a-kind, straights, three pairs)
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
- **Straight (1-2-3-4-5-6)**: 1,500 points
- **Three Pairs**: 1,500 points

### Hot Dice
If all 6 dice score in a single roll, you get all 6 dice back and continue rolling! This is your chance to rack up points.

### How to Play
1. Set your name (optional)
2. Click "Roll Dice" to start
3. Click dice to select which ones to keep (they highlight in yellow)
4. Click "Roll Dice" again to roll the remaining dice
5. Keep rolling until you decide to bank your points or you Farkle
6. If you Farkle (no scoring dice), you lose all accumulated points for that turn
7. Click "Bank Points" to save your turn's accumulated score
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

## Future Enhancements

- Multiplayer support with player turns
- Database persistence for game history
- Difficulty levels or variants
- Leaderboard system
- Mobile app version

## License

MIT