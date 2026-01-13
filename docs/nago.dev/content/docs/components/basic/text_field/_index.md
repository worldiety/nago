---
# Content is auto generated
# Manual changes will be overwritten!
title: Text Field
---
This component provides a text input field with optional supporting and error text,
leading/trailing views (e. g. , icons), debounce settings, styling, and keyboard options. It supports both controlled (via State) and uncontrolled (via value) modes. It is typically used in forms, search bars, and other user input scenarios.

## Constructors
### FloatField
FloatField is just a TextField using the according keyboard hints. Remember, that these IME hints are no guarantees
and a user may enter non-integer stuff anyway. However, any
incompatible inputs are ignored and the given int-state is just a kind of view on top of the string state.
See also [FloatFieldValue] if you just want to display a non-stateful float value.

### FloatFieldValue
FloatFieldValue just renders a non-stateful float value. See also [FloatField]. Due to the generic instantiation,
one can influence the float rendering through the Stringer interface.

### IntField
IntField is just a TextField using the according keyboard hints. Remember, that these IME hints are no guarantees
and a user may enter non-integer stuff anyway. However, any
incompatible inputs are ignored and the given int-state is just a kind of view on top of the string state.

### TextField
TextField creates a new text field with the given label and initial value.
By default, it is single-line and uncontrolled until InputValue is set.

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` | AccessibilityLabel is a placeholder implementation. |
| `Border(border Border)` | Border is a placeholder implementation. |
| `Debounce(enabled bool)` | Debounce is enabled by default. See also DebounceTime. |
| `DebounceTime(d time.Duration)` | DebounceTime sets a custom debouncing time when entering text. By default, this is 500ms and always applied. You can disable debouncing, but be very careful with that, as it may break your server, the client or network. |
| `Disabled(disabled bool)` | Disabled disables or enables the field. When disabled, the user cannot interact with the field. |
| `ErrorText(text string)` | ErrorText sets an error message for the field. When provided, this text is shown below the input in place of supporting text, usually styled to indicate an error state. |
| `Frame(frame Frame)` | Frame sets the layout frame of the field (size, width, height, etc.). |
| `FullWidth()` | FullWidth expands the text field to take the full available width. |
| `ID(id string)` | ID assigns a unique identifier to the text field. Useful for testing, accessibility, or programmatic interaction. |
| `InputValue(input *core.State[string])` | InputValue binds the text field to a reactive state. This enables controlled input behavior where the state is updated as the user types. |
| `KeyboardOptions(options TKeyboardOptions)` | KeyboardOptions sets advanced keyboard behavior (type, capitalization, return key, etc.). |
| `KeyboardType(keyboardType KeyboardType)` | KeyboardType sets the type of keyboard to display (e.g., text, number, email). |
| `KeydownEnter(fn func())` | KeydownEnter currently only works for one line text fields (lines=0) and not for text area. The enter key logic collides with the new line logic and it is currently not clear how this situation shall be handled:   - there must be a combined key gesture   - using shift and enter for new lines is surprising for any user   - using the inversion, which is shift and enter for submitting is also wrong, because that is already overloaded     in multiple ways (e.g. line break vs paragraph behavior or opening a new window in Chrome etc)   - same applies to Str + Enter which may also be overloaded, typically for a soft line break |
| `Label(label string)` | Label sets the label text of the field. Unlike other setters, this does not return a modified copy of TTextField. |
| `Leading(v core.View)` | Leading sets a leading view for the field. This view is displayed at the start of the input field, e.g., an icon. |
| `Lines(lines int)` | Lines are by default at 0 and enforces a single line text field. Otherwise, a text area is created. This is also true, if lines 1 to differentiate between subtile behavior of single line text fields and single line text areas, which may take even more lines, because e.g. a web browser allows to change that on demand. |
| `Max(max float64)` | Max defines the max value of number fields |
| `Min(min float64)` | Min defines the min value of number fields |
| `Padding(padding Padding)` | Padding is a placeholder implementation. |
| `ShowZero(showZero bool)` | ShowZero defines wheter the '0' character should be displayed for empty/zero values in number fields. |
| `Step(step int)` | Step defines the step size to increase/decrease number values stepwise |
| `Style(s TextFieldStyle)` | Style sets the wanted style. If empty, [proto.TextFieldOutlined] is applied. |
| `SupportingText(text string)` | SupportingText sets helper text for the field. This text is displayed below the input and is typically used to provide hints or guidance. |
| `TextAlignment(v TextAlignment)` |  |
| `Trailing(v core.View)` | Trailing sets a trailing view for the field. This view is displayed at the end of the input field, e.g., a clear button or icon. |
| `Visible(v bool)` | Visible toggles the visibility of the text field. When set to false, the field is hidden from view but still part of the layout. |
| `WithFrame(fn func(Frame) Frame)` | WithFrame updates the current frame of the field via a transformation function. |
---

## Related
- [Border](../../utility/border/)
- [Frame](../../layout/frame/)
- [Padding](../../utility/padding/)
- [Keyboard Options](../../utility/keyboard_options/)

