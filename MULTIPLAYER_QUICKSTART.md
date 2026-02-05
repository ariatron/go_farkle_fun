# Multiplayer Farkle - Quick Reference

## What Changes?

### Single-Player (Current)
```
1 player → rolls dice → banks points → wins at 10k
```

### Multiplayer (Proposed)
```
2-6 players → take turns → first to 10k triggers final round → highest score wins
```

## Key Additions

### 1. **Game Rooms**
- Each game is a "room" with a unique ID
- Players join rooms before playing
- 2-6 players per room

### 2. **Turn-Based Play**
- Players take turns in rotation
- Only current player can roll/bank
- Turn ends when player banks or farkles

### 3. **Real-Time Updates**
- All players see game state updates
- Using Server-Sent Events or WebSockets
- See other players' scores and actions

## File Changes

```
✅ REUSABLE (no changes):
   - internal/game/engine.go      (dice rolling logic)
   - internal/game/scoring.go     (scoring calculations)

🆕 NEW FILES:
   - internal/game/room.go        (GameRoom struct)
   - internal/game/player.go      (Player struct)
   - internal/game/room_manager.go (room state management)
   - internal/api/room_handlers.go (new API endpoints)
   - static/multiplayer.html      (multiplayer UI)
   - tests/k6/multiplayer-*.js    (multiplayer load tests)

📝 UPDATED FILES:
   - internal/observability/metrics.go (add room/player metrics)
   - README.md                         (add multiplayer docs)
```

## New API Endpoints

```
POST   /api/rooms/create           - Create game room
POST   /api/rooms/{id}/join        - Join room
POST   /api/rooms/{id}/start       - Start game
POST   /api/rooms/{id}/roll        - Roll dice (current player)
POST   /api/rooms/{id}/bank        - Bank points (current player)
POST   /api/rooms/{id}/end-turn    - Pass turn to next player
GET    /api/rooms/{id}/state       - Get game state
GET    /api/rooms/{id}/events      - Real-time updates (SSE)
```

## Game Flow

```
┌─────────────────┐
│ 1. Create Room  │  Player 1 creates room → gets room code
└────────┬────────┘
         │
┌────────▼────────┐
│ 2. Join Room    │  Players 2-6 join using room code
└────────┬────────┘
         │
┌────────▼────────┐
│ 3. Start Game   │  Host starts when 2+ players ready
└────────┬────────┘
         │
┌────────▼────────┐
│ 4. Take Turns   │  Players rotate turns
│    - Roll       │  Only current player can act
│    - Bank       │  Others wait and watch
│    - End Turn   │
└────────┬────────┘
         │
┌────────▼────────┐
│ 5. Final Round  │  When someone hits 10k, everyone gets
│                 │  one final turn
└────────┬────────┘
         │
┌────────▼────────┐
│ 6. Winner!      │  Highest score wins
└─────────────────┘
```

## Testing Strategy

### Unit Tests (20+ tests)
```
✅ Room creation & player joining
✅ Turn rotation & validation
✅ Win condition & final round
✅ Room capacity limits
✅ Player disconnection handling
```

### Integration Tests (10+ tests)
```
✅ Full game flow (create → join → play → win)
✅ Concurrent room access
✅ Invalid turn attempts
✅ Room cleanup
```

### Load Tests (K6)
```
✅ 100 concurrent games
✅ 1000 players joining rooms
✅ Turn-by-turn performance
✅ Real-time update latency
```

## Implementation Phases

### Phase 1: Core Logic (2-3 days)
- GameRoom, Player, RoomManager
- Turn rotation logic
- Unit tests

### Phase 2: API (2-3 days)
- Room endpoints
- Turn validation
- API tests

### Phase 3: Frontend (2-3 days)
- Lobby screen
- Waiting room
- Multiplayer game UI

### Phase 4: Real-Time (1-2 days)
- Server-Sent Events
- Auto-refresh
- Disconnection handling

### Phase 5: Testing (1-2 days)
- K6 load tests
- Metrics & observability
- Dashboard updates

### Phase 6: Polish (1 day)
- Documentation
- Error handling
- Final testing

**Total: 9-14 days**

## Recommendation

**Option A: Separate Project (Recommended)**
```
/farkle-fun              ← Keep original untouched
/farkle-multiplayer      ← New multiplayer version
```

**Option B: Same Project, Different Branch**
```
main branch              ← Single-player (original)
multiplayer branch       ← Multiplayer version
```

**Option C: Feature Flag**
```
GAME_MODE=single → Single-player
GAME_MODE=multi  → Multiplayer
```

## Key Challenges

1. **State Sync** → Use mutex locks + transactional updates
2. **Real-Time Updates** → SSE or WebSockets
3. **Player Disconnects** → Timeout + skip turn
4. **Room Cleanup** → Background job for stale rooms
5. **Concurrent Access** → Strict turn validation

## Metrics to Track

```
farkle_active_rooms              - Number of active games
farkle_players_online            - Total players in all rooms
farkle_turns_per_game            - Average turns per game
farkle_game_duration_seconds     - How long games last
farkle_rooms_created_total       - Total rooms created
farkle_player_disconnects_total  - Disconnection rate
```

## Next Steps

1. Review MULTIPLAYER_PLAN.md (full details)
2. Decide: separate project vs branch vs feature flag
3. Start with Phase 1 (core logic)
4. Test extensively before moving to next phase

See **MULTIPLAYER_PLAN.md** for complete implementation details.
