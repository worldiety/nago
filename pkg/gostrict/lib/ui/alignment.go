// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

// Alignment positions a child within its container. The zero Alignment is Center.
type Alignment uint32

const (
	Center         Alignment = 0
	Top            Alignment = 1
	Bottom         Alignment = 2
	Leading        Alignment = 3
	Trailing       Alignment = 4
	TopLeading     Alignment = 5
	TopTrailing    Alignment = 6
	BottomLeading  Alignment = 7
	BottomTrailing Alignment = 8
	Stretch        Alignment = 9
)
