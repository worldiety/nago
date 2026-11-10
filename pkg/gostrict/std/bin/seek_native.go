// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package bin

// offset is an int64. Seek takes it in Go under this name, because go vet takes a Seek with an int64 for the one of
// io.Seeker. The subset sees an int64, see seek.go.
type offset = int64

// Seek continues at the byte offset.
func (r Reader) Seek(o offset) { r.c().SeekTo(o) }
