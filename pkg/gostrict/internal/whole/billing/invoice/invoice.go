// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package invoice is a use case of the area billing.
package invoice

import (
	"errors"

	"go.wdy.de/nago/pkg/gostrict/internal/whole/store"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// Main writes an invoice and fails for a negative amount, which rolls the invoice back.
func Main(customer string, amount int32) (string, error) {
	store.Put("invoice/"+customer, text.FromInt(amount))
	if amount < 0 {
		return "", errors.New("negative amount")
	}
	return "invoice " + text.FromInt(amount), nil
}
