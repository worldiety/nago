package application

import (
	"go.wdy.de/nago/logging"
	"log/slog"
	"net/http"
	"os"
)

// Logger returns the applications default logger and initializes also the globals slog default once.
func (c *Configurator) Logger() *slog.Logger {

	if c.logger != nil {
		return c.logger
	}

	if c.debug {
		c.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	} else {
		c.logger = slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(slog.String("app", string(c.ApplicationID())))
	}

	slog.SetDefault(c.logger)

	return c.logger
}

// defaultLogger always returns a logger.
func (c *Configurator) defaultLogger() *slog.Logger {
	if c == nil {
		return slog.Default()
	}

	if c.applicationID != "" { // try to init that now
		return c.Logger()
	}

	if c.logger != nil {
		return c.logger
	}

	return slog.Default()
}

func (c *Configurator) loggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := c.defaultLogger().With(slog.String("url", r.URL.String()))
		r = r.WithContext(logging.WithContext(r.Context(), logger))
		h.ServeHTTP(w, r)
	})
}
