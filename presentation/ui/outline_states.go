// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/presentation/proto"
)

// OutlineStates represents different interactive Outline states.
type OutlineStates struct {
	Initial Outline
	Hovered Outline
	Active  Outline
	Focused Outline
}

// ora converts an OutlineStates object into its protocol representation for serialization.
func (o OutlineStates) ora() proto.OutlineStates {
	return proto.OutlineStates{
		Initial: o.Initial.ora(),
		Hovered: o.Hovered.ora(),
		Active:  o.Active.ora(),
		Focused: o.Focused.ora(),
	}
}
