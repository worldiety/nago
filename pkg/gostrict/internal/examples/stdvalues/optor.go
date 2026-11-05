// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/opt"

// OptOr reads an option through the method value Or.
func OptOr(a, b int32) int32 {
	f := opt.Some(a).Or
	return f(b) + b
}
