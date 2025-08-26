// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import "go.wdy.de/nago/presentation/proto"

// Shadow is a util component (Shadow).
// It defines shadow styling properties that can be applied to components.
// A shadow consists of a color, blur radius, and X/Y offsets.
type Shadow struct {
	Color  Color  // shadow color
	Radius Length // blur radius of the shadow
	X      Length // horizontal offset
	Y      Length // vertical offset
}

// ora converts the Shadow into its protocol representation.
func (s Shadow) ora() proto.Shadow {
	return proto.Shadow{
		Color:  proto.Color(s.Color),
		Radius: proto.Length(s.Radius),
		X:      proto.Length(s.X),
		Y:      proto.Length(s.Y),
	}
}
