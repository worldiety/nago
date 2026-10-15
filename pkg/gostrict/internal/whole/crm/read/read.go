// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package read is a use case of the area crm.
package read

import (
	"errors"

	"go.wdy.de/nago/pkg/gostrict/internal/whole/store"
)

// Main returns the name of the customer.
func Main(id string) (string, error) {
	name, ok := store.Get("customer/" + id)
	if !ok {
		return "", errors.New("unknown customer " + id)
	}
	return name, nil
}
