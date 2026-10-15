// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package entry has exported functions, which can not be entry points of a whole program.
package entry

// Shape is a sealed interface.
type Shape interface{ isShape() }

// Circle is a variant.
type Circle struct{ R float64 }

func (Circle) isShape() {}

// Main is a valid entry point.
func Main(n int32) int32 { return n }

// Identity is generic.
func Identity[T any](v T) T { return v }

// Area takes a sealed interface.
func Area(s Shape) float64 { return 1 }
