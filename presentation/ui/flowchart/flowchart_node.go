package flowchart

import (
	"go.wdy.de/nago/presentation/proto"
)

type NodeType uint64

const (
	NodeTypeDefault NodeType = NodeType(proto.FlowChartNodeTypeDefault)
	NodeTypeStart   NodeType = NodeType(proto.FlowChartNodeTypeStart)
	NodeTypeEnd     NodeType = NodeType(proto.FlowChartNodeTypeEnd)
)

func (t NodeType) ora() proto.FlowChartNodeType {
	return proto.FlowChartNodeType(t)
}

type NodeStyle uint64

const (
	NodeStyleDefault NodeStyle = NodeStyle(proto.FlowChartNodeStyleDefault)
	NodeStyleNone    NodeStyle = NodeStyle(proto.FlowChartNodeStyleNone)
)

func (s NodeStyle) ora() proto.FlowChartNodeStyle {
	return proto.FlowChartNodeStyle(s)
}

// Point describes a node position in the flowchart canvas coordinate system.
type Point struct {
	X float64
	Y float64
}

func (p Point) Ora() proto.FlowChartPoint {
	return proto.FlowChartPoint{
		X: proto.Float(p.X),
		Y: proto.Float(p.Y),
	}
}

// Node represents a single node in the flowchart model.
type Node struct {
	ID       string
	Type     NodeType
	Position Point
	Label    string
	Style    NodeStyle
}

func (n Node) render() proto.FlowChartNode {
	return proto.FlowChartNode{
		Id:       proto.Str(n.ID),
		Type:     n.Type.ora(),
		Position: n.Position.Ora(),
		Label:    proto.Str(n.Label),
		Style:    n.Style.ora(),
	}
}
