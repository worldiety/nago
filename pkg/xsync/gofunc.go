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
	"os"
	"runtime/debug"
	"sync"
)

func GoFn(fn func()) *sync.WaitGroup {
	return Go(func() error {
		fn()
		return nil
	}, func(err error) {
		if err != nil {
			slog.Error("failed to execute Go function", "err", err.Error())
		}
	})
}

// Go runs the given closure concurrently and execute it with a panic recover guard. The onDone callback may be
// nil and is either called with an error or with nil.
func Go(fn func() error, onDone func(error)) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := os.LookupEnv("XPC_SERVICE_NAME"); ok {
					debug.PrintStack()
				}
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
