package websocket_handler

import (
	"main/internal/pkg/connect_manager"

	"github.com/go-chi/chi/v5"
	"golang.org/x/net/websocket"
)

type WebSocketHandler struct {
	Hub *connect_manager.Hub
}

func (s *WebSocketHandler) StartTimer(conn *websocket.Conn) {
	r := conn.Request()
	eventID := chi.URLParam(r, "id")
	if eventID == "" {
		conn.Close()
		return
	}

	// client := &Client{

	// }

}
