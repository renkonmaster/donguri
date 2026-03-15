package stream

import (
	"testing"
	"time"
)

func TestHub_SubscribePublishUnsubscribe(t *testing.T) {
	t.Parallel()

	hub := NewHub()
	roomID := "room-1"

	subA := NewSubscriber("a", 1)
	subB := NewSubscriber("b", 1)

	hub.Subscribe(roomID, subA)
	hub.Subscribe(roomID, subB)

	if got := hub.SubscriberCount(roomID); got != 2 {
		t.Fatalf("unexpected subscriber count: got=%d want=2", got)
	}

	payload := []byte("event: room_updated\ndata: hello\n\n")
	hub.Publish(roomID, payload)

	select {
	case got := <-subA.Ch:
		if string(got) != string(payload) {
			t.Fatalf("unexpected payload for subA: got=%q want=%q", string(got), string(payload))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting payload for subA")
	}

	select {
	case got := <-subB.Ch:
		if string(got) != string(payload) {
			t.Fatalf("unexpected payload for subB: got=%q want=%q", string(got), string(payload))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting payload for subB")
	}

	hub.Unsubscribe(roomID, subA.ID)
	if got := hub.SubscriberCount(roomID); got != 1 {
		t.Fatalf("unexpected subscriber count after unsubscribe subA: got=%d want=1", got)
	}

	hub.Unsubscribe(roomID, subB.ID)
	if got := hub.SubscriberCount(roomID); got != 0 {
		t.Fatalf("unexpected subscriber count after unsubscribe subB: got=%d want=0", got)
	}
}

func TestHub_PublishIsNonBlockingForFullBuffer(t *testing.T) {
	t.Parallel()

	hub := NewHub()
	roomID := "room-2"
	sub := NewSubscriber("slow", 1)
	hub.Subscribe(roomID, sub)

	hub.Publish(roomID, []byte("first"))

	done := make(chan struct{})
	go func() {
		hub.Publish(roomID, []byte("second"))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("publish blocked on full buffer")
	}
}
