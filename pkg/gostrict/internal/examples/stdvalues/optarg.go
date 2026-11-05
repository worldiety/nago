// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/opt"

func optArgCall[T any](f func(T) opt.Option[T], v, def T) T {
	return f(v).Or(def)
}

// OptArg passes opt.Some, explicitly instantiated, to a generic function.
func OptArg(a, b int32) int32 {
	return optArgCall(opt.Some[int32], a, b)
}
