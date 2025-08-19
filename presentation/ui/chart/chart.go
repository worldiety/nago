// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package chart

import (
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// Chart defines common chart metadata such as labels, colors, frame, and axis titles,
// including options like download availability and a fallback message when no data exists.
type Chart struct {
	Labels        []string
	Colors        []ui.Color
	Frame         ui.Frame
	Downloadable  bool
	NoDataMessage string
	XAxisTitle    string
	YAxisTitle    string
}

// Ora converts Chart into its proto.Chart representation, mapping labels, colors,
// frame, download flag, no-data message, and axis titles.
func (c Chart) Ora() proto.Chart {
	protoColors := make([]proto.Color, len(c.Colors))
	for i, color := range c.Colors {
		protoColors[i] = proto.Color(color)
	}
	labels := make([]proto.Str, len(c.Labels))
	for i, label := range c.Labels {
		labels[i] = proto.Str(label)
	}

	return proto.Chart{
		Labels:        labels,
		Colors:        protoColors,
		Frame:         c.ora(),
		Downloadable:  proto.Bool(c.Downloadable),
		NoDataMessage: proto.Str(c.NoDataMessage),
		XAxisTitle:    proto.Str(c.XAxisTitle),
		YAxisTitle:    proto.Str(c.YAxisTitle),
	}
}

// ora maps the UI frame to its proto.Frame equivalent (size and constraints).
func (c Chart) ora() proto.Frame {
	return proto.Frame{
		MinWidth:  proto.Length(c.Frame.MinWidth),
		MaxWidth:  proto.Length(c.Frame.MaxWidth),
		MinHeight: proto.Length(c.Frame.MinHeight),
		MaxHeight: proto.Length(c.Frame.MaxHeight),
		Width:     proto.Length(c.Frame.Width),
		Height:    proto.Length(c.Frame.Height),
	}
}
