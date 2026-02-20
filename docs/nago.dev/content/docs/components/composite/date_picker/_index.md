---
# Content is auto generated
# Manual changes will be overwritten!
title: Date Picker
---
It allows users to select either a single date or a date range,
depending on its style. The component supports external state
bindings, validation messages, and layout configuration.

## Constructors
### RangeDatePicker
RangeDatePicker creates a date picker configured for selecting a date range,
binding start and end values to their respective states.

### SingleDatePicker
SingleDatePicker creates a date picker configured for selecting a single date,
binding the given value and optional state.

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` | AccessibilityLabel sets a label used by screen readers for accessibility. |
| `Border(border Border)` | Border sets the border styling of the date picker. |
| `Disabled(disabled bool)` | Disabled enables or disables user interaction with the date picker. |
| `DoubleMode(doubleMode bool)` | DoubleMode enables double-month mode for range pickers. |
| `ErrorText(text string)` | ErrorText sets the validation or error message displayed below the picker. |
| `Frame(frame Frame)` | Frame sets the layout frame of the date picker, including size and positioning. |
| `Padding(padding Padding)` | Padding sets the inner spacing around the date picker content. |
| `SupportingText(text string)` | SupportingText sets helper or secondary text displayed below the picker label. |
| `Visible(visible bool)` | Visible controls the visibility of the date picker; setting false hides it. |
| `WithFrame(fn func(Frame) Frame)` | WithFrame applies a transformation function to the picker's frame and returns the updated component. |
---

## Related
- [Border](../../utility/border/)
- [Frame](../../layout/frame/)
- [Padding](../../utility/padding/)

