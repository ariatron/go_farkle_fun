package game

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// GameStatus represents the current status of a game room
type GameStatus string

const (
	StatusWaiting    GameStatus = "waiting"     // Waiting for players to join
	StatusInProgress GameStatus = "in_progress" // Game is active
	StatusFinished   GameStatus = "finished"    // Game has ended
)

// GameRoom represents a multiplayer game instance
type GameRoom struct {
	RoomID             string      `json:"room_id"`
	Players            []*Player   `json:"players"`
	CurrentPlayerIndex int         `json:"current_player_index"`
	GameStatus         GameStatus  `json:"game_status"`
	TurnState          *Turn       `json:"turn_state"`
	LastRoll           []int       `json:"last_roll"`
	MaxPlayers         int         `json:"max_players"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
	FinalRoundStarted  bool        `json:"final_round_started"`
	FinalRoundStarter  int         `json:"final_round_starter"` // Index of player who triggered final round
	mu                 sync.Mutex  // Protects room state
}

// NewGameRoom creates a new game room with specified max players
func NewGameRoom(maxPlayers int) *GameRoom {
	if maxPlayers < 2 {
		maxPlayers = 2
	}
	if maxPlayers > 6 {
		maxPlayers = 6
	}

	return &GameRoom{
		RoomID:             generateRoomID(),
		Players:            make([]*Player, 0),
		CurrentPlayerIndex: 0,
		GameStatus:         StatusWaiting,
		TurnState:          NewTurn(),
		LastRoll:           []int{},
		MaxPlayers:         maxPlayers,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		FinalRoundStarted:  false,
		FinalRoundStarter:  -1,
	}
}

// generateRoomID generates a short, human-readable room code
func generateRoomID() string {
	// Generate UUID and take first 8 characters for a shorter code
	id := uuid.New().String()
	return id[:8]
}

// AddPlayer adds a player to the room
func (r *GameRoom) AddPlayer(player *Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.GameStatus != StatusWaiting {
		return errors.New("cannot join: game already started")
	}

	if len(r.Players) >= r.MaxPlayers {
		return errors.New("room is full")
	}

	// Check for duplicate names
	for _, p := range r.Players {
		if p.Name == player.Name {
			return errors.New("player name already taken")
		}
	}

	r.Players = append(r.Players, player)
	r.UpdatedAt = time.Now()
	return nil
}

// RemovePlayer removes a player from the room
func (r *GameRoom) RemovePlayer(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, p := range r.Players {
		if p.ID == playerID {
			r.Players = append(r.Players[:i], r.Players[i+1:]...)
			r.UpdatedAt = time.Now()

			// If game in progress and we removed current player, advance to next
			if r.GameStatus == StatusInProgress && i == r.CurrentPlayerIndex {
				r.advanceToNextPlayer()
			}
			return nil
		}
	}

	return errors.New("player not found")
}

// StartGame starts the game if conditions are met
func (r *GameRoom) StartGame() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) < 2 {
		return errors.New("need at least 2 players to start")
	}

	if r.GameStatus != StatusWaiting {
		return errors.New("game already started")
	}

	r.GameStatus = StatusInProgress
	r.CurrentPlayerIndex = 0
	r.TurnState = NewTurn()
	r.UpdatedAt = time.Now()
	return nil
}

// GetCurrentPlayer returns the current player
func (r *GameRoom) GetCurrentPlayer() *Player {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) == 0 {
		return nil
	}

	return r.Players[r.CurrentPlayerIndex]
}

// ValidatePlayerTurn checks if it's the specified player's turn
func (r *GameRoom) ValidatePlayerTurn(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.GameStatus != StatusInProgress {
		return errors.New("game not in progress")
	}

	if len(r.Players) == 0 {
		return errors.New("no players in game")
	}

	currentPlayer := r.Players[r.CurrentPlayerIndex]
	if currentPlayer.ID != playerID {
		return errors.New("not your turn")
	}

	return nil
}

// NextTurn advances to the next player's turn
func (r *GameRoom) NextTurn() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Increment turn count for current player
	r.Players[r.CurrentPlayerIndex].TurnCount++

	// Check if we should end the final round
	if r.FinalRoundStarted {
		// If everyone after the starter has had their turn, end game
		nextIndex := (r.CurrentPlayerIndex + 1) % len(r.Players)
		if nextIndex == (r.FinalRoundStarter+1)%len(r.Players) {
			r.GameStatus = StatusFinished
			r.determineWinner()
			r.UpdatedAt = time.Now()
			return
		}
	}

	r.advanceToNextPlayer()
	r.UpdatedAt = time.Now()
}

// advanceToNextPlayer moves to the next player (internal, must hold lock)
func (r *GameRoom) advanceToNextPlayer() {
	r.CurrentPlayerIndex = (r.CurrentPlayerIndex + 1) % len(r.Players)
	r.TurnState = NewTurn()
	r.LastRoll = []int{}
}

// CheckWinner checks if any player has won and starts final round if needed
func (r *GameRoom) CheckWinner() *Player {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, player := range r.Players {
		if player.TotalBank >= 10000 && !r.FinalRoundStarted {
			// Start final round
			r.FinalRoundStarted = true
			r.FinalRoundStarter = i
			return nil // Game continues for final round
		}
	}

	// If final round is complete, determine winner
	if r.GameStatus == StatusFinished {
		return r.getWinnerInternal()
	}

	return nil
}

// determineWinner finds the winner after final round (internal, must hold lock)
func (r *GameRoom) determineWinner() {
	if len(r.Players) == 0 {
		return
	}

	winner := r.Players[0]
	for _, player := range r.Players[1:] {
		if player.TotalBank > winner.TotalBank {
			winner = player
		}
	}

	winner.IsWinner = true
}

// getWinnerInternal returns the winner (internal, must hold lock)
func (r *GameRoom) getWinnerInternal() *Player {
	for _, player := range r.Players {
		if player.IsWinner {
			return player
		}
	}
	return nil
}

// GetWinner returns the winner (thread-safe)
func (r *GameRoom) GetWinner() *Player {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getWinnerInternal()
}

// IsStale checks if the room has been inactive for too long
func (r *GameRoom) IsStale(timeout time.Duration) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return time.Since(r.UpdatedAt) > timeout
}

// GetState returns the current room state (thread-safe)
func (r *GameRoom) GetState() *GameRoom {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Return a copy to avoid race conditions
	state := &GameRoom{
		RoomID:             r.RoomID,
		Players:            make([]*Player, len(r.Players)),
		CurrentPlayerIndex: r.CurrentPlayerIndex,
		GameStatus:         r.GameStatus,
		TurnState:          &Turn{
			AccumulatedScore: r.TurnState.AccumulatedScore,
			DiceRemaining:    r.TurnState.DiceRemaining,
			IsGameOver:       r.TurnState.IsGameOver,
		},
		LastRoll:          make([]int, len(r.LastRoll)),
		MaxPlayers:        r.MaxPlayers,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
		FinalRoundStarted: r.FinalRoundStarted,
		FinalRoundStarter: r.FinalRoundStarter,
	}

	// Deep copy players
	for i, p := range r.Players {
		playerCopy := *p
		state.Players[i] = &playerCopy
	}

	// Copy last roll
	copy(state.LastRoll, r.LastRoll)

	return state
}
