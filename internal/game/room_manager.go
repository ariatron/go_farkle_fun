package game

import (
	"errors"
	"sync"
	"time"
)

// RoomManager manages all active game rooms
type RoomManager struct {
	rooms map[string]*GameRoom
	mu    sync.RWMutex
}

// NewRoomManager creates a new room manager
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*GameRoom),
	}
}

// CreateRoom creates a new game room
func (rm *RoomManager) CreateRoom(maxPlayers int) *GameRoom {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room := NewGameRoom(maxPlayers)
	rm.rooms[room.RoomID] = room
	return room
}

// GetRoom retrieves a room by ID
func (rm *RoomManager) GetRoom(roomID string) (*GameRoom, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return nil, errors.New("room not found")
	}

	return room, nil
}

// GetAllRooms returns all active rooms
func (rm *RoomManager) GetAllRooms() []*GameRoom {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rooms := make([]*GameRoom, 0, len(rm.rooms))
	for _, room := range rm.rooms {
		rooms = append(rooms, room.GetState())
	}

	return rooms
}

// DeleteRoom removes a room
func (rm *RoomManager) DeleteRoom(roomID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.rooms[roomID]; !exists {
		return errors.New("room not found")
	}

	delete(rm.rooms, roomID)
	return nil
}

// CleanupStaleRooms removes rooms that have been inactive
func (rm *RoomManager) CleanupStaleRooms(timeout time.Duration) int {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	staleRooms := make([]string, 0)

	for roomID, room := range rm.rooms {
		if room.IsStale(timeout) {
			staleRooms = append(staleRooms, roomID)
		}
	}

	for _, roomID := range staleRooms {
		delete(rm.rooms, roomID)
	}

	return len(staleRooms)
}

// GetRoomCount returns the number of active rooms
func (rm *RoomManager) GetRoomCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return len(rm.rooms)
}

// GetPlayerCount returns the total number of players across all rooms
func (rm *RoomManager) GetPlayerCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	count := 0
	for _, room := range rm.rooms {
		room.mu.Lock()
		count += len(room.Players)
		room.mu.Unlock()
	}

	return count
}
