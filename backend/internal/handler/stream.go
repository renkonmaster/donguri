package handler

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/renkonmaster/donguri/internal/service/stream"
)

func (h *Handler) newRoomStreamReader(ctx context.Context, roomID string) io.ReadCloser {
	reader, writer := io.Pipe()
	subscriber := stream.NewSubscriber(uuid.NewString(), 16)
	h.hub.Subscribe(roomID, subscriber)

	go func() {
		defer h.hub.Unsubscribe(roomID, subscriber.ID)
		defer writer.Close()

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case payload := <-subscriber.Ch:
				if _, err := writer.Write(payload); err != nil {
					return
				}
			case <-ticker.C:
				if _, err := io.WriteString(writer, stream.KeepAliveFrame); err != nil {
					return
				}
			}
		}
	}()

	return reader
}
