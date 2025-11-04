---
# Content is auto generated
# Manual changes will be overwritten!
title: Stepper
---
It visually represents a sequence of steps in a process, highlighting
completed, current, and upcoming steps with distinct colors and styles. Each step can display a label, and the layout adapts between simple
or full-sized step representations.

## Constructors
### Stepper

---
## Methods
| Method | Description |
|--------| ------------|
| `FullCircleSize(length ui.Length)` | FullCircleSize sets the diameter of the step circles in full layout mode. |
| `FullStepWidth(length ui.Length)` | FullStepWidth sets the width allocated to each step in full layout mode. |
| `Index(idx int)` |  |
| `StepText(pattern string)` | StepText sets a different localized and parameterized (simple) step text, like "Schritt %d von %d". An empty string will omit the step text entirely. |
| `Style(style Style)` |  |
| `renderFull(ctx core.RenderContext)` |  |
| `renderSimple(ctx core.RenderContext)` |  |
---

