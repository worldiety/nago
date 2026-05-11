// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	_ "embed"
	"fmt"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/colorpicker"
	"go.wdy.de/nago/presentation/ui/flowchart"
	"go.wdy.de/nago/web/vuejs"
)

//go:embed joe_schmoe_green.svg
var JoeSchmoeGreen core.SVG

//go:embed joe_schmoe_red.svg
var JoeSchmoeRed core.SVG

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_98")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			colorState := core.StateOf[ui.Color](wnd, "colorState")
			actionState := core.StateOf[flowchart.FlowChartAction](wnd, "actionState")

			actionState.Observe(func(action flowchart.FlowChartAction) {
				fmt.Println("Latest action", action)
			})

			nodes := make([]flowchart.Node, 0)
			contents := make([]flowchart.CustomContent, 0)

			jbusseNode, jbusseContent := personNode("jochen-busse", "Jochen Busse", "Geschäftsführer", JoeSchmoeGreen, flowchart.Point{
				X: 100,
				Y: -50,
			}, flowchart.NodeTypeStart, flowchart.NodeStyleDefault)
			nodes = append(nodes, jbusseNode)
			contents = append(contents, jbusseContent)

			gkoesterNode, gkoesterContent := personNode("gabi-koester", "Gabi Köster", "Projektmanager", JoeSchmoeGreen, flowchart.Point{
				X: -100,
				Y: 100,
			}, flowchart.NodeTypeDefault, flowchart.NodeStyleDefault)
			nodes = append(nodes, gkoesterNode)
			contents = append(contents, gkoesterContent)

			kpohlNode, kpohlContent := personNode("kalle-pohl", "Kalle Pohl", "Projektmanager", JoeSchmoeGreen, flowchart.Point{
				X: 300,
				Y: 100,
			}, flowchart.NodeTypeDefault, flowchart.NodeStyleDefault)
			nodes = append(nodes, kpohlNode)
			contents = append(contents, kpohlContent)

			gcantzNode, gcantzContent := personNode("guido-cantz", "Guido Cantz", "Projektmanager", JoeSchmoeGreen, flowchart.Point{
				X: 100,
				Y: 100,
			}, flowchart.NodeTypeDefault, flowchart.NodeStyleDefault)
			nodes = append(nodes, gcantzNode)
			contents = append(contents, gcantzContent)

			bstelterNode, bstelterContent := personNode("bernd-stelter", "Bernd Stelter", "Softwareentwickler", JoeSchmoeRed, flowchart.Point{
				X: -100,
				Y: 250,
			}, flowchart.NodeTypeEnd, flowchart.NodeStyleDefault)
			nodes = append(nodes, bstelterNode)
			contents = append(contents, bstelterContent)

			gcantznichtNode, gcantznichtContent := personNode("guido-cantz-nicht", "Guido Cantz Nicht", "Praktikant", JoeSchmoeGreen, flowchart.Point{
				X: 100,
				Y: 250,
			}, flowchart.NodeTypeEnd, flowchart.NodeStyleDefault)
			nodes = append(nodes, gcantznichtNode)
			contents = append(contents, gcantznichtContent)

			colorNode, colorContent := colorNode(colorState, "color", flowchart.Point{
				X: 300,
				Y: 250,
			}, flowchart.NodeTypeDefault)
			nodes = append(nodes, colorNode)
			contents = append(contents, colorContent)

			edges := []flowchart.Edge{
				{
					ID:           "jbusse-gkoester",
					SourceNodeID: "jochen-busse",
					TargetNodeID: "gabi-koester",
					Animated:     true,
				},
				{
					ID:           "jbusse-kpohl",
					SourceNodeID: "jochen-busse",
					TargetNodeID: "kalle-pohl",
					Animated:     true,
				},
				{
					ID:           "jbusse-gcantz",
					SourceNodeID: "jochen-busse",
					TargetNodeID: "guido-cantz",
					Animated:     true,
				},
				{
					ID:           "gkoester-bstelter",
					SourceNodeID: "gabi-koester",
					TargetNodeID: "bernd-stelter",
					Label:        "hat entlassen",
					Width:        2,
					Style:        flowchart.EdgeStyleDashed,
					Color:        ui.ColorError,
					MarkerEnd:    flowchart.EdgeMarkerArrow,
				},
				{
					ID:           "gcantz-gcantznicht",
					SourceNodeID: "guido-cantz",
					TargetNodeID: "guido-cantz-nicht",
				},
				{
					ID:           "kpohl-color",
					SourceNodeID: "kalle-pohl",
					TargetNodeID: "color",
					Style:        flowchart.EdgeStyleDotted,
					Color:        ui.ColorSemanticGood,
				},
			}

			state := core.AutoState[flowchart.Model](wnd).Init(func() flowchart.Model {
				return flowchart.Model{
					Nodes: nodes,
					Edges: edges,
				}
			})

			return ui.Stack(
				flowchart.FlowChart(state.Get()).
					InputValue(state).
					ActionValue(actionState).
					NodesDraggable(true).
					NodesConnectable(true).
					ElementsSelectable(true).
					BackgroundColor(ui.M1).
					Frame(ui.Frame{}.MatchScreen()).
					Layout(flowchart.FlowChartLayoutVertical).
					CustomContents(contents).
					MaxZoom(1.5),
				ui.Stack(
					ui.PrimaryButton(func() {
						if wnd.Info().ColorScheme == core.Light {
							wnd.SetColorScheme(core.Dark)
						} else {
							wnd.SetColorScheme(core.Light)
						}
					}).Title("Toggle theme"),
				).Position(ui.Position{
					Type:  ui.PositionFixed,
					Left:  ui.L0,
					Top:   ui.L16,
					Right: ui.L0,
				}).Alignment(ui.Center),
				lastAction(actionState),
			).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}

func personNode(id, name, title string, icon core.SVG, position flowchart.Point, nodeType flowchart.NodeType, style flowchart.NodeStyle) (flowchart.Node, flowchart.CustomContent) {
	return flowchart.Node{
			ID:       id,
			Position: position,
			Label:    name,
			Type:     nodeType,
			Style:    style,
		},
		flowchart.CustomContent{
			NodeID: id,
			Content: ui.VStack(
				ui.ImageIcon(icon),
				ui.VStack(
					ui.Text(name).Font(ui.TitleLarge),
					ui.Text(title).Font(ui.TitleSmall),
				),
			).Gap(ui.L4).Padding(ui.Padding{}.All(ui.L8)),
		}
}

func colorNode(state *core.State[ui.Color], id string, position flowchart.Point, nodeType flowchart.NodeType) (flowchart.Node, flowchart.CustomContent) {
	return flowchart.Node{
			ID:       id,
			Position: position,
			Type:     nodeType,
			Style:    flowchart.NodeStyleNone,
		},
		flowchart.CustomContent{
			NodeID: id,
			Content: ui.VStack(
				colorpicker.PalettePicker("Farbe", colorpicker.DefaultPalette).Value(state.Get()).State(state),
			).Gap(ui.L4).BackgroundColor(state.Get().WithTransparency(50)).Padding(ui.Padding{}.All(ui.L8)).Border(ui.Border{}.Radius(ui.L8)),
		}
}

func lastAction(state *core.State[flowchart.FlowChartAction]) core.View {
	action := state.Get()

	return ui.Stack(
		ui.Grid(
			ui.GridCell(ui.Text("Node:")),
			ui.GridCell(ui.IfElse(len(action.Node.ID) > 0, ui.Text(fmt.Sprintf("%v", action.Node)), ui.Text("-"))),
			ui.GridCell(ui.Text("Edge:")),
			ui.GridCell(ui.IfElse(len(action.Edge.ID) > 0, ui.Text(fmt.Sprintf("%v", action.Edge)), ui.Text("-"))),
			ui.GridCell(ui.Text("Point:")),
			ui.GridCell(ui.Text(fmt.Sprintf("%d %d", int(action.ViewX), int(action.ViewY)))),
			ui.GridCell(ui.Text("Selected nodes:")),
			ui.GridCell(ui.Text(fmt.Sprintf("%v", action.SelectedNodes))),
			ui.GridCell(ui.Text("Selected edges:")),
			ui.GridCell(ui.Text(fmt.Sprintf("%v", action.SelectedEdges))),
		).
			Columns(2).
			RowGap(ui.L2).
			ColGap(ui.L8).
			Widths("auto", "auto"),
	).
		Position(ui.Position{
			Type:   ui.PositionAbsolute,
			Left:   ui.L0,
			Bottom: ui.L0,
		}).
		BackgroundColor(ui.ColorText.WithTransparency(90)).
		Font(ui.MonoSmall).
		Padding(ui.Padding{}.All(ui.L8))
}
