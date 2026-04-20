---
# Content is auto generated
# Manual changes will be overwritten!
title: RadioButton Field
---
It combines a radio button with a label
The field can be bound to external state and visibility controls.

## Constructors
### RadioButtonField
RadioButtonField combines a RadioButton with a label

---
## Methods
| Method | Description |
|--------| ------------|
| `Disabled(disabled bool)` | Disabled disables the radio button when set to true, preventing user interaction. |
| `ID(id string)` |  |
| `InputChecked(input *core.State[bool])` | InputChecked binds the radio button to the given state, enabling two-way data binding so that the selected state is synchronized with external logic. |
| `Label(label string)` | Label sets the label of the radio button field |
| `Name(name string)` | Name assigns a name to the checkbox field, useful for autocomplete |
| `Visible(v bool)` | Visible controls the visibility of the radio button. Passing false will hide the component from the UI. |
---

