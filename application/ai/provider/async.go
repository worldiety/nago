// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package provider

import (
	"go.wdy.de/nago/pkg/xsync"
)

type AsyncResult[T any] func(result T, err error)

func Async[In, Out any](in In, fn func(In) (Out, error), resFn AsyncResult[Out]) {
	xsync.Go(func() error {
		out, err := fn(in)
		if err != nil {
			return err
		}

		resFn(out, nil)
		return nil
	}, func(err error) {
		var zero Out
		if err != nil {
			resFn(zero, err)
		}
	})
}
