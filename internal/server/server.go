package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Server provides a gracefully-stoppable http server implementation. It is safe
// for concurrent use in goroutines.
type Server struct {
	ip       string
	port     string
	listener net.Listener
}

// NewServer creates a new server listening on the provided address that responds to
// the http.Handler. It starts the listener, but does not start the server. If
// an empty port is given, the server randomly chooses one.
func NewServer(host string, port int) (*Server, error) {
	// create the net listener first, so the connection ready when we return. This
	// guarantees that it can accept requests.
	addr := fmt.Sprintf(host + ":" + strconv.Itoa(port))

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener on %s: %w", addr, err)
	}

	return &Server{
		ip:       listener.Addr().(*net.TCPAddr).IP.String(),        //nolint:forcetypeassert
		port:     strconv.Itoa(listener.Addr().(*net.TCPAddr).Port), //nolint:forcetypeassert
		listener: listener,
	}, nil
}

// ServeHTTP starts the server and blocks until the provided context is closed.
// When the provided context is closed, the server is gracefully stopped with a
// timeout of 5 seconds.
//
// Once a server has been stopped, it is NOT safe for reuse.
func (s *Server) ServeHTTP(logger *slog.Logger, ctx context.Context, srv *http.Server) error {

	// Spawn a goroutine that listens for context closure. When the context is
	// closed, the server is stopped.
	errCh := make(chan error, 1)
	go func() {
		<-ctx.Done()
		logger.Debug("http.Serve: context closed")

		shutdownCtx, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()

		logger.Debug("http.Serve: shutting down")
		errCh <- srv.Shutdown(shutdownCtx)
	}()

	// Run the server. This will block until the provided context is closed.
	if err := srv.Serve(s.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to serve: %w", err)
	}

	logger.Debug("http.Serve: serving stopped")

	merr := wrapError{
		msg: "failed to shutdown server",
	}

	// Return any errors that happened during shutdown.
	if err := <-errCh; err != nil {
		merr.errs = append(merr.errs, err)
	}

	if len(merr.errs) > 0 {
		return merr
	}

	return nil
}

// ServeHTTPHandler is a convenience wrapper around ServeHTTP. It creates an
// HTTP server using the provided handler.
func (s *Server) ServeHTTPHandler(logger *slog.Logger, ctx context.Context, handler http.Handler) error {
	return s.ServeHTTP(logger, ctx, &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Handler:           handler,
	})
}

type wrapError struct {
	msg  string
	errs []error
}

func (e wrapError) Error() string {
	msg := ": "
	for _, err := range e.errs {
		msg += err.Error() + "|"
	}

	return e.msg + msg
}

func (e wrapError) Unwrap() []error {
	return e.errs
}
