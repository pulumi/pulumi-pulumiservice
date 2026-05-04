// Copyright 2016-2026, Pulumi Corporation.  All rights reserved.

package apiclient

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ssePayload struct {
	Msg string `json:"msg"`
}

func newReader(stream string) *SSEReader[ssePayload] {
	return NewSSEReader(
		io.NopCloser(strings.NewReader(stream)),
		func(data []byte, v *ssePayload) error { return json.Unmarshal(data, v) },
	)
}

func TestSSEReader_SingleEvent(t *testing.T) {
	r := newReader("id: 1\nevent: hello\ndata: {\"msg\":\"hi\"}\n\n")
	defer r.Close()

	ev, err := r.Next()
	require.NoError(t, err)
	assert.Equal(t, "1", ev.ID)
	assert.Equal(t, "hello", ev.Event)
	assert.Equal(t, "hi", ev.Data.Msg)
}

func TestSSEReader_MultipleEvents(t *testing.T) {
	r := newReader("data: {\"msg\":\"a\"}\n\ndata: {\"msg\":\"b\"}\n\n")
	defer r.Close()

	ev1, err := r.Next()
	require.NoError(t, err)
	assert.Equal(t, "a", ev1.Data.Msg)

	ev2, err := r.Next()
	require.NoError(t, err)
	assert.Equal(t, "b", ev2.Data.Msg)
}

func TestSSEReader_HeartbeatsAndComments(t *testing.T) {
	stream := ": heartbeat\n\n: another comment\n\ndata: {\"msg\":\"real\"}\n\n"
	r := newReader(stream)
	defer r.Close()

	ev, err := r.Next()
	require.NoError(t, err)
	assert.Equal(t, "real", ev.Data.Msg)
}

func TestSSEReader_EOFAfterEvent(t *testing.T) {
	r := newReader("data: {\"msg\":\"only\"}\n\n")
	defer r.Close()

	ev, err := r.Next()
	require.NoError(t, err)
	assert.Equal(t, "only", ev.Data.Msg)

	_, err = r.Next()
	assert.ErrorIs(t, err, io.EOF)
}

func TestSSEReader_EOFOnEmptyStream(t *testing.T) {
	r := newReader("")
	defer r.Close()

	_, err := r.Next()
	assert.ErrorIs(t, err, io.EOF)
}

func TestSSEReader_DataWithoutTrailingBlankLine(t *testing.T) {
	r := newReader("data: {\"msg\":\"trailing\"}\n")
	defer r.Close()

	ev, err := r.Next()
	require.NoError(t, err)
	assert.Equal(t, "trailing", ev.Data.Msg)
}

func TestSSEReader_UnmarshalError(t *testing.T) {
	r := newReader("data: not-json\n\n")
	defer r.Close()

	_, err := r.Next()
	require.Error(t, err)
	var syntaxErr *json.SyntaxError
	assert.True(t, errors.As(err, &syntaxErr) || strings.Contains(err.Error(), "invalid"))
}

func TestSSEReader_CRLFLineEndings(t *testing.T) {
	r := newReader("id: 7\r\ndata: {\"msg\":\"crlf\"}\r\n\r\n")
	defer r.Close()

	ev, err := r.Next()
	require.NoError(t, err)
	assert.Equal(t, "7", ev.ID)
	assert.Equal(t, "crlf", ev.Data.Msg)
}

func TestSSEReader_Close(t *testing.T) {
	r := newReader("data: {\"msg\":\"x\"}\n\n")
	assert.NoError(t, r.Close())
}
