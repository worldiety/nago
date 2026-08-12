---
# Content is auto generated
# Manual changes will be overwritten!
title: Button
---
It represents a clickable UI element with optional label, content, and supporting text. The button can be enabled/disabled, styled with frame, padding, and border,
and may include accessibility attributes for better usability.

## Constructors
### Button

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` | AccessibilityLabel sets screen reader label. |
| `Border(border ui.Border)` | Border defines the border styling (color, width, radius) of the button. |
| `Content(content core.View)` |  |
| `Dialog(dialog core.View)` | Dialog is just inserted into the rendered container as well and is not intended for a regular visible view. This is pure optional and for sure you can insert the dialog anywhere else and just ignore this. However, putting a normal view here, will break the component. |
| `Frame(frame ui.Frame)` | Frame sets size and position. |
| `Padding(padding ui.Padding)` | Padding sets inner spacing. |
| `Visible(visible bool)` | Visible toggles visibility. |
| `WithFrame(fn func(ui.Frame) ui.Frame)` | WithFrame modifies the current frame. |
---

