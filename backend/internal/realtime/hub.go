package realtime

import "sync"

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[chan []byte]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[chan []byte]struct{})}
}

func (h *Hub) Register(pollID string) chan []byte {
	client := make(chan []byte, 16)
	h.mu.Lock()
	if h.clients[pollID] == nil {
		h.clients[pollID] = make(map[chan []byte]struct{})
	}
	h.clients[pollID][client] = struct{}{}
	h.mu.Unlock()
	return client
}

func (h *Hub) Unregister(pollID string, client chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients, ok := h.clients[pollID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.clients, pollID)
		}
	}
	close(client)
}

func (h *Hub) Broadcast(pollID string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients[pollID] {
		select {
		case client <- payload:
		default:
		}
	}
}
