---
# Content is auto generated
# Manual changes will be overwritten!
title: Colored Text Pill
---
It displays a short text label inside a pill-shaped container,
styled with a background color, padding, and rounded borders. Pills are often used to represent tags, statuses, or categories.

## Constructors
### ColoredTextPill
ColoredTextPill creates a new pill with the given background color and text,
applying default padding and rounded borders.

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` | AccessibilityLabel sets a label used by screen readers for accessibility. |
| `Border(border ui.Border)` | Border sets the border style of the pill, such as radius or thickness. |
| `Frame(frame ui.Frame)` | Frame sets the layout frame of the pill, including size and alignment. |
| `Padding(padding ui.Padding)` | Padding sets the inner spacing around the pill's text. |
| `Visible(visible bool)` | Visible controls the visibility of the pill; setting false hides it. |
| `WithFrame(fn func(ui.Frame) ui.Frame)` | WithFrame applies a transformation function to the pill's frame and returns the updated component. |
---

