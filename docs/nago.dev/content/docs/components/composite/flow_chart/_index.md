---
# Content is auto generated
# Manual changes will be overwritten!
title: Flow Chart
---
It renders a node-edge diagram defined by a [Model].

## Constructors
### FlowChart
FlowChart creates a new flowchart component for the given model.

---
## Methods
| Method | Description |
|--------| ------------|
| `AppendCustomContent(content CustomContent)` |  |
| `BackgroundColor(color ui.Color)` |  |
| `CustomContents(contents []CustomContent)` |  |
| `EdgesEditable(val bool)` |  |
| `ElementsSelectable(val bool)` |  |
| `Frame(frame ui.Frame)` |  |
| `FullWidth()` |  |
| `InputValue(input *core.State[Model])` | InputValue binds the flowchart to a stateful model.  At the moment this acts as the render source of truth. The component reads the model from the state during rendering. A dedicated frontend write-back requires proto support for an InputValue pointer on proto.FlowChart. |
| `Layout(layout FlowChartLayout)` |  |
| `MaxZoom(maxZoom float64)` |  |
| `MinZoom(minZoom float64)` |  |
| `Model(model Model)` | Model sets the static flowchart model. |
| `NodesConnectable(val bool)` |  |
| `NodesDraggable(val bool)` |  |
| `WithFrame(fn func(ui.Frame) ui.Frame)` |  |
---

