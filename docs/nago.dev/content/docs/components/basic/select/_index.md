---
# Content is auto generated
# Manual changes will be overwritten!
title: Select
---
It allows users to select one option of a given set of options,.

## Constructors
### Dropdown
Dropdown represents a user interface element which lets the user select one option from a list.

### FromSlice
FromSlice mimics the default signature of the [picker.Picker] factory so that TDropdown can be used as a drop-in
replacement for single selection.

---
## Methods
| Method | Description |
|--------| ------------|
| `Disabled(disabled bool)` | Disabled enables or disables user interaction with the select. |
| `ErrorText(text string)` | ErrorText sets the error text displayed below the select. |
| `Frame(frame ui.Frame)` | Frame sets the layout frame of the field (size, width, height, etc.). |
| `ID(id string)` | ID assigns a unique identifier to the select, useful for testing or referencing. |
| `InputValue(input *core.State[ID])` | InputValue binds the select to an external value state, allowing it to be controlled from outside the component. |
| `Label(label string)` | Label sets the label displayed above or inside the select. |
| `Leading(v core.View)` | Leading sets a leading view for the select. This view is displayed at the start of the select, e.g., an icon. |
| `Options(options []Option[ID])` | Options sets the list of options available for selection. |
| `Style(s ui.TextFieldStyle)` | Style sets the visual style of the select. |
| `SupportingText(text string)` | SupportingText sets the supporting text displayed below the select. |
| `Value(value ID)` | Value sets the initial value of the select. |
---

