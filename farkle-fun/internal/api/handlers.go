package api

import (
	"encoding/json"
	"farkle-app/internal/game"
	"net/http"
	"sync"
)

// GameState tracks the active game in memory
type GameState struct {
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

// RollHandler handles the rolling logic and processing kept dice
func RollHandler(w http.ResponseWriter, r *http.Request) {
	currentGameState.mu.Lock()
	defer currentGameState.mu.Unlock()

	if currentGameState.Winner {
		renderJSON(w, currentGameState)
		return
	}

	// Process dice kept from the PREVIOUS roll
	if r.Method == http.MethodPost {
		var req KeepRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && len(req.DiceToKeep) > 0 {
			_, farkled := currentGameState.Turn.ProcessRoll(req.DiceToKeep)
			if farkled {
				renderJSON(w, currentGameState)
				return
			}
		}
	}

	// Perform the new roll
	dice := currentGameState.Turn.Roll()
	currentGameState.LastRoll = dice

	// Check for immediate Farkle
	res := game.CalculateScore(dice)
	if res.Score == 0 {
		currentGameState.Turn.IsGameOver = true
		currentGameState.Turn.AccumulatedScore = 0
		currentGameState.Turn.DiceRemaining = 6
	} else {
		currentGameState.Turn.IsGameOver = false
	}

	renderJSON(w, currentGameState)
}

// BankHandler saves turn score, updates history, and checks for win
func BankHandler(w http.ResponseWriter, r *http.Request) {
	currentGameState.mu.Lock()
	defer currentGameState.mu.Unlock()

	if !currentGameState.Turn.IsGameOver && !currentGameState.Winner {
		score := currentGameState.Turn.AccumulatedScore
		if score > 0 {
			currentGameState.TotalBank += score
			
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
	}

	// Reset turn state
	currentGameState.Turn = game.NewTurn()
	currentGameState.LastRoll = []int{}

	renderJSON(w, currentGameState)
}

// ResetHandler wipes the entire game state for a fresh start
func ResetHandler(w http.ResponseWriter, r *http.Request) {
	currentGameState.mu.Lock()
	defer currentGameState.mu.Unlock()

	currentGameState.TotalBank = 0
	currentGameState.Winner = false
	currentGameState.History = []int{}
	currentGameState.Turn = game.NewTurn()
	currentGameState.LastRoll = []int{}

	renderJSON(w, currentGameState)
}

// renderJSON helper
func renderJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}