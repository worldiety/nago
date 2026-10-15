// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package c4 is the end of a chain of use cases.
package c4

// Main returns n.
func Main(n int32) int32 {
	return depth(n)
}

func depth(n int32) int32 {
	return n
}
