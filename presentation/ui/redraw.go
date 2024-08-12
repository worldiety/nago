package ui

import (
	"context"
	"go.wdy.de/nago/presentation/core"
	"time"
)

func RedrawAtFixedRate[T core.View](wnd core.Window, rate time.Duration, v T) T {
	core.OnAppear(wnd, "", func(ctx context.Context) {
		for {
			if ctx.Err() != nil {
				break // exit
			}

			time.Sleep(rate)
			wnd.Invalidate()
		}
	})

	return v
}
