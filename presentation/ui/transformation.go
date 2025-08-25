// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import "go.wdy.de/nago/presentation/proto"

type Transformation struct {
	TranslateX Length
	TranslateY Length
	TranslateZ Length
	ScaleX     float64
	ScaleY     float64
	ScaleZ     float64
	// RotateZ defines rotation in degree
	RotateZ float64
}

func (t Transformation) ora() proto.Transformation {
	return proto.Transformation{
		TranslateX: proto.Length(t.TranslateX),
		TranslateY: proto.Length(t.TranslateY),
		TranslateZ: proto.Length(t.TranslateZ),
		ScaleX:     proto.Float(t.ScaleX),
		ScaleY:     proto.Float(t.ScaleY),
		ScaleZ:     proto.Float(t.ScaleZ),
		RotateZ:    proto.Float(t.RotateZ),
	}
}
