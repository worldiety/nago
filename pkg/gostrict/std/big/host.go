// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package big

import "math/big"

// The methods of this file convert an Int for a host in Go, e.g. into the values of an engine, see
// gostrict.ToNeutral. The subset does not see them.

// HostString returns the decimal of the integer.
func (i Int) HostString() string {
	return String(i)
}

// HostFromString returns the integer of a decimal, or 0 for an invalid one.
func (Int) HostFromString(s string) any {
	v, _ := Parse(s)
	return v
}

// Of returns the integer of v for a host in Go.
func Of(v *big.Int) Int {
	return wrap(new(big.Int).Set(v))
}

// Big returns a copy of the integer as math/big.Int for a host in Go.
func (i Int) Big() *big.Int {
	return new(big.Int).Set(val(i))
}
