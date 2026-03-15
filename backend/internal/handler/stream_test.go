package handler

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/renkonmaster/donguri/internal/api"
	"github.com/renkonmaster/donguri/internal/service/stream"
)

func TestNewRoomStreamReader_SubscribeAndCleanup(t *testing.T) {
	t.Parallel()

	hub := stream.NewHub()
	h := New(nil, hub)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	roomID := "room-1"
	reader := h.newRoomStreamReader(ctx, roomID)
	defer reader.Close()

	if got := hub.SubscriberCount(roomID); got != 1 {
		t.Fatalf("unexpected subscriber count after subscribe: got=%d want=1", got)
	}

	cancel()

	buf := make([]byte, 1)
	_, _ = reader.Read(buf)

	deadline := time.Now().Add(500 * time.Millisecond)
	for {
		if hub.SubscriberCount(roomID) == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("subscriber was not cleaned up after context cancellation")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestSubscribeRoomStream_ReturnsReader(t *testing.T) {
	t.Parallel()

	h := New(nil, stream.NewHub())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := h.SubscribeRoomStream(ctx, api.SubscribeRoomStreamParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data == nil {
		t.Fatal("expected response Data reader")
	}

	if closer, ok := resp.Data.(io.Closer); ok {
		_ = closer.Close()
	}
}

func TestPublishRoomEvent_DeliveredToSubscriber(t *testing.T) {
	t.Parallel()

	h := New(nil, stream.NewHub())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	roomID := uuid.New()
	reader := h.newRoomStreamReader(ctx, roomID.String())
	defer reader.Close()

	h.publishRoomEvent(roomID, "room_updated", []byte(`{"status":"playing"}`))

	buf := make([]byte, 256)
	deadline := time.Now().Add(500 * time.Millisecond)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			t.Fatalf("failed to read stream: %v", err)
		}

		chunk := string(buf[:n])
		if strings.Contains(chunk, "event: room_updated") {
			if !strings.Contains(chunk, `data: {"status":"playing"}`) {
				t.Fatalf("event data missing: %q", chunk)
			}
			break
		}

		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for room_updated event, last chunk=%q", chunk)
		}
	}
}
