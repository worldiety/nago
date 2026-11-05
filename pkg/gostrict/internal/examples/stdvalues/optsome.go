// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/opt"

// OptSome creates an option through opt.Some as a value.
func OptSome(a, b int32) int32 {
	f := opt.Some[int32]
	return f(a).Or(b)
}
