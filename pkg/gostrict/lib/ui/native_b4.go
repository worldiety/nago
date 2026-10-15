// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package ui

import (
	nui "go.wdy.de/nago/presentation/ui"
)

// The conversions of batch b4, see native.go. The lists are nil, if there are no elements, like the variadic
// arguments of nago.

func b4ColumnsOf(columns []TTableColumn) []nui.TTableColumn {
	var res []nui.TTableColumn
	for _, c := range columns {
		res = append(res, get[nui.TTableColumn](c.ref))
	}
	return res
}

func b4CellsOf(cells []TTableCell) []nui.TTableCell {
	var res []nui.TTableCell
	for _, c := range cells {
		res = append(res, get[nui.TTableCell](c.ref))
	}
	return res
}

func b4RowsOf(rows []TTableRow) []nui.TTableRow {
	var res []nui.TTableRow
	for _, r := range rows {
		res = append(res, get[nui.TTableRow](r.ref))
	}
	return res
}

func b4GridCellsOf(cells []TGridCell) []nui.TGridCell {
	var res []nui.TGridCell
	for _, c := range cells {
		res = append(res, get[nui.TGridCell](c.ref))
	}
	return res
}

func b4GroupsOf(groups []TMenuGroup) []nui.TMenuGroup {
	var res []nui.TMenuGroup
	for _, g := range groups {
		res = append(res, get[nui.TMenuGroup](g.ref))
	}
	return res
}

func b4ItemsOf(items []TMenuItem) []nui.TMenuItem {
	var res []nui.TMenuItem
	for _, i := range items {
		res = append(res, get[nui.TMenuItem](i.ref))
	}
	return res
}

func b4LengthsOf(lengths []Length) []nui.Length {
	var res []nui.Length
	for _, l := range lengths {
		res = append(res, nui.Length(l))
	}
	return res
}
