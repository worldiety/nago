// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flowchart

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

type FlowChartLayout int

const (
	FlowChartLayoutHorizontal = FlowChartLayout(proto.Horizontal)
	FlowChartLayoutVertical   = FlowChartLayout(proto.Vertical)
)

// Model contains the full declarative flowchart data.
//
// It can be used as a plain immutable value or be stored inside a [core.State]
// and bound via [TFlowChart.InputValue]. Because the current FlowChart proto type
// has no dedicated InputValue pointer yet, the state is currently used as the
// render source of truth (one-way binding).
type Model struct {
	Nodes []Node
	Edges []Edge
}

func (m Model) WithNodes(nodes ...Node) Model {
	m.Nodes = nodes
	return m
}

func (m Model) AppendNodes(nodes ...Node) Model {
	m.Nodes = append(m.Nodes, nodes...)
	return m
}

func (m Model) WithEdges(edges ...Edge) Model {
	m.Edges = edges
	return m
}

func (m Model) AppendEdges(edges ...Edge) Model {
	m.Edges = append(m.Edges, edges...)
	return m
}

type CustomContent struct {
	NodeID  string
	Content core.View
}

func (c CustomContent) render(ctx core.RenderContext) proto.FlowChartCustomContent {
	return proto.FlowChartCustomContent{
		NodeId:  proto.Str(c.NodeID),
		Content: c.Content.Render(ctx),
	}
}

type FlowChartBackgroundGridStyle int

const (
	FlowChartBackgroundGridStyleDots  = FlowChartBackgroundGridStyle(proto.FlowChartBackgroundGridStyleDots)
	FlowChartBackgroundGridStyleLines = FlowChartBackgroundGridStyle(proto.FlowChartBackgroundGridStyleLines)
)

type Background struct {
	Color     ui.Color
	GridColor ui.Color
	GridStyle FlowChartBackgroundGridStyle
	GridGap   uint64
}

type FlowChartActionData struct {
	Node          Node
	Edge          Edge
	ViewX         float64
	ViewY         float64
	PaneX         float64
	PaneY         float64
	SelectedNodes []string
	SelectedEdges []string
}

// TFlowChart is a composite component (Flow Chart).
// It renders a node-edge diagram defined by a [Model].
type TFlowChart struct {
	model              Model
	inputValue         *core.State[Model]
	actionValue        *core.State[FlowChartActionData]
	frame              ui.Frame
	background         Background
	nodesDraggable     bool
	nodesConnectable   bool
	edgesEditable      bool
	elementsSelectable bool
	layout             FlowChartLayout
	customContents     []CustomContent
	minZoom            float64
	maxZoom            float64
}

// FlowChart creates a new flowchart component for the given model.
func FlowChart(model Model) TFlowChart {
	return TFlowChart{model: model}
}

// Model sets the static flowchart model.
func (c TFlowChart) Model(model Model) TFlowChart {
	c.model = model
	return c
}

// InputValue binds the flowchart to a stateful model.
//
// At the moment this acts as the render source of truth. The component reads the
// model from the state during rendering. A dedicated frontend write-back requires
// proto support for an InputValue pointer on proto.FlowChart.
func (c TFlowChart) InputValue(input *core.State[Model]) TFlowChart {
	c.inputValue = input
	return c
}

func (c TFlowChart) ActionValue(state *core.State[FlowChartActionData]) TFlowChart {
	c.actionValue = state
	return c
}

func (c TFlowChart) Frame(frame ui.Frame) TFlowChart {
	c.frame = frame
	return c
}

func (c TFlowChart) WithFrame(fn func(ui.Frame) ui.Frame) TFlowChart {
	c.frame = fn(c.frame)
	return c
}

func (c TFlowChart) FullWidth() TFlowChart {
	c.frame = c.frame.FullWidth()
	return c
}

func (c TFlowChart) Background(background Background) TFlowChart {
	c.background = background
	return c
}

func (c TFlowChart) NodesDraggable(val bool) TFlowChart {
	c.nodesDraggable = val
	return c
}

func (c TFlowChart) NodesConnectable(val bool) TFlowChart {
	c.nodesConnectable = val
	return c
}

func (c TFlowChart) EdgesEditable(val bool) TFlowChart {
	c.edgesEditable = val
	return c
}

func (c TFlowChart) ElementsSelectable(val bool) TFlowChart {
	c.elementsSelectable = val
	return c
}

func (c TFlowChart) Layout(layout FlowChartLayout) TFlowChart {
	c.layout = layout
	return c
}

func (c TFlowChart) CustomContents(contents []CustomContent) TFlowChart {
	c.customContents = contents
	return c
}

func (c TFlowChart) AppendCustomContent(content CustomContent) TFlowChart {
	c.customContents = append(c.customContents, content)
	return c
}

func (c TFlowChart) MinZoom(minZoom float64) TFlowChart {
	c.minZoom = minZoom
	return c
}

func (c TFlowChart) MaxZoom(maxZoom float64) TFlowChart {
	c.maxZoom = maxZoom
	return c
}

func (c TFlowChart) Render(ctx core.RenderContext) core.RenderNode {
	m := c.model

	res := proto.FlowChart{
		InputValue:  c.inputValue.Ptr(),
		ActionValue: c.actionValue.Ptr(),
		Value: proto.FlowChartModel{
			Nodes: make(proto.FlowChartNodes, 0),
			Edges: make(proto.FlowChartEdges, 0),
		},
		Frame: frameToOra(c.frame),
		Background: proto.FlowChartBackground{
			Color:     proto.Color(c.background.Color),
			GridColor: proto.Color(c.background.GridColor),
			GridStyle: proto.FlowChartBackgroundGridStyle(c.background.GridStyle),
			GridGap:   proto.Uint(c.background.GridGap),
		},
		NodesDraggable:     proto.Bool(c.nodesDraggable),
		NodesConnectable:   proto.Bool(c.nodesConnectable),
		EdgesEditable:      proto.Bool(c.edgesEditable),
		ElementsSelectable: proto.Bool(c.elementsSelectable),
		Orientation:        proto.Orientation(c.layout),
		CustomContents:     make(proto.FlowChartCustomContents, 0),
		MinZoom:            proto.Float(c.minZoom),
		MaxZoom:            proto.Float(c.maxZoom),
	}

	for _, node := range m.Nodes {
		res.Value.Nodes = append(res.Value.Nodes, node.render())
	}

	for _, edge := range m.Edges {
		res.Value.Edges = append(res.Value.Edges, edge.render())
	}

	for _, content := range c.customContents {
		res.CustomContents = append(res.CustomContents, content.render(ctx))
	}

	return &res
}

func frameToOra(frame ui.Frame) proto.Frame {
	return proto.Frame{
		MinWidth:  proto.Length(frame.MinWidth),
		MaxWidth:  proto.Length(frame.MaxWidth),
		MinHeight: proto.Length(frame.MinHeight),
		MaxHeight: proto.Length(frame.MaxHeight),
		Width:     proto.Length(frame.Width),
		Height:    proto.Length(frame.Height),
	}
}
