package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"dockporter/internal/types"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan types.MigrationEvent
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.RWMutex // Only needed for counting clients
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan types.MigrationEvent, 100), // Buffered to prevent blocking
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		clients:    make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 WebSocket Hub shutting down...")
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("🔌 UI Client connected (Total: %d)", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()

		case event := <-h.broadcast:
			msg, _ := json.Marshal(event)
			h.mu.RLock()
			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Printf("⚠️ Failed to message client, closing: %v", err)
					client.Close()
					// We don't delete here to avoid mutating map during range
					// Use a separate chan or unregister if needed
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Publish(event types.MigrationEvent) {
	h.broadcast <- event
}

func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WS Upgrade error: %v", err)
		return
	}
	h.register <- conn
}
