// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build gostrict

package bin

// Seek continues at the byte offset.
//
//gostrict:js (r, o) => $bin_seek(r, o)
func (r Reader) Seek(offset int64) {}
