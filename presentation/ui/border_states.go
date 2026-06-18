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

// BorderStates represents different interactive Border states.
type BorderStates struct {
	Initial Border
	Hovered Border
	Active  Border
	Focused Border
}

// ora converts a BorderStates object into its protocol representation for serialization.
func (o BorderStates) ora() proto.BorderStates {
	return proto.BorderStates{
		Initial: o.Initial.ora(),
		Hovered: o.Hovered.ora(),
		Active:  o.Active.ora(),
		Focused: o.Focused.ora(),
	}
}
