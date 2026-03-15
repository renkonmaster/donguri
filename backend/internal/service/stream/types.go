package stream

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	ContentTypeEventStream = "text/event-stream"
	KeepAliveFrame         = ":\n\n"
)

type Event struct {
	Name string
	Data []byte
}

type Subscriber struct {
	ID string
	Ch chan []byte
}

func NewSubscriber(id string, bufferSize int) Subscriber {
	if bufferSize <= 0 {
		bufferSize = 1
	}

	return Subscriber{
		ID: id,
		Ch: make(chan []byte, bufferSize),
	}
}

func FormatEvent(event Event) []byte {
	name := event.Name
	if name == "" {
		name = "message"
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, "event: %s\n", name)

	data := string(event.Data)
	if data == "" {
		b.WriteString("data:\n")
		b.WriteString("\n")

		return b.Bytes()
	}

	for _, line := range strings.Split(data, "\n") {
		fmt.Fprintf(&b, "data: %s\n", line)
	}

	b.WriteString("\n")

	return b.Bytes()
}
