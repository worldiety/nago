// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package ui

import (
	"time"

	"go.wdy.de/nago/application/color"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/calendar"
	"go.wdy.de/nago/presentation/ui/flowchart"
)

// The conversions of batch t1, see native.go.

func (t Time) nago() time.Time {
	return time.UnixMilli(int64(t)).In(time.Local)
}

func timeFrom(t time.Time) Time {
	return Time(t.UnixMilli())
}

func (d Duration) nago() time.Duration {
	return time.Duration(float64(d) * float64(time.Millisecond))
}

func durationFrom(d time.Duration) Duration {
	return Duration(float64(d) / float64(time.Millisecond))
}

func (d Date) nago() xtime.Date {
	return xtime.Date{Day: int(d.Day), Month: time.Month(d.Month), Year: int(d.Year)}
}

func dateFrom(d xtime.Date) Date {
	return Date{Day: int32(d.Day), Month: int32(d.Month), Year: int32(d.Year)}
}

func (f TimeFrame) nago() xtime.TimeFrame {
	return xtime.TimeFrame{StartTime: xtime.UnixMilliseconds(f.StartTime), EndTime: xtime.UnixMilliseconds(f.EndTime), Timezone: xtime.Timezone(f.Timezone)}
}

func timeFrameFrom(f xtime.TimeFrame) TimeFrame {
	return TimeFrame{StartTime: Time(f.StartTime), EndTime: Time(f.EndTime), Timezone: string(f.Timezone)}
}

func (s DateState) nago() *core.State[xtime.Date] {
	return get[*core.State[xtime.Date]](s.ref)
}

func (s TimeFrameState) nago() *core.State[xtime.TimeFrame] {
	return get[*core.State[xtime.TimeFrame]](s.ref)
}

func (s DurationState) nago() *core.State[time.Duration] {
	return get[*core.State[time.Duration]](s.ref)
}

// slice returns the elements of the list.
func slice[T any](l list.List[T]) []T {
	var res []T
	for v := range l.Values() {
		res = append(res, v)
	}
	return res
}

// listOf returns a list of the elements.
func listOf[T any](s []T) list.List[T] {
	res := list.New[T]()
	res.Append(s...)
	return res
}

// mapSlice converts the elements of the list.
func mapSlice[T, R any](l list.List[T], fn func(T) R) []R {
	var res []R
	for v := range l.Values() {
		res = append(res, fn(v))
	}
	return res
}

// mapList converts the elements of the slice into a list.
func mapList[T, R any](s []T, fn func(T) R) list.List[R] {
	res := list.New[R]()
	for _, v := range s {
		res.Append(fn(v))
	}
	return res
}
func (c CalendarColors) nago() calendar.Colors {
	return calendar.Colors{
		Header:               nui.Color(c.Header),
		LaneBackground:       nui.Color(c.LaneBackground),
		PrePostBackground:    nui.Color(c.PrePostBackground),
		PrePostForeground:    nui.Color(c.PrePostForeground),
		EventBackground:      nui.Color(c.EventBackground),
		EventHoverBackground: nui.Color(c.EventHoverBackground),
		Text:                 nui.Color(c.Text),
		EventText:            nui.Color(c.EventText),
		Separator:            nui.Color(c.Separator),
	}
}

func calendarColorsFrom(v calendar.Colors) CalendarColors {
	return CalendarColors{
		Header:               Color(v.Header),
		LaneBackground:       Color(v.LaneBackground),
		PrePostBackground:    Color(v.PrePostBackground),
		PrePostForeground:    Color(v.PrePostForeground),
		EventBackground:      Color(v.EventBackground),
		EventHoverBackground: Color(v.EventHoverBackground),
		Text:                 Color(v.Text),
		EventText:            Color(v.EventText),
		Separator:            Color(v.Separator),
	}
}

func (c CalendarCategory) nago() calendar.Category {
	return calendar.Category{Color: nui.Color(c.Color), Label: c.Label}
}

func calendarCategoryFrom(v calendar.Category) CalendarCategory {
	return CalendarCategory{Color: Color(v.Color), Label: v.Label}
}

func (l CalendarLane) nago() calendar.Lane {
	return calendar.Lane{Label: l.Label}
}

func calendarLaneFrom(v calendar.Lane) CalendarLane {
	return CalendarLane{Label: v.Label}
}

func (c CalendarColumn) nago() calendar.Column {
	return calendar.Column{Label: c.Label}
}

func calendarColumnFrom(v calendar.Column) CalendarColumn {
	return CalendarColumn{Label: v.Label}
}

func (p CalendarPillar) nago() calendar.Pillar {
	return calendar.Pillar{Label: p.Label, Icon: core.SVG(p.Icon), Width: nui.Length(p.Width), Height: nui.Length(p.Height)}
}

func calendarPillarFrom(v calendar.Pillar) CalendarPillar {
	return CalendarPillar{Label: v.Label, Icon: SVG(v.Icon), Width: Length(v.Width), Height: Length(v.Height)}
}

func (c CalendarChip) nago() calendar.Chip {
	return calendar.Chip{
		Label:       c.Label,
		Icon:        core.SVG(c.Icon),
		FillColor:   color.Color(c.FillColor),
		StrokeColor: color.Color(c.StrokeColor),
		BgColor:     color.Color(c.BgColor),
		TextColor:   color.Color(c.TextColor),
		Alignment:   nui.Alignment(c.Alignment),
		FullWidth:   c.FullWidth,
	}
}

func calendarChipFrom(v calendar.Chip) CalendarChip {
	return CalendarChip{
		Label:       v.Label,
		Icon:        SVG(v.Icon),
		FillColor:   Color(v.FillColor),
		StrokeColor: Color(v.StrokeColor),
		BgColor:     Color(v.BgColor),
		TextColor:   Color(v.TextColor),
		Alignment:   Alignment(v.Alignment),
		FullWidth:   v.FullWidth,
	}
}

func (o CalendarOffset) nago() calendar.Offset {
	return calendar.Offset{Duration: time.Duration(o.Duration) * time.Millisecond, Icon: core.SVG(o.Icon), Label: o.Label}
}

func calendarOffsetFrom(v calendar.Offset) CalendarOffset {
	return CalendarOffset{Duration: int32(v.Duration / time.Millisecond), Icon: SVG(v.Icon), Label: v.Label}
}

func (e TCalendarEvent) with(fn func(v *calendar.Event)) TCalendarEvent {
	v := get[calendar.Event](e.ref)
	fn(&v)
	return TCalendarEvent{ref: wrap(v)}
}

func (p TCalendarViewPort) with(fn func(v *calendar.ViewPort)) TCalendarViewPort {
	v := get[calendar.ViewPort](p.ref)
	fn(&v)
	return TCalendarViewPort{ref: wrap(v)}
}

func calendarEvents(events []TCalendarEvent) []calendar.Event {
	res := make([]calendar.Event, 0, len(events))
	for _, e := range events {
		res = append(res, get[calendar.Event](e.ref))
	}
	return res
}

func calendarCategories(cats []CalendarCategory) []calendar.Category {
	res := make([]calendar.Category, 0, len(cats))
	for _, c := range cats {
		res = append(res, c.nago())
	}
	return res
}

func calendarChips(chips []CalendarChip) []calendar.Chip {
	res := make([]calendar.Chip, 0, len(chips))
	for _, c := range chips {
		res = append(res, c.nago())
	}
	return res
}

func calendarColumns(cols []CalendarColumn) []calendar.Column {
	res := make([]calendar.Column, 0, len(cols))
	for _, c := range cols {
		res = append(res, c.nago())
	}
	return res
}

func (p FlowchartPoint) nago() flowchart.Point {
	return flowchart.Point{X: p.X, Y: p.Y}
}

func flowchartPointFrom(p flowchart.Point) FlowchartPoint {
	return FlowchartPoint{X: p.X, Y: p.Y}
}

func (n FlowchartNode) nago() flowchart.Node {
	return flowchart.Node{ID: n.ID, Type: flowchart.NodeType(n.Type), Position: n.Position.nago(), Label: n.Label, Style: flowchart.NodeStyle(n.Style)}
}

func flowchartNodeFrom(n flowchart.Node) FlowchartNode {
	return FlowchartNode{ID: n.ID, Type: FlowchartNodeType(n.Type), Position: flowchartPointFrom(n.Position), Label: n.Label, Style: FlowchartNodeStyle(n.Style)}
}

func (e FlowchartEdge) nago() flowchart.Edge {
	return flowchart.Edge{
		ID: e.ID, SourceNodeID: e.SourceNodeID, TargetNodeID: e.TargetNodeID, Label: e.Label,
		LabelBgColor: nui.Color(e.LabelBgColor), LabelTextColor: nui.Color(e.LabelTextColor),
		Style: flowchart.EdgeStyle(e.Style), Color: nui.Color(e.Color), Width: e.Width, Animated: e.Animated,
		MarkerStart: flowchart.EdgeMarker(e.MarkerStart), MarkerEnd: flowchart.EdgeMarker(e.MarkerEnd),
	}
}

func flowchartEdgeFrom(e flowchart.Edge) FlowchartEdge {
	return FlowchartEdge{
		ID: e.ID, SourceNodeID: e.SourceNodeID, TargetNodeID: e.TargetNodeID, Label: e.Label,
		LabelBgColor: Color(e.LabelBgColor), LabelTextColor: Color(e.LabelTextColor),
		Style: FlowchartEdgeStyle(e.Style), Color: Color(e.Color), Width: e.Width, Animated: e.Animated,
		MarkerStart: FlowchartEdgeMarker(e.MarkerStart), MarkerEnd: FlowchartEdgeMarker(e.MarkerEnd),
	}
}

func (b FlowchartBackground) nago() flowchart.Background {
	return flowchart.Background{Color: nui.Color(b.Color), GridColor: nui.Color(b.GridColor), GridStyle: flowchart.FlowChartBackgroundGridStyle(b.GridStyle), GridGap: uint64(b.GridGap)}
}

func (t FlowchartToolbar) nago() flowchart.Toolbar {
	return flowchart.Toolbar{
		Position:    flowchart.FlowChartToolbarPosition(t.Position),
		Orientation: flowchart.FlowChartToolbarOrientation(t.Orientation),
		Actions: mapSlice(t.Actions, func(a FlowChartToolbarAction) flowchart.FlowChartToolbarAction {
			return flowchart.FlowChartToolbarAction(a)
		}),
	}
}

func (m FlowchartModel) nago() flowchart.Model {
	return flowchart.Model{Nodes: flowchartNodes(slice(m.Nodes)), Edges: flowchartEdges(slice(m.Edges))}
}

func flowchartModelFrom(m flowchart.Model) FlowchartModel {
	return FlowchartModel{Nodes: mapList(m.Nodes, flowchartNodeFrom), Edges: mapList(m.Edges, flowchartEdgeFrom)}
}

func (d FlowChartActionData) nago() flowchart.FlowChartActionData {
	return flowchart.FlowChartActionData{
		Node: d.Node.nago(), Edge: d.Edge.nago(), ViewX: d.ViewX, ViewY: d.ViewY, PaneX: d.PaneX, PaneY: d.PaneY,
		SelectedNodes: slice(d.SelectedNodes), SelectedEdges: slice(d.SelectedEdges), RightClick: d.RightClick,
	}
}

func flowChartActionDataFrom(d flowchart.FlowChartActionData) FlowChartActionData {
	return FlowChartActionData{
		Node: flowchartNodeFrom(d.Node), Edge: flowchartEdgeFrom(d.Edge), ViewX: d.ViewX, ViewY: d.ViewY, PaneX: d.PaneX, PaneY: d.PaneY,
		SelectedNodes: listOf(d.SelectedNodes), SelectedEdges: listOf(d.SelectedEdges), RightClick: d.RightClick,
	}
}

func flowchartNodes(nodes []FlowchartNode) []flowchart.Node {
	res := make([]flowchart.Node, 0, len(nodes))
	for _, n := range nodes {
		res = append(res, n.nago())
	}
	return res
}

func flowchartEdges(edges []FlowchartEdge) []flowchart.Edge {
	res := make([]flowchart.Edge, 0, len(edges))
	for _, e := range edges {
		res = append(res, e.nago())
	}
	return res
}

func flowchartCustomContents(contents []TFlowchartCustomContent) []flowchart.CustomContent {
	res := make([]flowchart.CustomContent, 0, len(contents))
	for _, c := range contents {
		res = append(res, get[flowchart.CustomContent](c.ref))
	}
	return res
}

func (s FlowchartModelState) nago() *core.State[flowchart.Model] {
	return get[*core.State[flowchart.Model]](s.ref)
}

func (s FlowchartActionState) nago() *core.State[flowchart.FlowChartActionData] {
	return get[*core.State[flowchart.FlowChartActionData]](s.ref)
}
