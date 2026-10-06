// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package logging

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
)

type slogKey string

// FromContext extracts the current logger from the context or returns the system default. See also [WithContext].
func FromContext(ctx context.Context) *slog.Logger {
	logger := ctx.Value(slogKey("logger"))
	if logger == nil {
		return slog.Default()
	}

	return logger.(*slog.Logger)
}

// WithContext allocates a new context with the supplied logger value. See also [FromContext].
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, slogKey("logger"), logger)
}

// Secret returns a short, stable fingerprint of a secret identifier like a session or scope id, which grants
// access to whoever knows it. Logs must never contain such ids, but the fingerprint still correlates log lines.
func Secret(id string) string {
	if id == "" {
		return ""
	}

	sum := sha256.Sum256([]byte(id))
	return "#" + hex.EncodeToString(sum[:4])
}
