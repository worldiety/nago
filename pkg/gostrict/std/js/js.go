// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package js provides the reference, which a handle of a header package holds. A handle is a struct type of a
// header package with a single unexported field of type [Ref], e.g.
//
//	type Mesh struct{ ref js.Ref }
//
// In JavaScript, a handle is the object of the host itself, e.g. a THREE.Mesh, so it costs nothing. A native
// implementation of the header package decides, what its handles refer to. Only header packages import this
// package.
//
//gostrict:header
package js

// Ref refers to an object of the host. The zero value refers to nothing and never occurs in a checked program.
type Ref struct {
	v any
}

// NewRef returns a reference to v.
func NewRef(v any) Ref {
	return Ref{v: v}
}

// Value returns, what the reference refers to.
func (r Ref) Value() any {
	return r.v
}
