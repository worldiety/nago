// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package store is the database of the test of a whole program: the host keeps a transaction for each frame of an
// area, see the aspect of whole_test.go.
//
//gostrict:header
package store

// Get returns the value of the key in the transaction of the current frame.
//
//gostrict:host
func Get(key string) (string, bool) {
	return Backend.Get(key)
}

// Put sets the value of the key in the transaction of the current frame.
//
//gostrict:host
func Put(key, value string) {
	Backend.Put(key, value)
}
