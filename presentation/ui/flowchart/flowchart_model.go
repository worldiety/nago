// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flowchart

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
