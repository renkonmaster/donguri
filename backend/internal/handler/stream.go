package handler

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ras0q/go-backend-template/internal/service/stream"
)

func (h *Handler) StreamRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomID, ok := parseRoomStreamPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	if _, err := uuid.Parse(roomID); err != nil {
		http.Error(w, "invalid room id", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", stream.ContentTypeEventStream)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	subscriber := stream.NewSubscriber(uuid.NewString(), 16)
	h.hub.Subscribe(roomID, subscriber)
	defer h.hub.Unsubscribe(roomID, subscriber.ID)

	if _, err := io.WriteString(w, stream.KeepAliveFrame); err != nil {
		return
	}
	flusher.Flush()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case payload := <-subscriber.Ch:
			if _, err := w.Write(payload); err != nil {
				return
			}
			flusher.Flush()
		case <-ticker.C:
			if _, err := io.WriteString(w, stream.KeepAliveFrame); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func parseRoomStreamPath(path string) (roomID string, ok bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 4 {
		return "", false
	}

	if parts[0] != "api" || parts[1] != "rooms" || parts[3] != "stream" {
		return "", false
	}
	if parts[2] == "" {
		return "", false
	}

	return parts[2], true
}
