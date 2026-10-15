// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package create is a use case of the area crm, which calls a use case of its area and one of the area billing.
package create

import (
	"go.wdy.de/nago/pkg/gostrict/internal/whole/billing/invoice"
	"go.wdy.de/nago/pkg/gostrict/internal/whole/crm/read"
	"go.wdy.de/nago/pkg/gostrict/internal/whole/store"
)

// Main creates a customer and its invoice.
func Main(id, name string, amount int32) (string, error) {
	store.Put("customer/"+id, name)
	// a function value of an entry point calls it within the aspect as well
	readName := read.Main
	got, err := readName(id)
	if err != nil {
		return "", err
	}
	inv, err := invoice.Main(id, amount)
	if err != nil {
		return got + " without invoice: " + err.Error(), nil
	}
	return got + " " + inv, nil
}
