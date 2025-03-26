// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import "go.wdy.de/nago/presentation/proto"

type Shadow struct {
	Color  Color
	Radius Length
	X      Length
	Y      Length
}

func (s Shadow) ora() proto.Shadow {
	return proto.Shadow{
		Color:  proto.Color(s.Color),
		Radius: proto.Length(s.Radius),
		X:      proto.Length(s.X),
		Y:      proto.Length(s.Y),
	}
}
