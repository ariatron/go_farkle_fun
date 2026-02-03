package api

import (
	"encoding/json"
	"farkle-app/internal/game"
	"farkle-app/internal/observability"
	"net/http"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// GameState tracks the active game in memory
type GameState struct {
	PlayerName string     `json:"player_name"` // Player's name
	Turn       *game.Turn `json:"turn"`
	LastRoll   []int      `json:"last_roll"`
	TotalBank  int        `json:"total_bank"`
	Winner     bool       `json:"winner"`
	History    []int      `json:"history"` // Track last few successful banks
	mu         sync.Mutex
}

// Global state for our local server session
var currentGameState = &GameState{
	Turn:    game.NewTurn(),
	History: []int{},
}

// KeepRequest receives the dice values the user wants to score
type KeepRequest struct {
	DiceToKeep []int `json:"dice_to_keep"`
}

// SetPlayerNameRequest receives the player's name
type SetPlayerNameRequest struct {
	PlayerName string `json:"player_name"`
}

// RollHandler handles the rolling logic and processing kept dice
func RollHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	currentGameState.mu.Lock()
	defer currentGameState.mu.Unlock()

	if currentGameState.Winner {
		renderJSON(w, currentGameState)
		return
	}

	// Process dice kept from the PREVIOUS roll
	var keptDice []int
	if r.Method == http.MethodPost {
		var req KeepRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && len(req.DiceToKeep) > 0 {
			keptDice = req.DiceToKeep
			score, farkled := currentGameState.Turn.ProcessRoll(req.DiceToKeep)

			span.SetAttributes(
				attribute.IntSlice("kept_dice", keptDice),
				attribute.Int("score", score),
			)

			if farkled {
				// Record farkle event
				observability.AppMetrics.RecordFarkle()
				observability.Logger.InfoContext(ctx, "Player farkled",
					"player", currentGameState.PlayerName,
					"accumulated_score_lost", currentGameState.Turn.AccumulatedScore,
				)
				span.AddEvent("farkle")
				renderJSON(w, currentGameState)
				return
			}
		}
	}

	// Perform the new roll
	dice := currentGameState.Turn.Roll()
	currentGameState.LastRoll = dice

	span.SetAttributes(
		attribute.IntSlice("rolled_dice", dice),
		attribute.Int("dice_count", len(dice)),
	)

	// Check for immediate Farkle
	res := game.CalculateScore(dice)
	if res.Score == 0 {
		currentGameState.Turn.IsGameOver = true
		currentGameState.Turn.AccumulatedScore = 0
		currentGameState.Turn.DiceRemaining = 6

		// Record farkle event
		observability.AppMetrics.RecordFarkle()
		observability.Logger.InfoContext(ctx, "Player farkled on roll",
			"player", currentGameState.PlayerName,
			"dice", dice,
		)
		span.AddEvent("farkle_on_roll")
	} else {
		currentGameState.Turn.IsGameOver = false

		// Record roll event
		observability.AppMetrics.RecordRoll(res.Score)
		observability.Logger.InfoContext(ctx, "Dice rolled",
			"player", currentGameState.PlayerName,
			"dice", dice,
			"possible_score", res.Score,
			"accumulated", currentGameState.Turn.AccumulatedScore,
		)
	}

	renderJSON(w, currentGameState)
}

// BankHandler saves turn score, updates history, and checks for win
func BankHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	currentGameState.mu.Lock()
	defer currentGameState.mu.Unlock()

	if !currentGameState.Turn.IsGameOver && !currentGameState.Winner {
		score := currentGameState.Turn.AccumulatedScore
		if score > 0 {
			// House Rule: First bank must be at least 500 points
			if currentGameState.TotalBank == 0 && score < 500 {
				observability.Logger.WarnContext(ctx, "First bank attempt below minimum",
					"player", currentGameState.PlayerName,
					"attempted_score", score,
					"minimum_required", 500,
				)

				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "First bank must be at least 500 points",
				})
				return
			}

			currentGameState.TotalBank += score

			// Record bank event
			observability.AppMetrics.RecordBank(score)
			observability.Logger.InfoContext(ctx, "Points banked",
				"player", currentGameState.PlayerName,
				"banked_score", score,
				"total_bank", currentGameState.TotalBank,
			)

			span.SetAttributes(
				attribute.Int("banked_score", score),
				attribute.Int("total_bank", currentGameState.TotalBank),
			)

			// Add to history (newest at the top)
			currentGameState.History = append([]int{score}, currentGameState.History...)

			// Keep only the last 10 entries
			if len(currentGameState.History) > 10 {
				currentGameState.History = currentGameState.History[:10]
			}
		}
	}

	// Check for Win Condition
	if currentGameState.TotalBank >= 10000 {
		currentGameState.Winner = true

		// Record win event
		observability.AppMetrics.RecordWin()
		observability.Logger.InfoContext(ctx, "Player won the game!",
			"player", currentGameState.PlayerName,
			"final_score", currentGameState.TotalBank,
		)

		span.AddEvent("game_won", trace.WithAttributes(attribute.Int("final_score", currentGameState.TotalBank)))
	}

	// Reset turn state
	currentGameState.Turn = game.NewTurn()
	currentGameState.LastRoll = []int{}

	renderJSON(w, currentGameState)
}

// ResetHandler wipes the entire game state for a fresh start
func ResetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	currentGameState.mu.Lock()
	defer currentGameState.mu.Unlock()

	observability.Logger.InfoContext(ctx, "Game reset",
		"player", currentGameState.PlayerName,
		"previous_score", currentGameState.TotalBank,
	)

	currentGameState.TotalBank = 0
	currentGameState.Winner = false
	currentGameState.History = []int{}
	currentGameState.Turn = game.NewTurn()
	currentGameState.LastRoll = []int{}

	span.AddEvent("game_reset")

	renderJSON(w, currentGameState)
}

// SetPlayerNameHandler updates the player's name
func SetPlayerNameHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	currentGameState.mu.Lock()
	defer currentGameState.mu.Unlock()

	var req SetPlayerNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		observability.Logger.WarnContext(ctx, "Invalid player name request", "error", err)
		observability.AddSpanError(ctx, err)
		w.WriteHeader(http.StatusBadRequest)
		renderJSON(w, map[string]string{"error": "invalid request"})
		return
	}

	if req.PlayerName != "" {
		oldName := currentGameState.PlayerName
		currentGameState.PlayerName = req.PlayerName

		observability.Logger.InfoContext(ctx, "Player name updated",
			"old_name", oldName,
			"new_name", req.PlayerName,
		)

		span.SetAttributes(
			attribute.String("player_name", req.PlayerName),
		)
	}

	renderJSON(w, currentGameState)
}

// renderJSON helper
func renderJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
