# Farkle Multiplayer Implementation Plan

## Overview

Transform the single-player Farkle game into a multiplayer game supporting 2-6 players, with turn-based gameplay, game rooms, and real-time updates.

## Current Architecture (Single-Player)

```
┌─────────────────────────────────────────┐
│ Single Game State (1 player)           │
├─────────────────────────────────────────┤
│ - player_name: string                   │
│ - turn: TurnState                       │
│ - last_roll: []int                      │
│ - total_bank: int                       │
│ - winner: bool                          │
│ - history: []int                        │
└─────────────────────────────────────────┘
```

## Proposed Multiplayer Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ Game Room (1 room = 1 game instance)                       │
├─────────────────────────────────────────────────────────────┤
│ - room_id: string (UUID)                                    │
│ - players: []Player (2-6 players)                           │
│ - current_player_index: int                                 │
│ - game_status: enum (waiting, in_progress, finished)        │
│ - turn_state: TurnState (for current player)                │
│ - last_roll: []int (for current player)                     │
│ - created_at: timestamp                                     │
│ - updated_at: timestamp                                     │
└─────────────────────────────────────────────────────────────┘
         │
         ├──> Player 1 { id, name, total_bank, turn_count, winner }
         ├──> Player 2 { id, name, total_bank, turn_count, winner }
         ├──> Player 3 { id, name, total_bank, turn_count, winner }
         └──> ...
```

## Key Changes Required

### 1. Data Models

#### New: `Player` struct
```go
type Player struct {
    ID           string    `json:"id"`           // UUID
    Name         string    `json:"name"`
    TotalBank    int       `json:"total_bank"`
    TurnCount    int       `json:"turn_count"`   // Number of turns taken
    IsWinner     bool      `json:"is_winner"`
    HasFirstBank bool      `json:"has_first_bank"` // House rule tracking
    JoinedAt     time.Time `json:"joined_at"`
}
```

#### New: `GameRoom` struct
```go
type GameRoom struct {
    RoomID             string      `json:"room_id"`
    Players            []Player    `json:"players"`
    CurrentPlayerIndex int         `json:"current_player_index"`
    GameStatus         GameStatus  `json:"game_status"`
    TurnState          TurnState   `json:"turn_state"`
    LastRoll           []int       `json:"last_roll"`
    MaxPlayers         int         `json:"max_players"`
    CreatedAt          time.Time   `json:"created_at"`
    UpdatedAt          time.Time   `json:"updated_at"`
}

type GameStatus string
const (
    StatusWaiting    GameStatus = "waiting"     // Waiting for players to join
    StatusInProgress GameStatus = "in_progress" // Game is active
    StatusFinished   GameStatus = "finished"    // Game has a winner
)
```

#### Modified: `TurnState` (unchanged, but now applies to current player)
```go
type TurnState struct {
    AccumulatedScore int  `json:"accumulated_score"`
    DiceRemaining    int  `json:"dice_remaining"`
    IsGameOver       bool `json:"is_game_over"`
}
```

### 2. API Endpoints

#### New Endpoints

```
POST   /api/rooms/create               Create new game room
POST   /api/rooms/{room_id}/join       Join existing room
GET    /api/rooms/{room_id}/state      Get current game state
POST   /api/rooms/{room_id}/start      Start game (when enough players)
POST   /api/rooms/{room_id}/roll       Roll dice (current player only)
POST   /api/rooms/{room_id}/bank       Bank points (current player only)
POST   /api/rooms/{room_id}/end-turn   End turn (pass to next player)
GET    /api/rooms                       List active rooms
DELETE /api/rooms/{room_id}/leave      Leave room
GET    /api/rooms/{room_id}/events     Server-sent events for real-time updates
```

#### Modified/Removed Endpoints

```
❌ DELETE /api/reset          → Replaced by /api/rooms/create
❌ POST   /api/set-player-name → Now part of /api/rooms/{room_id}/join
❌ GET    /api/state           → Replaced by /api/rooms/{room_id}/state
```

### 3. Game Flow

#### Phase 1: Room Setup
```
1. Player 1 creates room → receives room_id
2. Players 2-6 join using room_id
3. When 2+ players joined → Player 1 can start game
4. Game status changes: waiting → in_progress
```

#### Phase 2: Turn-Based Play
```
1. Player 1's turn (current_player_index = 0)
   - Roll dice
   - Bank points OR continue rolling
   - End turn → current_player_index = 1

2. Player 2's turn (current_player_index = 1)
   - Roll dice
   - Bank points OR continue rolling
   - End turn → current_player_index = 2

3. Continue rotation...

4. When any player reaches 10,000:
   - Mark player as winner
   - Game status → finished
```

#### Phase 3: Game End
```
1. Display winner
2. Show final scores
3. Option to create new game
```

### 4. State Management

#### Room Manager (New)
```go
type RoomManager struct {
    rooms map[string]*GameRoom
    mu    sync.RWMutex
}

func (rm *RoomManager) CreateRoom(maxPlayers int) (*GameRoom, error)
func (rm *RoomManager) GetRoom(roomID string) (*GameRoom, error)
func (rm *RoomManager) AddPlayer(roomID, playerName string) (*Player, error)
func (rm *RoomManager) RemovePlayer(roomID, playerID string) error
func (rm *RoomManager) StartGame(roomID string) error
func (rm *RoomManager) GetCurrentPlayer(roomID string) (*Player, error)
func (rm *RoomManager) NextTurn(roomID string) error
```

### 5. Validation & Rules

#### Turn Validation
```go
// Only current player can roll/bank
func (room *GameRoom) ValidatePlayerTurn(playerID string) error {
    currentPlayer := room.Players[room.CurrentPlayerIndex]
    if currentPlayer.ID != playerID {
        return errors.New("not your turn")
    }
    return nil
}
```

#### Game Start Validation
```go
// Need 2+ players to start
func (room *GameRoom) CanStart() error {
    if len(room.Players) < 2 {
        return errors.New("need at least 2 players")
    }
    if room.GameStatus != StatusWaiting {
        return errors.New("game already started")
    }
    return nil
}
```

#### Win Condition
```go
// First player to 10,000 wins
// After first player reaches 10k, everyone gets one final turn
func (room *GameRoom) CheckWinner() (*Player, bool) {
    for i, player := range room.Players {
        if player.TotalBank >= 10000 {
            // Final round logic
            if room.AllPlayersHadFinalTurn() {
                return &room.Players[i], true
            }
        }
    }
    return nil, false
}
```

### 6. Frontend Changes

#### New UI Components

**Lobby Screen**
```
┌─────────────────────────────────────┐
│ Create New Game                     │
│ [Max Players: 2-6] [Create]         │
│                                     │
│ Join Existing Game                  │
│ [Room Code: ______] [Join]          │
└─────────────────────────────────────┘
```

**Waiting Room**
```
┌─────────────────────────────────────┐
│ Room Code: ABC123                   │
│                                     │
│ Players (2/4):                      │
│ ✓ Alice                             │
│ ✓ Bob                               │
│ ⏳ Waiting...                        │
│ ⏳ Waiting...                        │
│                                     │
│ [Start Game] (if host)              │
└─────────────────────────────────────┘
```

**Game Screen**
```
┌─────────────────────────────────────────┐
│ Room: ABC123           Turn: 5          │
│                                         │
│ Players:                                │
│ 🎯 Alice:  2,450  (YOUR TURN)           │
│    Bob:    1,800                        │
│    Carol:    500                        │
│                                         │
│ Current Turn: 350 points                │
│ Dice: [⚀][⚂][⚃][⚄][⚄][⚅]              │
│                                         │
│ [Roll Again] [Bank Points] [End Turn]   │
└─────────────────────────────────────────┘
```

#### Real-Time Updates

**Option A: Polling (Simple)**
```javascript
// Poll every 2 seconds for game state updates
setInterval(() => {
  fetch(`/api/rooms/${roomId}/state`)
    .then(res => res.json())
    .then(state => updateGameUI(state))
}, 2000);
```

**Option B: Server-Sent Events (Better)**
```javascript
// Real-time updates via SSE
const eventSource = new EventSource(`/api/rooms/${roomId}/events`);
eventSource.onmessage = (event) => {
  const gameState = JSON.parse(event.data);
  updateGameUI(gameState);
};
```

**Option C: WebSockets (Best, more complex)**
```javascript
// Two-way real-time communication
const ws = new WebSocket(`ws://localhost:8080/ws/rooms/${roomId}`);
ws.onmessage = (event) => {
  const gameState = JSON.parse(event.data);
  updateGameUI(gameState);
};
```

### 7. Testing Strategy

#### Unit Tests (Game Logic)

```go
// internal/game/multiplayer_test.go

func TestRoomCreation(t *testing.T)
func TestPlayerJoin(t *testing.T)
func TestGameStart(t *testing.T)
func TestTurnRotation(t *testing.T)
func TestCurrentPlayerValidation(t *testing.T)
func TestMultiplayerScoring(t *testing.T)
func TestWinCondition(t *testing.T)
func TestFinalRound(t *testing.T)
func TestRoomCapacity(t *testing.T)
func TestPlayerLeaving(t *testing.T)
```

#### Integration Tests (API)

```go
// internal/api/multiplayer_handlers_test.go

func TestCreateRoomAPI(t *testing.T)
func TestJoinRoomAPI(t *testing.T)
func TestStartGameAPI(t *testing.T)
func TestRollDiceMultiplayer(t *testing.T)
func TestBankPointsMultiplayer(t *testing.T)
func TestEndTurnAPI(t *testing.T)
func TestInvalidTurnAttempt(t *testing.T)
func TestRoomNotFound(t *testing.T)
```

#### K6 Load Tests

```javascript
// tests/k6/multiplayer-load-test.js

export default function () {
  // Scenario 1: Create room
  const roomResp = http.post(`${BASE_URL}/api/rooms/create`, {
    max_players: 4
  });
  const roomId = roomResp.json().room_id;

  // Scenario 2: Multiple players join
  for (let i = 0; i < 3; i++) {
    http.post(`${BASE_URL}/api/rooms/${roomId}/join`, {
      player_name: `Player${i}`
    });
  }

  // Scenario 3: Start game
  http.post(`${BASE_URL}/api/rooms/${roomId}/start`);

  // Scenario 4: Play multiple turns
  for (let turn = 0; turn < 10; turn++) {
    playMultiplayerTurn(roomId);
  }
}
```

### 8. Observability Updates

#### New Metrics

```go
// Multiplayer-specific metrics
var (
    activeRooms = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "farkle_active_rooms",
        Help: "Number of active game rooms",
    })

    playersOnline = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "farkle_players_online",
        Help: "Total number of players in all rooms",
    })

    turnsPerGame = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name: "farkle_turns_per_game",
        Help: "Distribution of turns per completed game",
        Buckets: []float64{10, 20, 30, 50, 100, 200},
    })

    gameDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name: "farkle_game_duration_seconds",
        Help: "Duration of completed games",
        Buckets: []float64{60, 300, 600, 1800, 3600},
    })
)
```

#### New Traces

```go
// Trace game room operations
ctx, span := observability.StartSpan(ctx, "CreateRoom")
ctx, span := observability.StartSpan(ctx, "JoinRoom")
ctx, span := observability.StartSpan(ctx, "StartGame")
ctx, span := observability.StartSpan(ctx, "PlayerTurn")
```

### 9. File Structure

```
farkle-multiplayer/
├── cmd/
│   └── server/
│       └── main.go                    # Server entry point
├── internal/
│   ├── game/
│   │   ├── engine.go                  # Dice & turn logic (reused)
│   │   ├── scoring.go                 # Scoring logic (reused)
│   │   ├── room.go                    # NEW: Room management
│   │   ├── player.go                  # NEW: Player model
│   │   ├── room_manager.go            # NEW: Room state manager
│   │   ├── engine_test.go
│   │   ├── scoring_test.go
│   │   ├── room_test.go               # NEW: Room tests
│   │   └── multiplayer_test.go        # NEW: Integration tests
│   ├── api/
│   │   ├── room_handlers.go           # NEW: Room API handlers
│   │   ├── events.go                  # NEW: SSE/WebSocket handlers
│   │   └── room_handlers_test.go      # NEW: API tests
│   └── observability/
│       └── metrics.go                 # Updated with multiplayer metrics
├── static/
│   ├── index.html                     # Single-player (original)
│   └── multiplayer.html               # NEW: Multiplayer UI
├── tests/
│   └── k6/
│       ├── multiplayer-load-test.js   # NEW: Multiplayer load tests
│       └── multiplayer-scenario.js    # NEW: Realistic multiplayer gameplay
├── docs/
│   └── MULTIPLAYER_RULES.md           # NEW: Multiplayer game rules
└── README.md                          # Updated with multiplayer info
```

### 10. Implementation Phases

#### Phase 1: Core Multiplayer Logic (2-3 days)
- [ ] Create `Player` struct
- [ ] Create `GameRoom` struct
- [ ] Implement `RoomManager`
- [ ] Write unit tests for room management
- [ ] Write unit tests for turn rotation
- [ ] Write unit tests for win conditions

#### Phase 2: API Endpoints (2-3 days)
- [ ] Implement room creation endpoint
- [ ] Implement join room endpoint
- [ ] Implement start game endpoint
- [ ] Implement multiplayer roll/bank/end-turn endpoints
- [ ] Add validation for turn order
- [ ] Write API integration tests

#### Phase 3: Frontend (2-3 days)
- [ ] Create lobby screen (create/join room)
- [ ] Create waiting room screen
- [ ] Update game screen for multiplayer
- [ ] Implement player list UI
- [ ] Add turn indicators
- [ ] Implement polling for game state updates

#### Phase 4: Real-Time Updates (1-2 days)
- [ ] Implement Server-Sent Events OR WebSockets
- [ ] Add auto-refresh for game state
- [ ] Handle player disconnections
- [ ] Test concurrent updates

#### Phase 5: Testing & Observability (1-2 days)
- [ ] Write comprehensive K6 tests
- [ ] Add multiplayer metrics
- [ ] Add multiplayer traces
- [ ] Update dashboards with new metrics
- [ ] Load test with 10+ concurrent games

#### Phase 6: Polish & Documentation (1 day)
- [ ] Error handling & edge cases
- [ ] Update README
- [ ] Create MULTIPLAYER_RULES.md
- [ ] Add API documentation
- [ ] Final testing

**Total Estimated Time: 9-14 days**

### 11. Migration Strategy

#### Keep Both Versions

```
Main Branch (Single-Player)
  └── /farkle-fun              ← Original game (untouched)

New Branch: multiplayer
  └── /farkle-multiplayer      ← New multiplayer version
```

#### Or: Feature Flag Approach

```go
// Support both modes in same codebase
type GameMode string
const (
    ModeSinglePlayer GameMode = "single"
    ModeMultiplayer  GameMode = "multi"
)

// Toggle at startup
mode := os.Getenv("GAME_MODE") // "single" or "multi"
```

### 12. Key Challenges & Solutions

#### Challenge 1: State Synchronization
**Problem:** Multiple players need consistent game state
**Solution:** Use mutex locks + transactional updates in RoomManager

#### Challenge 2: Player Disconnections
**Problem:** Player leaves mid-game
**Solution:**
- Option A: Allow AI to take over
- Option B: Skip player's turn after timeout
- Option C: End game if player disconnects

#### Challenge 3: Concurrent Access
**Problem:** Race conditions when multiple players act simultaneously
**Solution:** Validate turn order strictly, use mutex locks

#### Challenge 4: Real-Time Updates
**Problem:** Players need to see others' actions immediately
**Solution:** Server-Sent Events (simpler) or WebSockets (better)

#### Challenge 5: Room Cleanup
**Problem:** Abandoned rooms waste memory
**Solution:** Background goroutine cleans up rooms after inactivity

```go
// Cleanup abandoned rooms every 5 minutes
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        roomManager.CleanupStaleRooms(30 * time.Minute)
    }
}()
```

### 13. Testing Scenarios

#### Scenario 1: Basic 2-Player Game
```
1. Alice creates room
2. Bob joins room
3. Alice starts game
4. Alice's turn → roll, bank
5. Bob's turn → roll, farkle
6. Alice's turn → roll, bank
7. Continue until Alice reaches 10,000
8. Verify Alice wins
```

#### Scenario 2: Player Leaves Mid-Game
```
1. Create 4-player game
2. Player 2 disconnects
3. Verify game continues
4. Player 2's turn is skipped or AI plays
```

#### Scenario 3: Final Round
```
1. Alice reaches 10,000
2. Bob, Carol get one final turn
3. Carol scores 11,000 (beats Alice)
4. Verify Carol wins (not Alice)
```

#### Scenario 4: Concurrent Room Creation
```
1. 100 users create rooms simultaneously
2. Verify all rooms get unique IDs
3. Verify no race conditions
4. Verify rooms are isolated
```

### 14. API Examples

#### Create Room
```bash
curl -X POST http://localhost:8080/api/rooms/create \
  -H "Content-Type: application/json" \
  -d '{"max_players": 4}'

# Response:
{
  "room_id": "abc123",
  "max_players": 4,
  "status": "waiting",
  "players": []
}
```

#### Join Room
```bash
curl -X POST http://localhost:8080/api/rooms/abc123/join \
  -H "Content-Type: application/json" \
  -d '{"player_name": "Alice"}'

# Response:
{
  "player_id": "uuid-1234",
  "room_id": "abc123",
  "players": [
    {"id": "uuid-1234", "name": "Alice", "total_bank": 0}
  ]
}
```

#### Get Game State
```bash
curl http://localhost:8080/api/rooms/abc123/state

# Response:
{
  "room_id": "abc123",
  "status": "in_progress",
  "current_player_index": 0,
  "players": [
    {"id": "uuid-1234", "name": "Alice", "total_bank": 2500, "is_winner": false},
    {"id": "uuid-5678", "name": "Bob", "total_bank": 1800, "is_winner": false}
  ],
  "turn_state": {
    "accumulated_score": 350,
    "dice_remaining": 3
  },
  "last_roll": [2, 4, 6]
}
```

## Summary

Creating a multiplayer version involves:
- **New models:** GameRoom, Player, RoomManager
- **New APIs:** 8 new endpoints for room management
- **Frontend changes:** Lobby, waiting room, multiplayer UI
- **Real-time updates:** SSE or WebSockets
- **Testing:** 20+ new tests covering multiplayer scenarios
- **Time estimate:** 9-14 days of focused development

The core game logic (dice rolling, scoring, house rules) can be **reused** from the single-player version. The main additions are room/player management, turn rotation, and real-time state synchronization.

**Recommendation:** Start with a separate `farkle-multiplayer` directory to keep the original intact, then decide whether to merge or keep separate.
