---
# Content is auto generated
# Manual changes will be overwritten!
title: Password Field
---
It provides a secure input field for entering passwords or secrets. Unlike normal text fields, it ensures that sensitive values are not exposed
after input. It supports validation feedback, debouncing, autocomplete
control, multiline behavior, accessibility labels, and styling options.

## Constructors
### PasswordField
PasswordField represents a secret entered by the user.
It is very important for the security of your implementation, that you
must not expose any secret through the value after the user has finished the secret step.
So, consider the following situations:
  - a user wants to change his password and fills out 3 password fields. Thus, you don't want, that the user requires
    to enter his secret any time, e.g. if the new password does not match the guidelines
  - by design, Nago requires to fill in the value whatever that was, so that it will not get lost between render
    cycles.
  - by security, you should never store any password, neither in plain text and also never encrypted and never using
    just a simple cryptographic hash. Use a password derivation function instead, like argonid or others.
  - if you need to store a secret in plain text, e.g. like an API token, you must not show that later again to
    the user, after the insertion phase of you form flow is over.

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` | AccessibilityLabel sets an accessibility label for screen readers. |
| `AutoComplete(autoComplete bool)` | AutoComplete enables or disables browser autocomplete for the field. |
| `Border(border Border)` | Border sets the border styling of the field. |
| `Debounce(enabled bool)` | Debounce is enabled by default. See also DebounceTime. |
| `DebounceTime(d time.Duration)` | DebounceTime sets a custom debouncing time when entering text. By default, this is 500ms and always applied. You can disable debouncing, but be very careful with that, as it may break your server, the client or network. |
| `Disabled(disabled bool)` | Disabled enables or disables the field. |
| `ErrorText(text string)` | ErrorText sets the error message displayed when validation fails. |
| `Frame(frame Frame)` | Frame sets the layout frame for the field. |
| `FullWidth()` | FullWidth expands the field to take up the full available width. |
| `ID(id string)` | ID sets a unique identifier for the field. |
| `InputValue(input *core.State[string])` | InputValue binds the password field to the given state for two-way data binding. |
| `KeydownEnter(fn func())` | KeydownEnter sets a callback function to be triggered when the Enter key is pressed. |
| `Label(label string)` | Label sets the field label. |
| `Lines(lines int)` | Lines are by default at 0 and enforces a single line text field. Otherwise, a text area is created. This is also true, if lines 1 to differentiate between subtile behavior of single line text fields and single line text areas, which may take even more lines, because e.g. a web browser allows to change that on demand. |
| `Padding(padding Padding)` | Padding sets the padding around the field. |
| `Style(s TextFieldStyle)` | Style sets the wanted style. If empty, [proto.TextFieldOutlined] is applied. |
| `SupportingText(text string)` | SupportingText sets helper text displayed below the field. |
| `Visible(v bool)` | Visible sets the field's visibility. |
| `WithFrame(fn func(Frame) Frame)` | WithFrame applies a transformation function to the field's frame. |
---

## Related
- [Border](../../utility/border/)
- [Frame](../../layout/frame/)
- [Padding](../../utility/padding/)

