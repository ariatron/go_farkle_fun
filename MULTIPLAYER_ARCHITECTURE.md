# Farkle Multiplayer Architecture

## System Architecture Comparison

### Current: Single-Player

```mermaid
graph TB
    Browser[Browser] -->|HTTP| API[API Handlers]
    API -->|Read/Write| GameState[Game State<br/>mutex-protected]
    GameState -->|Contains| Turn[TurnState]
    GameState -->|Contains| Player[Player Data]
    GameState -->|Contains| Score[Score History]

    API -->|Uses| Engine[Game Engine]
    Engine -->|Dice Logic| Roll[Roll Dice]
    Engine -->|Scoring| Calc[Calculate Score]

    API -->|Metrics| Prom[Prometheus]
    API -->|Traces| Tempo[Tempo]
    API -->|Logs| Loki[Loki]

    style GameState fill:#90EE90
    style Browser fill:#87CEEB
```

### Proposed: Multiplayer

```mermaid
graph TB
    B1[Browser 1] -->|HTTP/SSE| API[API Handlers]
    B2[Browser 2] -->|HTTP/SSE| API
    B3[Browser 3] -->|HTTP/SSE| API

    API -->|Manages| RM[RoomManager]
    RM -->|Contains| R1[Room 1]
    RM -->|Contains| R2[Room 2]
    RM -->|Contains| R3[Room N...]

    R1 -->|Has| P1[Player 1]
    R1 -->|Has| P2[Player 2]
    R1 -->|Has| P3[Player 3]
    R1 -->|Tracks| TS1[Turn State]

    API -->|Uses| Engine[Game Engine<br/>Reused!]
    Engine -->|Dice Logic| Roll[Roll Dice]
    Engine -->|Scoring| Calc[Calculate Score]

    API -->|Broadcasts| SSE[Server-Sent Events]
    SSE -->|Updates| B1
    SSE -->|Updates| B2
    SSE -->|Updates| B3

    API -->|Metrics| Prom[Prometheus]
    API -->|Traces| Tempo[Tempo]
    API -->|Logs| Loki[Loki]

    style RM fill:#FFD700
    style R1 fill:#90EE90
    style R2 fill:#90EE90
    style R3 fill:#90EE90
    style Engine fill:#87CEEB
```

## Data Model Comparison

### Current: Single-Player State

```mermaid
classDiagram
    class GameState {
        +string PlayerName
        +TurnState Turn
        +[]int LastRoll
        +int TotalBank
        +bool Winner
        +[]int History
        +sync.Mutex mu
        +Roll()
        +Bank()
        +Reset()
    }

    class TurnState {
        +int AccumulatedScore
        +int DiceRemaining
        +bool IsGameOver
    }

    GameState "1" --> "1" TurnState
```

### Proposed: Multiplayer State

```mermaid
classDiagram
    class RoomManager {
        +map[string]*GameRoom rooms
        +sync.RWMutex mu
        +CreateRoom() *GameRoom
        +GetRoom(id) *GameRoom
        +AddPlayer(roomID, name) *Player
        +NextTurn(roomID)
        +CleanupStaleRooms()
    }

    class GameRoom {
        +string RoomID
        +[]Player Players
        +int CurrentPlayerIndex
        +GameStatus Status
        +TurnState Turn
        +[]int LastRoll
        +int MaxPlayers
        +time.Time CreatedAt
        +time.Time UpdatedAt
        +sync.Mutex mu
        +ValidatePlayerTurn(playerID)
        +CheckWinner() *Player
        +NextTurn()
    }

    class Player {
        +string ID
        +string Name
        +int TotalBank
        +int TurnCount
        +bool IsWinner
        +bool HasFirstBank
        +time.Time JoinedAt
    }

    class TurnState {
        +int AccumulatedScore
        +int DiceRemaining
        +bool IsGameOver
    }

    class GameStatus {
        <<enumeration>>
        Waiting
        InProgress
        Finished
    }

    RoomManager "1" --> "*" GameRoom
    GameRoom "1" --> "2..6" Player
    GameRoom "1" --> "1" TurnState
    GameRoom "1" --> "1" GameStatus
```

## API Flow: Single-Player vs Multiplayer

### Current: Single-Player Flow

```mermaid
sequenceDiagram
    participant B as Browser
    participant A as API
    participant G as GameState

    B->>A: POST /api/set-player-name
    A->>G: Set player name
    G-->>A: OK
    A-->>B: Game state

    B->>A: POST /api/roll
    A->>G: Roll dice
    G->>G: Generate roll
    G->>G: Update state
    G-->>A: Updated state
    A-->>B: Game state

    B->>A: GET /api/bank
    A->>G: Bank points
    G->>G: Add to total
    G->>G: Reset turn
    G-->>A: Updated state
    A-->>B: Game state
```

### Proposed: Multiplayer Flow

```mermaid
sequenceDiagram
    participant B1 as Browser 1
    participant B2 as Browser 2
    participant A as API
    participant RM as RoomManager
    participant R as GameRoom
    participant SSE as SSE Broadcaster

    B1->>A: POST /api/rooms/create
    A->>RM: Create room
    RM->>R: New room (waiting)
    R-->>RM: Room created
    RM-->>A: room_id
    A-->>B1: {room_id: "abc123"}

    B2->>A: POST /api/rooms/abc123/join
    A->>RM: Add player
    RM->>R: Add player 2
    R-->>RM: Player added
    RM-->>A: Updated room
    A->>SSE: Broadcast update
    SSE-->>B1: Player joined event
    A-->>B2: {player_id: "xyz"}

    B1->>A: POST /api/rooms/abc123/start
    A->>RM: Start game
    RM->>R: Change status to in_progress
    R-->>RM: Game started
    A->>SSE: Broadcast update
    SSE-->>B1: Game started event
    SSE-->>B2: Game started event
    A-->>B1: Game state

    B1->>A: POST /api/rooms/abc123/roll
    A->>R: Validate turn (Player 1)
    R->>R: Player 1's turn? ✓
    R->>R: Roll dice
    R->>R: Update turn state
    R-->>A: Updated state
    A->>SSE: Broadcast update
    SSE-->>B1: Your roll result
    SSE-->>B2: Player 1 rolled
    A-->>B1: Game state

    B1->>A: POST /api/rooms/abc123/bank
    A->>R: Bank points
    R->>R: Add to Player 1 total
    R->>R: Next turn (Player 2)
    R-->>A: Updated state
    A->>SSE: Broadcast update
    SSE-->>B1: Turn ended
    SSE-->>B2: Your turn now!
    A-->>B1: Game state

    B2->>A: POST /api/rooms/abc123/roll
    A->>R: Validate turn (Player 2)
    R->>R: Player 2's turn? ✓
    Note over R: Player 2's turn...
```

## Real-Time Updates: Options Comparison

### Option A: Polling (Simplest)

```mermaid
sequenceDiagram
    participant B as Browser
    participant A as API
    participant R as GameRoom

    loop Every 2 seconds
        B->>A: GET /api/rooms/{id}/state
        A->>R: Get current state
        R-->>A: State
        A-->>B: Game state JSON
        B->>B: Update UI if changed
    end
```

**Pros:**
- Simple to implement
- No server-side state for connections

**Cons:**
- Wastes bandwidth
- 2-second delay for updates
- High server load

### Option B: Server-Sent Events (Recommended)

```mermaid
sequenceDiagram
    participant B as Browser
    participant SSE as SSE Endpoint
    participant R as GameRoom
    participant BC as Broadcaster

    B->>SSE: GET /api/rooms/{id}/events
    SSE->>BC: Register client
    BC-->>SSE: Connection open
    SSE-->>B: Connection established

    Note over R: Player rolls dice
    R->>BC: Broadcast update
    BC->>SSE: Send event
    SSE-->>B: data: {"event":"roll",...}
    B->>B: Update UI immediately

    Note over R: Player banks
    R->>BC: Broadcast update
    BC->>SSE: Send event
    SSE-->>B: data: {"event":"bank",...}
    B->>B: Update UI immediately
```

**Pros:**
- Real-time updates
- Server pushes to clients
- HTTP-based (no special ports)

**Cons:**
- One-way (server → client only)
- Need to track open connections

### Option C: WebSockets (Most Powerful)

```mermaid
sequenceDiagram
    participant B as Browser
    participant WS as WebSocket
    participant R as GameRoom

    B->>WS: ws://host/ws/rooms/{id}
    WS-->>B: Connection open

    B->>WS: {"action":"roll"}
    WS->>R: Process roll
    R-->>WS: Updated state
    WS-->>B: {"event":"roll","data":{...}}

    Note over R: Another player acts
    R->>WS: Broadcast
    WS-->>B: {"event":"update","data":{...}}
```

**Pros:**
- Two-way communication
- True real-time
- Can replace HTTP for actions

**Cons:**
- More complex
- Requires WebSocket library
- Need connection management

## Room Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Waiting: Create Room
    Waiting --> Waiting: Players Join
    Waiting --> InProgress: Start Game (2+ players)
    Waiting --> [*]: Timeout (no players)

    InProgress --> InProgress: Players Take Turns
    InProgress --> FinalRound: Player Reaches 10k

    FinalRound --> FinalRound: Other Players Get Final Turn
    FinalRound --> Finished: All Players Finished

    Finished --> [*]: Room Cleanup (after 1 hour)

    InProgress --> Abandoned: All Players Leave
    Abandoned --> [*]: Cleanup
```

## Turn Rotation Logic

```mermaid
graph LR
    P1[Player 1<br/>Turn] -->|Banks/Farkles| P2[Player 2<br/>Turn]
    P2 -->|Banks/Farkles| P3[Player 3<br/>Turn]
    P3 -->|Banks/Farkles| P4[Player 4<br/>Turn]
    P4 -->|Banks/Farkles| P1

    style P1 fill:#90EE90
    style P2 fill:#FFE4B5
    style P3 fill:#FFE4B5
    style P4 fill:#FFE4B5
```

**Turn End Conditions:**
1. Player banks points → Next player
2. Player farkles → Next player
3. Player manually ends turn → Next player

**Turn Validation:**
```
if currentPlayerIndex != playerRequestingAction:
    return error("Not your turn!")
```

## Deployment Options

### Option 1: Single Instance

```mermaid
graph TB
    LB[Load Balancer]
    LB --> S1[Single Farkle Server<br/>All Rooms In-Memory]
    S1 --> GC[Grafana Cloud]

    style S1 fill:#FFD700
```

**Pros:** Simple
**Cons:** No scaling, rooms lost on restart

### Option 2: Stateful Instances (Sticky Sessions)

```mermaid
graph TB
    LB[Load Balancer<br/>Sticky Sessions]
    LB --> S1[Server 1<br/>Rooms: A, B, C]
    LB --> S2[Server 2<br/>Rooms: D, E, F]
    LB --> S3[Server 3<br/>Rooms: G, H, I]

    S1 --> GC[Grafana Cloud]
    S2 --> GC
    S3 --> GC

    style S1 fill:#90EE90
    style S2 fill:#90EE90
    style S3 fill:#90EE90
```

**Pros:** Can scale
**Cons:** Need sticky sessions, room loss on restart

### Option 3: Shared State (Redis/Database)

```mermaid
graph TB
    LB[Load Balancer]
    LB --> S1[Server 1]
    LB --> S2[Server 2]
    LB --> S3[Server 3]

    S1 --> Redis[(Redis<br/>Shared State)]
    S2 --> Redis
    S3 --> Redis

    S1 --> GC[Grafana Cloud]
    S2 --> GC
    S3 --> GC

    style Redis fill:#FF6B6B
```

**Pros:** Scalable, persistent
**Cons:** More complex, added latency

**Recommendation for MVP:** Option 1 (single instance)

## Summary

The multiplayer architecture:
- Reuses **80% of existing game logic** (dice, scoring)
- Adds **room management** for multiple games
- Implements **turn-based coordination** between players
- Uses **real-time updates** (SSE recommended)
- Maintains **same observability** (metrics, traces, logs)

The single-player version remains **completely intact** - multiplayer is a separate implementation that shares the core game engine.
