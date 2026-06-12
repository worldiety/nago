// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package completion

import (
	"errors"
	"fmt"
)

// ContextWindowExceeded is the sentinel reported by a [Completions] provider when the request (the full
// history plus system prompt and tool definitions) does not fit into the model's context window. Use
// errors.Is(err, ContextWindowExceeded) to detect it. A provider should additionally wrap a
// [ContextWindowError] when it can extract the concrete numbers from the API response, so callers can size
// their compaction accordingly.
//
// This is the trigger the agentic loop in [Run] reacts to: instead of failing, it asks the configured
// [Compactor] to shrink the history and retries the turn.
var ContextWindowExceeded = errors.New("context window exceeded")

// ContextWindowError carries the (optional) provider-reported details of a context window overflow. It
// satisfies errors.Is(err, ContextWindowExceeded) and unwraps to [ContextWindowExceeded].
type ContextWindowError struct {
	// Limit is the model's maximum context window in tokens as reported by the provider, or 0 if unknown.
	Limit int
	// Tokens is the token count of the rejected request as reported by the provider, or 0 if unknown.
	Tokens int
}

func (e ContextWindowError) Error() string {
	if e.Limit > 0 || e.Tokens > 0 {
		return fmt.Sprintf("context window exceeded: %d tokens > %d maximum", e.Tokens, e.Limit)
	}
	return ContextWindowExceeded.Error()
}

func (e ContextWindowError) Is(target error) bool {
	return target == ContextWindowExceeded
}

func (e ContextWindowError) Unwrap() error {
	return ContextWindowExceeded
}

