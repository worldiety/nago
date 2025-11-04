---
# Content is auto generated
# Manual changes will be overwritten!
title: Checkbox Field
---
It combines a checkbox with a label, supporting text, and optional
error messages. The field can be bound to external state and styled
with padding, frame, and border. It also supports accessibility,
keyboard options, and visibility controls.

## Constructors
### CheckboxField
A CheckboxField aggregates a checkbox together with form field typical labels, hints and error texts.

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` | AccessibilityLabel sets the label used for accessibility purposes. |
| `Border(border Border)` | Border sets the border styling of the checkbox field. |
| `Disabled(b bool)` | Disabled enables or disables user interaction with the checkbox field. |
| `Enabled(b bool)` | Enabled sets whether the checkbox field is interactive. Equivalent to Disabled(!b). |
| `ErrorText(text string)` | ErrorText sets the validation or error message displayed below the field. |
| `Frame(frame Frame)` | Frame sets the layout frame of the checkbox field, including size and positioning. |
| `ID(id string)` | ID assigns a unique identifier to the checkbox field, useful for testing or referencing. |
| `InputValue(inputValue *core.State[bool])` | InputValue binds the checkbox field to an external boolean state. |
| `Padding(padding Padding)` | Padding sets the inner spacing around the checkbox field. |
| `SupportingText(text string)` | SupportingText sets helper or secondary text shown below the label. |
| `Visible(visible bool)` | Visible controls the visibility of the checkbox field; setting false hides it. |
| `WithFrame(fn func(Frame) Frame)` | WithFrame applies a transformation function to the field's frame and returns the updated component. |
---

## Related
- [Border](../../utility/border/)
- [Frame](../../layout/frame/)
- [Padding](../../utility/padding/)

