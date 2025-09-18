// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xsync

import (
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
)

// Go runs the given closure concurrently and execute it with a panic recover guard. The onDone callback may be
// nil and is either called with an error or with nil.
func Go(fn func() error, onDone func(error)) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error(fmt.Sprintf("%v", r), slog.String("stack", string(debug.Stack())))
				if onDone != nil {
					onDone(fmt.Errorf("panic: %v", r))
				}
			}
		}()

		if err := fn(); err != nil {
			if onDone != nil {
				onDone(err)
			}
		} else {
			if onDone != nil {
				onDone(nil)
			}
		}
	})

	return &wg
}
