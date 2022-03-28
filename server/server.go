package server

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/diptanw/go-toolkit/logger"
)

// Listener is an interface for server listener.
type Listener interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

// Server is a wrapper for server listener. It handles process termination and
// shutdown.
type Server struct {
	srv Listener
	log logger.Logger
}

// New returns a new instance of Server.
func New(srv Listener, log logger.Logger) Server {
	return Server{srv: srv, log: log}
}

// Serve starts a new server listener and handles interrupt and termination
// signals.
func (s Server) Serve(ctx context.Context) error {
	errsCh := make(chan error)

	go func() {
		s.log.Infof("server: starting...")
		errsCh <- s.srv.ListenAndServe()
	}()

	var err error

	defer func() {
		// Wait for completion before exiting.
		timedCtx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		s.log.Infof("server: shutting down...")
		err = s.srv.Shutdown(timedCtx)
	}()

	signalCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()

	// Block until an error or signal is received.
	select {
	case err = <-errsCh:
		return err
	case <-signalCtx.Done():
		return err
	}
}
