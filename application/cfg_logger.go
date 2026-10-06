// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"go.wdy.de/nago/logging"
)

// Logger returns the applications default logger and initializes also the globals slog default once.
func (c *Configurator) Logger() *slog.Logger {

	if c.logger != nil {
		return c.logger
	}

	if c.IsDebug() {
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
		logger := c.defaultLogger().With(slog.String("url", redactURL(r.URL)))
		r = r.WithContext(logging.WithContext(r.Context(), logger))
		h.ServeHTTP(w, r)
	})
}

// redactURL replaces the scope id of the wire, which grants access to a window, by its fingerprint.
func redactURL(u *url.URL) string {
	q := u.Query()
	if sid := q.Get("_sid"); sid != "" {
		c := *u
		q.Set("_sid", logging.Secret(sid))
		c.RawQuery = q.Encode()
		return c.String()
	}

	return u.String()
}
