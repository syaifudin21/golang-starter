package websocket

import (
	"encoding/json"
	"exam/internal/dtos"
	"fmt"
	"log"
)

// Hub maintains the set of active rooms and broadcasts messages to the
// rooms.
type Hub struct {
	// Registered rooms.
	Rooms map[string]*Room

	// Register requests for rooms.
	Register chan *Room

	// Unregister requests for rooms.
	Unregister chan *Room
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Room),
		Unregister: make(chan *Room),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case room := <-h.Register:
			h.Rooms[room.QuizID] = room
			log.Printf("Room %s registered", room.QuizID)
		case room := <-h.Unregister:
			if _, ok := h.Rooms[room.QuizID]; ok {
				delete(h.Rooms, room.QuizID)
				log.Printf("Room %s unregistered", room.QuizID)
			}
		}
	}
}

func (h *Hub) GetRoomClientCount(quizUUID string) int {
	if room, ok := h.Rooms[quizUUID]; ok {
		return len(room.Clients)
	}
	return 0
}

func (h *Hub) GetRoomClients(quizUUID string) []dtos.ConnectedStudentDTO {
	var students []dtos.ConnectedStudentDTO
	if room, ok := h.Rooms[quizUUID]; ok {
		for client := range room.Clients {
			// Assuming client.UserName is available or can be derived
			// For now, we'll use a placeholder or derive from UserID
			userName := fmt.Sprintf("Player %d", client.UserID)
			students = append(students, dtos.ConnectedStudentDTO{
				UserID:   client.UserID,
				UserName: userName,
			})
		}
	}
	return students
}

func (h *Hub) StartQuizInRoom(quizUUID string, sessionID uint, mode string) error {
	if room, ok := h.Rooms[quizUUID]; ok {
		// Marshal the session ID and mode into a JSON payload
		payload, err := json.Marshal(struct {
			SessionID uint   `json:"session_id"`
			Mode      string `json:"mode"`
		}{
			SessionID: sessionID,
			Mode:      mode,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal start_game payload: %w", err)
		}
		// Send a message to the room's inbound channel to start the game
		// This simulates the "start_game" websocket message but from the API
		room.Inbound <- &InboundMessage{Type: "start_game", Payload: payload}
		return nil
	}
	return fmt.Errorf("quiz room %s not found", quizUUID)
}