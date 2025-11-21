---
# Content is auto generated
# Manual changes will be overwritten!
title: Radio Button
---
It represents a selectable option in a group where only one element can be active at a time. Radio buttons are typically used in forms or settings where the user must pick exactly one choice.

## Constructors
### RadioButton
RadioButton represents a user interface element which spans a visible area to click or tap from the user.
Use it for controls, which do not cause an immediate effect and only one element can be picked at a time.
See also [Toggle], [Checkbox] and [Select].

---
## Methods
| Method | Description |
|--------| ------------|
| `Disabled(disabled bool)` | Disabled disables the radio button when set to true, preventing user interaction. |
| `ID(id string)` |  |
| `InputChecked(input *core.State[bool])` | InputChecked binds the radio button to the given state, enabling two-way data binding so that the selected state is synchronized with external logic. |
| `Visible(v bool)` | Visible controls the visibility of the radio button. Passing false will hide the component from the UI. |
---

