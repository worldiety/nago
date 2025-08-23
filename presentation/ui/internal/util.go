// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package internal

import (
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

func FrameToOra(f ui.Frame) proto.Frame {
	return proto.Frame{
		MinWidth:  proto.Length(f.MinWidth),
		MaxWidth:  proto.Length(f.MaxWidth),
		MinHeight: proto.Length(f.MinHeight),
		MaxHeight: proto.Length(f.MaxHeight),
		Width:     proto.Length(f.Width),
		Height:    proto.Length(f.Height),
	}
}
