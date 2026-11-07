// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package guard is a header of the test of the mark of an aspect: Reject ends the entry point, in which it runs, with
// an error, see gostrict.ReturnError and enter_value_test.go.
//
//gostrict:header
package guard

// Reject ends the entry point, in which it runs, with the error msg.
//
//gostrict:host
func Reject(msg string) {
	Backend.Reject(msg)
}
