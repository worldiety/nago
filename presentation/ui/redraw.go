// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"context"
	"go.wdy.de/nago/presentation/core"
	"time"
)

// RedrawAtFixedRate just passes the given view and causes a redraw using the given rate.
// Note, that a rate lower than the applications fps rate, will have no effect and changes
// between render cycles become not visible.
func RedrawAtFixedRate[T core.View](wnd core.Window, rate time.Duration, v T) T {
	frames := core.AutoState[int64](wnd)
	core.OnAppear(wnd, "", func(ctx context.Context) {
		for {
			if ctx.Err() != nil {
				break // exit
			}

			time.Sleep(rate)
			// mark the state as dirty in the current render generation and cause a sliced redraw
			frames.Set(frames.Get() + 1)
		}
	})

	return v
}
