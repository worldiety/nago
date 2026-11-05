// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/list"

// ListLen counts through the method value of a list.
func ListLen(a, b int32) int32 {
	l := list.Of(a, b)
	f := l.Len
	return 10*f() + b
}
