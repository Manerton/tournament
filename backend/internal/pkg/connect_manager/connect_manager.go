package connect_manager

import (
	"encoding/json"
	"sync"

	"golang.org/x/net/websocket"
)

type Hub struct {
	clients map[string]map[*websocket.Conn]bool
	mu      sync.Mutex
}

func (h *Hub) AddClient(eventID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[eventID] == nil {
		h.clients[eventID] = make(map[*websocket.Conn]bool)
	}
	h.clients[eventID][conn] = true
}

func (h *Hub) RemoveClient(eventID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients[eventID], conn)
}

func (h *Hub) Broadcast(eventID string, message any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	data, _ := json.Marshal(message)
	for conn := range h.clients[eventID] {
		conn.Write(data)
	}
}
