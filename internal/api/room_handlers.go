package api

import (
	"encoding/json"
	"farkle-app/internal/game"
	"farkle-app/internal/observability"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// jsonError sends a JSON-formatted error response
func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// CreateRoomHandler handles room creation
func CreateRoomHandler(rm *game.RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)

		if r.Method != http.MethodPost {
			jsonError(w,"Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			MaxPlayers int `json:"max_players"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			req.MaxPlayers = 4 // Default to 4 players
		}

		room := rm.CreateRoom(req.MaxPlayers)

		span.SetAttributes(
			attribute.String("room_id", room.RoomID),
			attribute.Int("max_players", room.MaxPlayers),
		)

		observability.Logger.Info("Room created",
			"room_id", room.RoomID,
			"max_players", room.MaxPlayers,
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(room.GetState())
	}
}

// JoinRoomHandler handles player joining a room
func JoinRoomHandler(rm *game.RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)

		if r.Method != http.MethodPost {
			jsonError(w,"Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		roomID := extractRoomID(r.URL.Path)
		if roomID == "" {
			jsonError(w,"Room ID required", http.StatusBadRequest)
			return
		}

		var req struct {
			PlayerName string `json:"player_name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PlayerName == "" {
			jsonError(w,"Player name required", http.StatusBadRequest)
			return
		}

		room, err := rm.GetRoom(roomID)
		if err != nil {
			jsonError(w,"Room not found", http.StatusNotFound)
			return
		}

		player := game.NewPlayer(req.PlayerName)
		if err := room.AddPlayer(player); err != nil {
			jsonError(w,err.Error(), http.StatusBadRequest)
			return
		}

		span.SetAttributes(
			attribute.String("room_id", roomID),
			attribute.String("player_id", player.ID),
			attribute.String("player_name", player.Name),
		)

		observability.Logger.Info("Player joined room",
			"room_id", roomID,
			"player_id", player.ID,
			"player_name", player.Name,
		)

		response := map[string]interface{}{
			"player_id": player.ID,
			"room":      room.GetState(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// StartGameHandler handles starting a game
func StartGameHandler(rm *game.RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)

		if r.Method != http.MethodPost {
			jsonError(w,"Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		roomID := extractRoomID(r.URL.Path)
		if roomID == "" {
			jsonError(w,"Room ID required", http.StatusBadRequest)
			return
		}

		room, err := rm.GetRoom(roomID)
		if err != nil {
			jsonError(w,"Room not found", http.StatusNotFound)
			return
		}

		if err := room.StartGame(); err != nil {
			jsonError(w,err.Error(), http.StatusBadRequest)
			return
		}

		span.SetAttributes(
			attribute.String("room_id", roomID),
			attribute.Int("player_count", len(room.Players)),
		)

		observability.Logger.Info("Game started",
			"room_id", roomID,
			"player_count", len(room.Players),
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(room.GetState())
	}
}

// GetRoomStateHandler returns the current room state
func GetRoomStateHandler(rm *game.RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonError(w,"Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		roomID := extractRoomID(r.URL.Path)
		if roomID == "" {
			jsonError(w,"Room ID required", http.StatusBadRequest)
			return
		}

		room, err := rm.GetRoom(roomID)
		if err != nil {
			jsonError(w,"Room not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(room.GetState())
	}
}

// MultiplayerRollHandler handles dice rolling in multiplayer
func MultiplayerRollHandler(rm *game.RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)

		if r.Method != http.MethodPost {
			jsonError(w,"Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		roomID := extractRoomID(r.URL.Path)
		if roomID == "" {
			jsonError(w,"Room ID required", http.StatusBadRequest)
			return
		}

		room, err := rm.GetRoom(roomID)
		if err != nil {
			jsonError(w,"Room not found", http.StatusNotFound)
			return
		}

		var req struct {
			PlayerID    string `json:"player_id"`
			DiceToKeep  []int  `json:"dice_to_keep"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w,"Invalid request", http.StatusBadRequest)
			return
		}

		// Validate it's the player's turn
		if err := room.ValidatePlayerTurn(req.PlayerID); err != nil {
			jsonError(w,err.Error(), http.StatusForbidden)
			return
		}

		// Process dice to keep (similar to single-player)
		scoreResult := game.CalculateScore(req.DiceToKeep)
		score := scoreResult.Score
		if score == 0 && len(req.DiceToKeep) > 0 {
			jsonError(w,"Invalid dice selection: no scoring dice", http.StatusBadRequest)
			return
		}

		// Update turn state
		room.TurnState.AccumulatedScore += score
		room.TurnState.DiceRemaining -= len(req.DiceToKeep)

		// Check for hot dice
		if room.TurnState.DiceRemaining == 0 && score > 0 {
			room.TurnState.DiceRemaining = 6 // Hot dice! Get all dice back
		}

		// Roll remaining dice
		newRoll := room.TurnState.Roll()
		room.LastRoll = newRoll

		// Check for farkle
		if game.CalculateScore(newRoll).Score == 0 {
			room.TurnState.IsGameOver = true
			room.TurnState.AccumulatedScore = 0
			observability.Logger.Info("Player farkled", "room_id", roomID, "player_id", req.PlayerID)
		}

		span.SetAttributes(
			attribute.String("room_id", roomID),
			attribute.String("player_id", req.PlayerID),
			attribute.Int("accumulated_score", room.TurnState.AccumulatedScore),
			attribute.Bool("farkle", room.TurnState.IsGameOver),
		)

		observability.AppMetrics.RecordRoll(score)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(room.GetState())
	}
}

// MultiplayerBankHandler handles banking points in multiplayer
func MultiplayerBankHandler(rm *game.RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)

		if r.Method != http.MethodPost {
			jsonError(w,"Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		roomID := extractRoomID(r.URL.Path)
		if roomID == "" {
			jsonError(w,"Room ID required", http.StatusBadRequest)
			return
		}

		room, err := rm.GetRoom(roomID)
		if err != nil {
			jsonError(w,"Room not found", http.StatusNotFound)
			return
		}

		var req struct {
			PlayerID string `json:"player_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w,"Invalid request", http.StatusBadRequest)
			return
		}

		// Validate it's the player's turn
		if err := room.ValidatePlayerTurn(req.PlayerID); err != nil {
			jsonError(w,err.Error(), http.StatusForbidden)
			return
		}

		currentPlayer := room.GetCurrentPlayer()
		if currentPlayer == nil {
			jsonError(w,"No current player", http.StatusInternalServerError)
			return
		}

		// Check if player can bank (house rule: 500+ for first bank)
		if !currentPlayer.CanBank(room.TurnState.AccumulatedScore) {
			jsonError(w,"Must score at least 500 points for first bank", http.StatusBadRequest)
			return
		}

		// Bank the points
		currentPlayer.Bank(room.TurnState.AccumulatedScore)

		span.SetAttributes(
			attribute.String("room_id", roomID),
			attribute.String("player_id", req.PlayerID),
			attribute.Int("banked_points", room.TurnState.AccumulatedScore),
			attribute.Int("total_bank", currentPlayer.TotalBank),
		)

		observability.Logger.Info("Player banked points",
			"room_id", roomID,
			"player_id", req.PlayerID,
			"points", room.TurnState.AccumulatedScore,
			"total", currentPlayer.TotalBank,
		)

		observability.AppMetrics.RecordBank(room.TurnState.AccumulatedScore)

		// Check for winner
		room.CheckWinner()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(room.GetState())
	}
}

// EndTurnHandler handles ending a turn and moving to next player
func EndTurnHandler(rm *game.RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)

		if r.Method != http.MethodPost {
			jsonError(w,"Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		roomID := extractRoomID(r.URL.Path)
		if roomID == "" {
			jsonError(w,"Room ID required", http.StatusBadRequest)
			return
		}

		room, err := rm.GetRoom(roomID)
		if err != nil {
			jsonError(w,"Room not found", http.StatusNotFound)
			return
		}

		var req struct {
			PlayerID string `json:"player_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w,"Invalid request", http.StatusBadRequest)
			return
		}

		// Validate it's the player's turn
		if err := room.ValidatePlayerTurn(req.PlayerID); err != nil {
			jsonError(w,err.Error(), http.StatusForbidden)
			return
		}

		room.NextTurn()

		span.SetAttributes(
			attribute.String("room_id", roomID),
			attribute.String("player_id", req.PlayerID),
		)

		observability.Logger.Info("Turn ended",
			"room_id", roomID,
			"player_id", req.PlayerID,
			"next_player", room.GetCurrentPlayer().Name,
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(room.GetState())
	}
}

// LeaveRoomHandler handles a player leaving a room
func LeaveRoomHandler(rm *game.RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)

		if r.Method != http.MethodDelete {
			jsonError(w,"Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		roomID := extractRoomID(r.URL.Path)
		if roomID == "" {
			jsonError(w,"Room ID required", http.StatusBadRequest)
			return
		}

		room, err := rm.GetRoom(roomID)
		if err != nil {
			jsonError(w,"Room not found", http.StatusNotFound)
			return
		}

		var req struct {
			PlayerID string `json:"player_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w,"Invalid request", http.StatusBadRequest)
			return
		}

		if err := room.RemovePlayer(req.PlayerID); err != nil {
			jsonError(w,err.Error(), http.StatusBadRequest)
			return
		}

		span.SetAttributes(
			attribute.String("room_id", roomID),
			attribute.String("player_id", req.PlayerID),
		)

		observability.Logger.Info("Player left room",
			"room_id", roomID,
			"player_id", req.PlayerID,
		)

		// If room is empty, delete it
		if len(room.Players) == 0 {
			rm.DeleteRoom(roomID)
			observability.Logger.Info("Empty room deleted", "room_id", roomID)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "left"})
	}
}

// ListRoomsHandler lists all active rooms
func ListRoomsHandler(rm *game.RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonError(w,"Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rooms := rm.GetAllRooms()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"rooms": rooms,
			"count": len(rooms),
		})
	}
}

// extractRoomID extracts room ID from URL path
// Expects format: /api/rooms/{roomId}/...
func extractRoomID(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 4 && parts[1] == "api" && parts[2] == "rooms" {
		return parts[3]
	}
	return ""
}
