package application

import (
	"log/slog"
	"os"
)

// Logger returns the applications default logger.
func (c *Configurator) Logger() *slog.Logger {
	if c.appName == "" {
		panic("set app name first")
	}

	if c.logger != nil {
		return c.logger
	}

	if c.debug {
		c.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	} else {
		c.logger = slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(slog.String("app", c.appName))
	}

	return c.logger
}

// defaultLogger always returns a logger.
func (c *Configurator) defaultLogger() *slog.Logger {
	if c == nil {
		return slog.Default()
	}

	if c.logger != nil {
		return c.logger
	}

	return slog.Default()
}
