package game

import (
	"math/rand"
	"time"
)

type Turn struct {
	AccumulatedScore int  `json:"accumulated_score"` // Score banked in previous rolls this turn
	DiceRemaining    int  `json:"dice_remaining"`    // Starts at 6
	IsGameOver       bool `json:"is_game_over"`      // True if they Farkled
}

func NewTurn() *Turn {
	return &Turn{DiceRemaining: 6}
}

// Roll simulates rolling the remaining dice
func (t *Turn) Roll() []int {
	rand.Seed(time.Now().UnixNano())
	dice := make([]int, t.DiceRemaining)
	for i := 0; i < t.DiceRemaining; i++ {
		dice[i] = rand.Intn(6) + 1
	}
	return dice
}

// ProcessRoll calculates the score and manages "Hot Dice"
func (t *Turn) ProcessRoll(keptDice []int) (int, bool) {
	result := CalculateScore(keptDice)
	
	if result.Score == 0 {
		t.IsGameOver = true
		t.AccumulatedScore = 0
		return 0, true // Farkle!
	}

	t.AccumulatedScore += result.Score
	
	// Hot Dice Logic: 
	// If all dice used, reset to 6. Otherwise, subtract used dice.
	remaining := t.DiceRemaining - result.DiceUsed
	if remaining == 0 {
		t.DiceRemaining = 6 // Hot Dice!
	} else {
		t.DiceRemaining = remaining
	}

	return result.Score, false
}