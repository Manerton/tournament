package websocket_handler

import (
	"log"
	"main/internal/pkg/connect_manager"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/net/websocket"
)

type WebSocketHandler struct {
	Hub *connect_manager.Hub
}

func New(hub *connect_manager.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		Hub: hub,
	}
}

func (h *WebSocketHandler) StartTimer(conn *websocket.Conn) {
	r := conn.Request()
	eventID := chi.URLParam(r, "id")
	if eventID == "" {
		conn.Close()
		return
	}

	client := &connect_manager.Client{
		Conn:    conn,
		EventID: eventID,
	}

	h.Hub.AddClient(eventID, client)
	defer func() {
		h.Hub.RemoveClient(eventID, client)
		conn.Close()

	}()

	for {
		var raw string
		if err := websocket.Message.Receive(conn, &raw); err != nil {
			log.Println("[ws] read error", err)
			return
		}

		log.Printf("[ws] recv form event=%s: %s", eventID, raw)

		switch raw {
		case "start_timer":
			// отправим всем в этом event'е сообщение с серверным временем старта
			h.Hub.BroadcastToEvent(eventID, connect_manager.Message{
				Action: "start_timer",
				Payload: map[string]interface{}{
					"start_at": time.Now().Unix(),
				},
			})
		case "reset_timer":
			h.Hub.BroadcastToEvent(eventID, connect_manager.Message{
				Action: "reset_timer",
			})
		default:
			// можно парсить JSON и реагировать на более сложные payload'ы
			log.Println("[ws] unknown command:", raw)
		}
	}

}
