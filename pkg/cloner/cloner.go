// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cloner

// Cloner declares a contract which performs a deep copy of T. The instantiation for use in type declarations
// is like
//
//	type X[SomeType Cloner[SomeType]] ...
//
// See also packages xmaps and xslices which provide according clone implementations.
type Cloner[T any] interface {
	// Clone returns a deep copy of itself. An implementation must not return mutable memory shared with
	// any clone.
	Clone() T
}

// Cloneable is an incompatible interface until self rereferencing generics are allowed (F-bounded).
type Cloneable interface {
	Clone() Cloneable
}
