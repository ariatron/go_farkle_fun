package api

import (
	"encoding/json"
	"farkle-app/internal/game"
	"net/http"
	"sync"
)

type GameState struct {
	Turn       *game.Turn `json:"turn"`
	LastRoll   []int      `json:"last_roll"`
	TotalBank  int        `json:"total_bank"`
	Winner     bool       `json:"winner"`
	mu         sync.Mutex
}

var currentGameState = &GameState{
	Turn: game.NewTurn(),
}

type KeepRequest struct {
	DiceToKeep []int `json:"dice_to_keep"`
}

func RollHandler(w http.ResponseWriter, r *http.Request) {
	currentGameState.mu.Lock()
	defer currentGameState.mu.Unlock()

	if currentGameState.Winner {
		renderJSON(w, currentGameState)
		return
	}

	// If this is a POST, the user is keeping dice from a previous roll
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

	// Perform the actual roll
	dice := currentGameState.Turn.Roll()
	currentGameState.LastRoll = dice

	// Check if this specific roll is a Farkle
	res := game.CalculateScore(dice)
	if res.Score == 0 {
		currentGameState.Turn.IsGameOver = true
		currentGameState.Turn.AccumulatedScore = 0
		currentGameState.Turn.DiceRemaining = 6
	} else {
		// Ensure it's not marked as game over if we have points
		currentGameState.Turn.IsGameOver = false
	}

	renderJSON(w, currentGameState)
}

func BankHandler(w http.ResponseWriter, r *http.Request) {
	currentGameState.mu.Lock()
	defer currentGameState.mu.Unlock()

	if !currentGameState.Turn.IsGameOver && !currentGameState.Winner {
		currentGameState.TotalBank += currentGameState.Turn.AccumulatedScore
	}

	if currentGameState.TotalBank >= 10000 {
		currentGameState.Winner = true
	}

	// Reset Turn and Clear Dice
	currentGameState.Turn = game.NewTurn()
	currentGameState.LastRoll = []int{} 

	renderJSON(w, currentGameState)
}

func ResetHandler(w http.ResponseWriter, r *http.Request) {
	currentGameState.mu.Lock()
	defer currentGameState.mu.Unlock()

	currentGameState.TotalBank = 0
	currentGameState.Winner = false
	currentGameState.Turn = game.NewTurn()
	currentGameState.LastRoll = []int{}

	renderJSON(w, currentGameState)
}

func renderJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}