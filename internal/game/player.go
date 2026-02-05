package game

import (
	"time"

	"github.com/google/uuid"
)

// Player represents a player in a multiplayer game
type Player struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	TotalBank    int       `json:"total_bank"`
	TurnCount    int       `json:"turn_count"`
	IsWinner     bool      `json:"is_winner"`
	HasFirstBank bool      `json:"has_first_bank"` // House rule: must bank 500+ first
	JoinedAt     time.Time `json:"joined_at"`
}

// NewPlayer creates a new player with a generated UUID
func NewPlayer(name string) *Player {
	return &Player{
		ID:           uuid.New().String(),
		Name:         name,
		TotalBank:    0,
		TurnCount:    0,
		IsWinner:     false,
		HasFirstBank: false,
		JoinedAt:     time.Now(),
	}
}

// CanBank checks if player can bank points (house rule: 500+ for first bank)
func (p *Player) CanBank(points int) bool {
	if p.HasFirstBank {
		return true
	}
	return points >= 500
}

// Bank adds points to player's total bank
func (p *Player) Bank(points int) {
	p.TotalBank += points
	if !p.HasFirstBank && points >= 500 {
		p.HasFirstBank = true
	}
}
