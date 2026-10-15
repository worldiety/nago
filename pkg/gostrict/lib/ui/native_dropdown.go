// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License//go:build !gostrict

package ui

import "go.wdy.de/nago/presentation/ui/dropdown"

func dropdownOptionsOf(options []DropdownOption) []dropdown.Option[string] {
	res := make([]dropdown.Option[string], 0, len(options))
	for _, o := range options {
		res = append(res, dropdown.Option[string]{Value: o.Value, Label: o.Label, Description: o.Description, Disabled: o.Disabled})
	}
	return res
}
