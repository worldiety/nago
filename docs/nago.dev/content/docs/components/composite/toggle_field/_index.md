---
# Content is auto generated
# Manual changes will be overwritten!
title: Toggle Field
---
This component combines a toggle with form-related elements such as
a label, supporting text, and error messages.

## Constructors
### ToggleField
A ToggleField aggregates a toggle together with form field typical labels, hints and error texts.

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` | AccessibilityLabel sets the accessibility label for screen readers. |
| `Border(border Border)` | Border sets the border styling of the toggle field. |
| `ErrorText(text string)` | ErrorText sets the error message displayed when validation fails. |
| `Frame(frame Frame)` | Frame sets the layout frame of the toggle field. |
| `InputValue(inputValue *core.State[bool])` | InputValue binds the toggle field to a reactive state for two-way binding. |
| `Padding(padding Padding)` | Padding sets the inner padding of the toggle field. |
| `SupportingText(text string)` | SupportingText sets optional supporting text displayed below the field. |
| `Visible(visible bool)` | Visible controls the visibility of the toggle field. |
| `WithFrame(fn func(Frame) Frame)` | WithFrame modifies the layout frame using the provided function. |
---

## Related
- [Border](../../utility/border/)
- [Frame](../../layout/frame/)
- [Padding](../../utility/padding/)

