// Package server provides a transport service for the application.
package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server is a struct that wraps an http transport listener.
type Server struct {
	srv Listener
}

// Listener is an interface for http server listener.
type Listener interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

// New returns a new instance of Server.
func New(srv Listener) *Server {
	return &Server{srv: srv}
}

// Serve runs the bootstrapped http server listener.
func (s *Server) Serve(ctx context.Context) error {
	errsCh := make(chan error)

	go func() {
		errsCh <- s.srv.ListenAndServe()
	}()

	// Listen for interrupt and termination signals
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	// Block until an error or a signal is received
	select {
	case err := <-errsCh:
		if err == http.ErrServerClosed {
			return err
		}
	case <-shutdownCh:
		return errors.New("process terminated")
	}

	err := s.srv.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("shutting down: %w", err)
	}

	// Cancel slow operations when exiting
	_, cancel := context.WithTimeout(context.Background(), time.Second)
	cancel()

	return err
}
