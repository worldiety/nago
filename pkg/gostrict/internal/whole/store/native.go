// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package store

// Backend implements the store natively.
var Backend interface {
	Get(key string) (string, bool)
	Put(key, value string)
}
