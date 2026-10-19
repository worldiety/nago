// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package lookup calls the entry point of another package through a function value, which the aspect surrounds as
// well.
package lookup

import "go.wdy.de/nago/pkg/gostrict/internal/whole/crm/read"

// Main reads the customers through a function value and joins their names.
func Main(a, b string) string {
	find := read.Main
	return name(find, a) + "+" + name(find, b)
}

func name(find func(id string) (string, error), id string) string {
	n, err := find(id)
	if err != nil {
		return err.Error()
	}
	return n
}
