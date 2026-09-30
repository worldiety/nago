// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagotest

import (
	"encoding/json"
	"time"

	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui/flowchart"
)

// FlowChartAction delivers an action of the user to the single selected flow chart, like a click on a node, an
// edge or the pane. The frontend reports these through the state bound by [flowchart.TFlowChart.ActionValue].
// The window settles afterward.
func (w *Window) FlowChartAction(s Selection, action flowchart.FlowChartActionData) {
	w.t.Helper()

	start := time.Now()
	buf, err := json.Marshal(action)
	if err != nil {
		w.t.Fatalf("nagotest: cannot encode flow chart action: %v", err)
	}

	w.act(s, func(s Selection) proto.NagoEvent {
		n := s.Node()
		chart, ok := n.Component.(*proto.FlowChart)
		if !ok {
			w.t.Fatalf("nagotest: %s: %T is not a flow chart", s.matcher, n.Component)
		}

		if chart.ActionValue == 0 {
			w.t.Fatalf("nagotest: %s: the flow chart has no action binding", s.matcher)
		}

		w.assertInteractive(s, n)
		return &proto.UpdateStateValueRequested{StatePointer: chart.ActionValue, Value: proto.Str(buf), RID: w.nextRID()}
	})
	w.observe("flowchart", s.matcher.desc, start)
}

// ClickFlowChartNode clicks the node with the given ID within the single selected flow chart, which also selects
// it, see [Window.FlowChartAction].
func (w *Window) ClickFlowChartNode(s Selection, nodeID string) {
	w.t.Helper()

	w.FlowChartAction(s, flowchart.FlowChartActionData{
		Node:          flowchart.Node{ID: nodeID},
		SelectedNodes: []string{nodeID},
	})
}

// ClickFlowChartEdge clicks the edge with the given ID within the single selected flow chart, which also selects
// it, see [Window.FlowChartAction].
func (w *Window) ClickFlowChartEdge(s Selection, edgeID string) {
	w.t.Helper()

	w.FlowChartAction(s, flowchart.FlowChartActionData{
		Edge:          flowchart.Edge{ID: edgeID},
		SelectedEdges: []string{edgeID},
	})
}
