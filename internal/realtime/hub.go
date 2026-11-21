package realtime

import "sync"

const clientSendBufferSize = 16

type Hub struct {
	mu        sync.RWMutex
	clients   map[string]map[*Client]struct{}
	closeOnce sync.Once
	done      chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*Client]struct{}),
		done:    make(chan struct{}),
	}
}

func NewClient(userID string) *Client {
	return &Client{
		userID: userID,
		send:   make(chan Event, clientSendBufferSize),
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	userClients, ok := h.clients[client.userID]
	if !ok {
		userClients = make(map[*Client]struct{})
		h.clients[client.userID] = userClients
	}

	userClients[client] = struct{}{}

}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	userClients, ok := h.clients[client.userID]
	if !ok {
		return
	}
	delete(userClients, client)

	if len(userClients) == 0 {
		delete(h.clients, client.userID)
	}
}

func (h *Hub) Publish(userID string, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userClients := h.clients[userID]

	for client := range userClients {
		select {
		case client.send <- event:
		default:
		}
	}
}

func (h *Hub) Done() <-chan struct{} {
	return h.done
}

func (h *Hub) Close() {
	h.closeOnce.Do(func() {
		close(h.done)
	})
}
