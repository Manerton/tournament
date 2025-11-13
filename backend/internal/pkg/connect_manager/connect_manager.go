package connect_manager

import (
	"encoding/json"
	"log"
	"sync"

	"golang.org/x/net/websocket"
)

// Сообщение стандартное
type Message struct {
	Action  string      `json:"action"`
	Payload interface{} `json:"payload,omitempty"`
}

// Клиент
type Client struct {
	Conn    *websocket.Conn
	EventID string
}

// Hub хранит клиентов по eventId
type Hub struct {
	mu      sync.Mutex
	clients map[string]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*Client]struct{}),
	}
}

func (h *Hub) AddClient(eventID string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[eventID]; !ok {
		h.clients[eventID] = make(map[*Client]struct{})
	}
	h.clients[eventID][c] = struct{}{}
	log.Printf("[hub] client added for event=%s, total=%d", eventID, len(h.clients[eventID]))
}

func (h *Hub) RemoveClient(eventID string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.clients[eventID]; ok {
		delete(set, c)
		if len(set) == 0 {
			delete(h.clients, eventID)
		}
	}
	// close connection (caller может уже закрыть)
}

func (h *Hub) BroadcastToEvent(eventID string, msg Message) {
	h.mu.Lock()
	set := h.clients[eventID]
	h.mu.Unlock()

	if set == nil {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("[hub] marshal error:", err)
		return
	}

	for c := range set {
		if err := websocket.Message.Send(c.Conn, string(data)); err != nil {
			log.Println("[hub] send error, removing client:", err)
			// при ошибке закрываем соединение и удаляем клиента
			_ = c.Conn.Close()
			h.RemoveClient(eventID, c)
		}
	}
}

// Broadcast to all events (опционально)
func (h *Hub) BroadcastAll(msg Message) {
	h.mu.Lock()
	clientsCopy := make(map[string][]*Client, len(h.clients))
	for id, set := range h.clients {
		for c := range set {
			clientsCopy[id] = append(clientsCopy[id], c)
		}
	}
	h.mu.Unlock()

	data, _ := json.Marshal(msg)
	for _, lst := range clientsCopy {
		for _, c := range lst {
			if err := websocket.Message.Send(c.Conn, string(data)); err != nil {
				log.Println("[hub] send all error:", err)
				_ = c.Conn.Close()
				h.RemoveClient(c.EventID, c)
			}
		}
	}
}

// Graceful close all connections (например при остановке сервера)
func (h *Hub) CloseAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, set := range h.clients {
		for c := range set {
			_ = c.Conn.Close()
		}
	}
	h.clients = make(map[string]map[*Client]struct{})
}
