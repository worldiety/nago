// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flowchart

import (
	"go.wdy.de/nago/presentation/proto"
)

type FlowChartToolbarPosition int

const (
	FlowChartToolbarPositionTopRight     FlowChartToolbarPosition = FlowChartToolbarPosition(proto.FlowChartToolbarPositionTopRight)
	FlowChartToolbarPositionCenterRight  FlowChartToolbarPosition = FlowChartToolbarPosition(proto.FlowChartToolbarPositionCenterRight)
	FlowChartToolbarPositionBottomRight  FlowChartToolbarPosition = FlowChartToolbarPosition(proto.FlowChartToolbarPositionBottomRight)
	FlowChartToolbarPositionBottomCenter FlowChartToolbarPosition = FlowChartToolbarPosition(proto.FlowChartToolbarPositionBottomCenter)
	FlowChartToolbarPositionBottomLeft   FlowChartToolbarPosition = FlowChartToolbarPosition(proto.FlowChartToolbarPositionBottomLeft)
	FlowChartToolbarPositionCenterLeft   FlowChartToolbarPosition = FlowChartToolbarPosition(proto.FlowChartToolbarPositionCenterLeft)
	FlowChartToolbarPositionTopLeft      FlowChartToolbarPosition = FlowChartToolbarPosition(proto.FlowChartToolbarPositionTopLeft)
	FlowChartToolbarPositionTopCenter    FlowChartToolbarPosition = FlowChartToolbarPosition(proto.FlowChartToolbarPositionTopCenter)
)

type FlowChartToolbarOrientation int

const (
	FlowChartToolbarOrientationHorizontal FlowChartToolbarOrientation = FlowChartToolbarOrientation(proto.Horizontal)
	FlowChartToolbarOrientationVertical   FlowChartToolbarOrientation = FlowChartToolbarOrientation(proto.Vertical)
)

type FlowChartToolbarAction int

const (
	FlowChartToolbarActionAutoLayout = FlowChartToolbarAction(proto.FlowChartToolbarActionAutoLayout)
)

type Toolbar struct {
	Position    FlowChartToolbarPosition
	Orientation FlowChartToolbarOrientation
	Actions     []FlowChartToolbarAction
}
