// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package std

// Try is best used with defer and an according function pointer, which is evaluated when the defer runs the try.
func Try(f func() error, err *error) {
	newErr := f()
	if *err == nil {
		*err = newErr
	}
}
