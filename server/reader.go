package server

import (
	"context"
	"io"
)

// Handler is a callback function to process messages from io.Reader.
type Handler func(Message) error

// Message is a message struct to be received/published.
type Message struct {
	Data []byte
}

// StreamReader is the server listener that reads messages from the
// byte stream provided by io.Reader.
type StreamReader struct {
	reader   io.ReadCloser
	handlers []Handler
}

// NewStreamReader returns a new instance of StreamReader.
func NewStreamReader(reader io.ReadCloser, handlers ...Handler) *StreamReader {
	return &StreamReader{
		reader:   reader,
		handlers: handlers,
	}
}

// ListenAndServe continuously reads bytes from io.Reader
// and invokes handlers once the complete message is received.
func (l *StreamReader) ListenAndServe() error {
	for {
		b, err := io.ReadAll(l.reader)
		if err != nil {
			return err
		}

		for _, handler := range l.handlers {
			if handler == nil {
				continue
			}

			if err := handler(Message{Data: b}); err != nil {
				return err
			}
		}
	}
}

// Shutdown closes the io.Reader.
func (l *StreamReader) Shutdown(context.Context) error {
	return l.reader.Close()
}
