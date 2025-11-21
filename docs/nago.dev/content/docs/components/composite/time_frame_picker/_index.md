---
# Content is auto generated
# Manual changes will be overwritten!
title: Time Frame Picker
---
It allows users to pick a date and a start/end time, optionally binding
the result to an external state. The picker supports different formats,
validation messages, and can be configured with a specific time zone.

## Constructors
### Picker
Picker renders a xtime.TimeFrame picker to select at least a day and a start and end time (inclusive).

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` | AccessibilityLabel sets a label used by screen readers for accessibility. (currently not implemented) |
| `Border(border ui.Border)` | Border sets the border style of the picker. (currently not implemented) |
| `Disabled(disabled bool)` | Disabled enables or disables user interaction with the picker. |
| `ErrorText(text string)` | ErrorText sets the validation or error message displayed below the picker. |
| `Format(format PickerFormat)` | Format sets the picker format, which controls its display and interaction style. |
| `Frame(frame ui.Frame)` | Frame sets the layout frame of the picker, including size and positioning. |
| `Padding(padding ui.Padding)` | Padding sets the inner spacing around the picker content. (currently not implemented) |
| `SupportingText(text string)` | SupportingText sets helper or secondary text displayed below the picker label. |
| `Title(title string)` | Title sets the title of the picker, typically shown in dialogs. |
| `Visible(visible bool)` | Visible controls the visibility of the picker; setting false hides it. (currently not implemented) |
| `WithFrame(fn func(ui.Frame) ui.Frame)` | WithFrame applies a transformation function to the picker's frame and returns the updated component. |
---

