// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/opt"

func optInferCall[T any](f func(T) opt.Option[T], v, def T) T {
	return f(v).Or(def)
}

// OptInfer passes opt.Some to a generic function, which infers its type argument.
func OptInfer(a, b int32) int32 {
	return optInferCall(opt.Some, a, b)
}
