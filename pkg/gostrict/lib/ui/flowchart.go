// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/flowchart"
)

// The names of this file follow the package flowchart of nago: FlowChart and the names that start with FlowChart keep
// their name, every other name gets the prefix Flowchart.

// FlowChartLayout is the orientation of the automatic layout, like flowchart.FlowChartLayout of nago.
type FlowChartLayout int32

const (
	FlowChartLayoutHorizontal FlowChartLayout = 1
	FlowChartLayoutVertical   FlowChartLayout = 2
)

// FlowChartBackgroundGridStyle is the pattern of the background, like flowchart.FlowChartBackgroundGridStyle of nago.
type FlowChartBackgroundGridStyle int32

const (
	FlowChartBackgroundGridStyleDots  FlowChartBackgroundGridStyle = 0
	FlowChartBackgroundGridStyleLines FlowChartBackgroundGridStyle = 1
)

// FlowchartEdgeStyle is the stroke style of an edge, like flowchart.EdgeStyle of nago.
type FlowchartEdgeStyle uint32

const (
	FlowchartEdgeStyleSolid  FlowchartEdgeStyle = 0
	FlowchartEdgeStyleDashed FlowchartEdgeStyle = 1
	FlowchartEdgeStyleDotted FlowchartEdgeStyle = 2
)

// FlowchartEdgeMarker is the decoration at the start or end of an edge, like flowchart.EdgeMarker of nago.
type FlowchartEdgeMarker uint32

const (
	FlowchartEdgeMarkerNone  FlowchartEdgeMarker = 0
	FlowchartEdgeMarkerArrow FlowchartEdgeMarker = 1
)

// FlowchartNodeType is the role of a node, like flowchart.NodeType of nago.
type FlowchartNodeType uint32

const (
	FlowchartNodeTypeDefault FlowchartNodeType = 0
	FlowchartNodeTypeStart   FlowchartNodeType = 1
	FlowchartNodeTypeEnd     FlowchartNodeType = 2
)

// FlowchartNodeStyle is the predefined style of a node, like flowchart.NodeStyle of nago.
type FlowchartNodeStyle uint32

const (
	FlowchartNodeStyleDefault FlowchartNodeStyle = 0
	FlowchartNodeStyleNone    FlowchartNodeStyle = 1
)

// FlowChartToolbarPosition is the position of the toolbar, like flowchart.FlowChartToolbarPosition of nago.
type FlowChartToolbarPosition int32

const (
	FlowChartToolbarPositionTopRight     FlowChartToolbarPosition = 0
	FlowChartToolbarPositionCenterRight  FlowChartToolbarPosition = 1
	FlowChartToolbarPositionBottomRight  FlowChartToolbarPosition = 2
	FlowChartToolbarPositionBottomCenter FlowChartToolbarPosition = 3
	FlowChartToolbarPositionBottomLeft   FlowChartToolbarPosition = 4
	FlowChartToolbarPositionCenterLeft   FlowChartToolbarPosition = 5
	FlowChartToolbarPositionTopLeft      FlowChartToolbarPosition = 6
	FlowChartToolbarPositionTopCenter    FlowChartToolbarPosition = 7
)

// FlowChartToolbarOrientation is the orientation of the toolbar, like flowchart.FlowChartToolbarOrientation of nago.
type FlowChartToolbarOrientation int32

const (
	FlowChartToolbarOrientationHorizontal FlowChartToolbarOrientation = 1
	FlowChartToolbarOrientationVertical   FlowChartToolbarOrientation = 2
)

// FlowChartToolbarAction is a predefined action of the toolbar, like flowchart.FlowChartToolbarAction of nago.
type FlowChartToolbarAction int32

const (
	FlowChartToolbarActionAutoLayout FlowChartToolbarAction = 1
)

// FlowchartPoint is a position within the chart, like flowchart.Point of nago.
type FlowchartPoint struct {
	X float64
	Y float64
}

// FlowchartNode is a node of the chart, like flowchart.Node of nago.
type FlowchartNode struct {
	ID       string
	Type     FlowchartNodeType
	Position FlowchartPoint
	Label    string
	Style    FlowchartNodeStyle
}

// FlowchartEdge connects two nodes, like flowchart.Edge of nago.
type FlowchartEdge struct {
	ID             string
	SourceNodeID   string
	TargetNodeID   string
	Label          string
	LabelBgColor   Color
	LabelTextColor Color
	Style          FlowchartEdgeStyle
	Color          Color
	Width          float64
	Animated       bool
	MarkerStart    FlowchartEdgeMarker
	MarkerEnd      FlowchartEdgeMarker
}

// FlowchartBackground is the background of the chart, like flowchart.Background of nago.
type FlowchartBackground struct {
	Color     Color
	GridColor Color
	GridStyle FlowChartBackgroundGridStyle
	GridGap   uint32
}

// FlowchartToolbar is the toolbar of the chart, like flowchart.Toolbar of nago.
type FlowchartToolbar struct {
	Position    FlowChartToolbarPosition
	Orientation FlowChartToolbarOrientation
	Actions     list.List[FlowChartToolbarAction]
}

// FlowchartModel is the content of the chart, like flowchart.Model of nago. Create the lists with list.Of.
type FlowchartModel struct {
	Nodes list.List[FlowchartNode]
	Edges list.List[FlowchartEdge]
}

// FlowChartActionData describes the last action of the user in the chart, like flowchart.FlowChartActionData of nago.
type FlowChartActionData struct {
	Node          FlowchartNode
	Edge          FlowchartEdge
	ViewX         float64
	ViewY         float64
	PaneX         float64
	PaneY         float64
	SelectedNodes list.List[string]
	SelectedEdges list.List[string]
	RightClick    bool
}

// FlowchartModelState is a model of the window, which it keeps between two renders, see TFlowChart.InputValue.
type FlowchartModelState struct{ ref js.Ref }

// FlowchartActionState is the last action of the user in the chart of the window, see TFlowChart.ActionValue.
type FlowchartActionState struct{ ref js.Ref }

// TFlowchartMenu is the context menu of the chart, like flowchart.Menu of nago, see FlowchartMenu.
type TFlowchartMenu struct{ ref js.Ref }

// TFlowchartCustomContent replaces the content of a node, like flowchart.CustomContent of nago, see
// FlowchartCustomContent.
type TFlowchartCustomContent struct{ ref js.Ref }

// TFlowChart shows a chart of nodes and edges, which the user edits, like flowchart.TFlowChart of nago.
type TFlowChart struct{ ref js.Ref }

func (TFlowChart) isView() {}

// WithNodes returns the model with the nodes, like flowchart.Model.WithNodes of nago.
//
//gostrict:host
func (m FlowchartModel) WithNodes(nodes ...FlowchartNode) FlowchartModel {
	return flowchartModelFrom(m.nago().WithNodes(flowchartNodes(nodes)...))
}

// AppendNodes returns the model with the additional nodes, like flowchart.Model.AppendNodes of nago.
//
//gostrict:host
func (m FlowchartModel) AppendNodes(nodes ...FlowchartNode) FlowchartModel {
	return flowchartModelFrom(m.nago().AppendNodes(flowchartNodes(nodes)...))
}

// WithEdges returns the model with the edges, like flowchart.Model.WithEdges of nago.
//
//gostrict:host
func (m FlowchartModel) WithEdges(edges ...FlowchartEdge) FlowchartModel {
	return flowchartModelFrom(m.nago().WithEdges(flowchartEdges(edges)...))
}

// AppendEdges returns the model with the additional edges, like flowchart.Model.AppendEdges of nago.
//
//gostrict:host
func (m FlowchartModel) AppendEdges(edges ...FlowchartEdge) FlowchartModel {
	return flowchartModelFrom(m.nago().AppendEdges(flowchartEdges(edges)...))
}

// FlowchartModelStateOf returns the state with the id of the window, which has an empty model at first.
//
//gostrict:host
func FlowchartModelStateOf(wnd Window, id string) FlowchartModelState {
	return FlowchartModelState{ref: wrap(core.StateOf[flowchart.Model](wnd.nago(), id))}
}

// Get returns the model.
//
//gostrict:host
func (s FlowchartModelState) Get() FlowchartModel {
	return flowchartModelFrom(s.nago().Get())
}

// Set sets the model and renders the window again.
//
//gostrict:host
func (s FlowchartModelState) Set(v FlowchartModel) {
	s.nago().Set(v.nago())
}

// Init sets the model of fn, if the state is new.
//
//gostrict:host
func (s FlowchartModelState) Init(fn func() FlowchartModel) FlowchartModelState {
	return FlowchartModelState{ref: wrap(s.nago().Init(func() flowchart.Model { return fn().nago() }))}
}

// Observe calls fn, when the user changes the model.
//
//gostrict:host
func (s FlowchartModelState) Observe(fn func(v FlowchartModel)) FlowchartModelState {
	return FlowchartModelState{ref: wrap(s.nago().Observe(func(v flowchart.Model) { fn(flowchartModelFrom(v)) }))}
}

// FlowchartActionStateOf returns the state with the id of the window, which has no action at first.
//
//gostrict:host
func FlowchartActionStateOf(wnd Window, id string) FlowchartActionState {
	return FlowchartActionState{ref: wrap(core.StateOf[flowchart.FlowChartActionData](wnd.nago(), id))}
}

// Get returns the last action.
//
//gostrict:host
func (s FlowchartActionState) Get() FlowChartActionData {
	return flowChartActionDataFrom(s.nago().Get())
}

// Set sets the action and renders the window again.
//
//gostrict:host
func (s FlowchartActionState) Set(v FlowChartActionData) {
	s.nago().Set(v.nago())
}

// Init sets the action of fn, if the state is new.
//
//gostrict:host
func (s FlowchartActionState) Init(fn func() FlowChartActionData) FlowchartActionState {
	return FlowchartActionState{ref: wrap(s.nago().Init(func() flowchart.FlowChartActionData { return fn().nago() }))}
}

// Observe calls fn, when the user acts in the chart.
//
//gostrict:host
func (s FlowchartActionState) Observe(fn func(v FlowChartActionData)) FlowchartActionState {
	return FlowchartActionState{ref: wrap(s.nago().Observe(func(v flowchart.FlowChartActionData) { fn(flowChartActionDataFrom(v)) }))}
}

// FlowchartMenu returns the context menu, which shows the content at the position, like flowchart.Menu of nago.
//
//gostrict:host
func FlowchartMenu(position FlowchartPoint, content View) TFlowchartMenu {
	return TFlowchartMenu{ref: wrap(flowchart.Menu{Position: position.nago(), Content: viewOf(content)})}
}

// FlowchartCustomContent returns the content, which replaces the content of the node, like flowchart.CustomContent of
// nago.
//
//gostrict:host
func FlowchartCustomContent(nodeID string, content View) TFlowchartCustomContent {
	return TFlowchartCustomContent{ref: wrap(flowchart.CustomContent{NodeID: nodeID, Content: viewOf(content)})}
}

// Menu sets the menu of the node.
//
//gostrict:host
func (c TFlowchartCustomContent) Menu(menu View) TFlowchartCustomContent {
	v := get[flowchart.CustomContent](c.ref)
	v.Menu = viewOf(menu)
	return TFlowchartCustomContent{ref: wrap(v)}
}

// FlowChart returns a chart of the model, like flowchart.FlowChart of nago.
//
//gostrict:host
func FlowChart(model FlowchartModel) TFlowChart {
	return TFlowChart{ref: wrap(flowchart.FlowChart(model.nago()))}
}

// Model sets the model.
//
//gostrict:host
func (c TFlowChart) Model(model FlowchartModel) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).Model(model.nago()))}
}

// InputValue binds the model to the state, so that the edits of the user change the state.
//
//gostrict:host
func (c TFlowChart) InputValue(input FlowchartModelState) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).InputValue(input.nago()))}
}

// ActionValue stores the last action of the user in the state.
//
//gostrict:host
func (c TFlowChart) ActionValue(state FlowchartActionState) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).ActionValue(state.nago()))}
}

// Frame sets the size.
//
//gostrict:host
func (c TFlowChart) Frame(frame Frame) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).Frame(frame.nago()))}
}

// WithFrame changes the size with fn, which gets the current one.
//
//gostrict:host
func (c TFlowChart) WithFrame(fn func(Frame) Frame) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }))}
}

// FullWidth sets the width to the full width of the parent.
//
//gostrict:host
func (c TFlowChart) FullWidth() TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).FullWidth())}
}

// Background sets the background.
//
//gostrict:host
func (c TFlowChart) Background(background FlowchartBackground) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).Background(background.nago()))}
}

// NodesDraggable lets the user drag the nodes.
//
//gostrict:host
func (c TFlowChart) NodesDraggable(val bool) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).NodesDraggable(val))}
}

// NodesConnectable lets the user connect the nodes.
//
//gostrict:host
func (c TFlowChart) NodesConnectable(val bool) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).NodesConnectable(val))}
}

// EdgesEditable lets the user edit the edges.
//
//gostrict:host
func (c TFlowChart) EdgesEditable(val bool) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).EdgesEditable(val))}
}

// ElementsSelectable lets the user select the nodes and edges.
//
//gostrict:host
func (c TFlowChart) ElementsSelectable(val bool) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).ElementsSelectable(val))}
}

// Layout sets the orientation of the layout.
//
//gostrict:host
func (c TFlowChart) Layout(layout FlowChartLayout) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).Layout(flowchart.FlowChartLayout(layout)))}
}

// CustomContents sets the contents, which replace the contents of nodes.
//
//gostrict:host
func (c TFlowChart) CustomContents(contents ...TFlowchartCustomContent) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).CustomContents(flowchartCustomContents(contents)))}
}

// AppendCustomContent adds a content, which replaces the content of a node.
//
//gostrict:host
func (c TFlowChart) AppendCustomContent(content TFlowchartCustomContent) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).AppendCustomContent(get[flowchart.CustomContent](content.ref)))}
}

// MinZoom sets the smallest zoom.
//
//gostrict:host
func (c TFlowChart) MinZoom(minZoom float64) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).MinZoom(minZoom))}
}

// MaxZoom sets the largest zoom.
//
//gostrict:host
func (c TFlowChart) MaxZoom(maxZoom float64) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).MaxZoom(maxZoom))}
}

// Toolbar sets the toolbar.
//
//gostrict:host
func (c TFlowChart) Toolbar(toolbar FlowchartToolbar) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).Toolbar(toolbar.nago()))}
}

// Menu sets the context menu.
//
//gostrict:host
func (c TFlowChart) Menu(menu TFlowchartMenu) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).Menu(get[flowchart.Menu](menu.ref)))}
}

// ReadOnly forbids any change by the user.
//
//gostrict:host
func (c TFlowChart) ReadOnly(val bool) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).ReadOnly(val))}
}

// Movable lets the user move the view, which is the default.
//
//gostrict:host
func (c TFlowChart) Movable(val bool) TFlowChart {
	return TFlowChart{ref: wrap(get[flowchart.TFlowChart](c.ref).Movable(val))}
}

// AutoLayout arranges the nodes automatically, like flowchart.TFlowChart.AutoLayout of nago, which returns nothing.
//
//gostrict:host
func (c TFlowChart) AutoLayout(wnd Window) {
	get[flowchart.TFlowChart](c.ref).AutoLayout(wnd.nago())
}
