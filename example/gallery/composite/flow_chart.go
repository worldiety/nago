// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/flowchart"
)

func init() {
	app.Register("flow-chart", func(wnd core.Window) core.View {
		model := flowchart.Model{
			Nodes: []flowchart.Node{
				{ID: "order", Label: "Order received", Type: flowchart.NodeTypeStart, Position: flowchart.Point{X: 200, Y: 0}},
				{ID: "check", Label: "Check stock", Position: flowchart.Point{X: 200, Y: 120}},
				{ID: "ship", Label: "Ship parcel", Position: flowchart.Point{X: 50, Y: 240}},
				{ID: "backorder", Label: "Backorder", Position: flowchart.Point{X: 350, Y: 240}},
				{ID: "done", Label: "Done", Type: flowchart.NodeTypeEnd, Position: flowchart.Point{X: 200, Y: 360}},
			},
			Edges: []flowchart.Edge{
				{ID: "e1", SourceNodeID: "order", TargetNodeID: "check"},
				{ID: "e2", SourceNodeID: "check", TargetNodeID: "ship", Label: "in stock"},
				{ID: "e3", SourceNodeID: "check", TargetNodeID: "backorder", Label: "missing", Animated: true},
				{ID: "e4", SourceNodeID: "ship", TargetNodeID: "done"},
				{ID: "e5", SourceNodeID: "backorder", TargetNodeID: "done"},
			},
		}

		return flowchart.FlowChart(model).
			Layout(flowchart.FlowChartLayoutVertical).
			NodesDraggable(true).
			Background(flowchart.Background{GridGap: 16}).
			Frame(Frame{Width: L880, Height: L480})
	})
}
