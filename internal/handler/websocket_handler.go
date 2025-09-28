package handler

import (
	"exam/internal/service"
	appWebsocket "exam/internal/websocket"
	"log"
	"net/http"

	ws "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections for now
	},
}

// WebsocketHandler handles the websocket connection requests.
type WebsocketHandler struct {
	hub         *appWebsocket.Hub
	quizService *service.QuizService
}

func NewWebsocketHandler(hub *appWebsocket.Hub, quizService *service.QuizService) *WebsocketHandler {
	return &WebsocketHandler{hub: hub, quizService: quizService}
}

// ServeWs handles websocket requests from the peer.
func (h *WebsocketHandler) ServeWs(c echo.Context) error {
	quizUUID := c.Param("quizUUID")
	if quizUUID == "" {
		return c.String(http.StatusBadRequest, "Quiz UUID is required")
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println(err)
		return err
	}

	room, exists := h.hub.Rooms[quizUUID]
	if !exists {
		room = appWebsocket.NewRoom(quizUUID, h.quizService)
		h.hub.Register <- room
		go room.Run()
	}

	userID := c.Get("userID").(uint)

	client := &appWebsocket.Client{
		Room:   room,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: userID,
	}

	client.Room.Register <- client

	go client.WritePump()
	go client.ReadPump()

	log.Printf("Client %d connected to quiz %s", userID, quizUUID)
	return nil
}
