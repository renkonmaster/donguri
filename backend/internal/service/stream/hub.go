package stream

import "sync"

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[string]chan []byte
}

func NewHub() *Hub {
	return &Hub{ 
		rooms: map[string]map[string]chan []byte{},
	}
}

func (h *Hub) Subscribe(roomID string, sub Subscriber) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[roomID]; !ok {
		h.rooms[roomID] = map[string]chan []byte{}
	}

	h.rooms[roomID][sub.ID] = sub.Ch
}

func (h *Hub) Unsubscribe(roomID, subscriberID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.rooms[roomID]
	if !ok {
		return
	}

	delete(room, subscriberID)
	if len(room) == 0 {
		delete(h.rooms, roomID)
	}
}

func (h *Hub) Publish(roomID string, payload []byte) {
	h.mu.RLock()
	room, ok := h.rooms[roomID]
	if !ok {
		h.mu.RUnlock()
		return
	}

	chs := make([]chan []byte, 0, len(room))
	for _, ch := range room {
		chs = append(chs, ch)
	}
	h.mu.RUnlock()

	for _, ch := range chs {
		select {
		case ch <- payload:
		default:
		}
	}
}

func (h *Hub) SubscriberCount(roomID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, ok := h.rooms[roomID]
	if !ok {
		return 0
	}

	return len(room)
}
