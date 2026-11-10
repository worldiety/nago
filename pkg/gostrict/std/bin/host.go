// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package bin

import "go.wdy.de/nago/pkg/gostrict/std/bin/core"

func (r Reader) c() *core.Reader {
	v, _ := r.ref.Value().(*core.Reader)
	return v
}

func (b Buffer) c() *core.Buffer {
	v, _ := b.ref.Value().(*core.Buffer)
	return v
}

// The methods of this file convert bytes for a host in Go, e.g. into the values of an engine, see gostrict.ToNeutral.
// The subset does not see them.

// HostBytes returns the bytes, which the host must not change.
func (b Bytes) HostBytes() []byte {
	return raw(b)
}

// HostFromBytes returns bytes for the memory of b, which must not change anymore.
func (Bytes) HostFromBytes(b []byte) any {
	return wrap(b)
}

// OfSlice returns bytes for the memory of b, which must not change anymore.
func OfSlice(b []byte) Bytes {
	return wrap(b)
}
