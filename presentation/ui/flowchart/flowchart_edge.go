package flowchart

import (
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

type EdgeStyle uint64

const (
	EdgeStyleSolid  EdgeStyle = EdgeStyle(proto.FlowChartEdgeStyleSolid)
	EdgeStyleDashed EdgeStyle = EdgeStyle(proto.FlowChartEdgeStyleDashed)
	EdgeStyleDotted EdgeStyle = EdgeStyle(proto.FlowChartEdgeStyleDotted)
)

func (s EdgeStyle) ora() proto.FlowChartEdgeStyle {
	return proto.FlowChartEdgeStyle(s)
}

type EdgeMarker uint64

const (
	EdgeMarkerNone        EdgeMarker = EdgeMarker(proto.FlowChartEdgeMarkerNone)
	EdgeMarkerArrow       EdgeMarker = EdgeMarker(proto.FlowChartEdgeMarkerArrow)
	EdgeMarkerArrowClosed EdgeMarker = EdgeMarker(proto.FlowChartEdgeMarkerArrowClosed)
	EdgeMarkerCircle      EdgeMarker = EdgeMarker(proto.FlowChartEdgeMarkerCircle)
	EdgeMarkerDiamond     EdgeMarker = EdgeMarker(proto.FlowChartEdgeMarkerDiamond)
)

func (m EdgeMarker) ora() proto.FlowChartEdgeMarker {
	return proto.FlowChartEdgeMarker(m)
}

// Edge represents a connection between two nodes.
type Edge struct {
	ID           string
	SourceNodeID string
	TargetNodeID string
	Label        string
	Style        EdgeStyle
	Color        ui.Color
	Width        float64
	Animated     bool
	MarkerStart  EdgeMarker
	MarkerEnd    EdgeMarker
}

func (e Edge) render() proto.FlowChartEdge {
	return proto.FlowChartEdge{
		Id:           proto.Str(e.ID),
		SourceNodeId: proto.Str(e.SourceNodeID),
		TargetNodeId: proto.Str(e.TargetNodeID),
		Label:        proto.Str(e.Label),
		Style:        e.Style.ora(),
		Color:        proto.Color(e.Color),
		Width:        proto.Float(e.Width),
		Animated:     proto.Bool(e.Animated),
		MarkerStart:  e.MarkerStart.ora(),
		MarkerEnd:    e.MarkerEnd.ora(),
	}
}
