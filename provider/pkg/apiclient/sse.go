// Copyright 2026, Pulumi Corporation.  All rights reserved.

package apiclient

import (
	"bufio"
	"io"
	"strings"
)

// SSEEvent is the envelope for a single Server-Sent Event. T is the type of the
// data payload.
type SSEEvent[T any] struct {
	// ID is the event ID, used with Last-Event-ID for resumption.
	ID string

	// Event is the event type. Empty means the default "message" type.
	Event string

	// Data is the deserialized event payload.
	Data T
}

// SSEReader reads Server-Sent Events from an HTTP response body and
// deserializes each event's data field as JSON into type T.
type SSEReader[T any] struct {
	body        io.ReadCloser
	reader      *bufio.Reader
	unmarshaler func([]byte, *T) error
}

// NewSSEReader creates a new SSEReader that reads from the given body.
func NewSSEReader[T any](body io.ReadCloser, unmarshaler func([]byte, *T) error) *SSEReader[T] {
	return &SSEReader[T]{
		body:        body,
		reader:      bufio.NewReader(body),
		unmarshaler: unmarshaler,
	}
}

// Next reads the next SSE event. Returns io.EOF when the stream ends.
// Heartbeat-only events (comment with no data) are skipped.
func (r *SSEReader[T]) Next() (SSEEvent[T], error) {
	var event SSEEvent[T]
	var hasData bool

	for {
		line, err := r.reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")

		if err != nil && line == "" {
			if hasData {
				return event, nil
			}
			if err == io.EOF {
				return event, io.EOF
			}
			return event, err
		}

		// Empty line = end of event frame.
		if line == "" {
			if hasData {
				return event, nil
			}
			// Reset for next event (was a heartbeat or comment-only frame).
			event = SSEEvent[T]{}
			continue
		}

		if strings.HasPrefix(line, ":") {
			// Comment line (heartbeat or debug); skip.
			continue
		} else if id, ok := strings.CutPrefix(line, "id: "); ok {
			event.ID = id
		} else if eventType, ok := strings.CutPrefix(line, "event: "); ok {
			event.Event = eventType
		} else if data, ok := strings.CutPrefix(line, "data: "); ok {
			if err := r.unmarshaler([]byte(data), &event.Data); err != nil {
				return event, err
			}
			hasData = true
		}
	}
}

// Close closes the underlying response body.
func (r *SSEReader[T]) Close() error {
	return r.body.Close()
}
