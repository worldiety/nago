---
# Content is auto generated
# Manual changes will be overwritten!
title: Code Editor
---
This component provides a text editor interface
optimized for writing and displaying code. It supports syntax highlighting,
configurable tab size, and optional read-only or disabled states.

## Constructors
### CodeEditor
CodeEditor creates a new code editor with the given initial value
and a default tab size of 4 spaces.

---
## Methods
| Method | Description |
|--------| ------------|
| `Disabled(b bool)` | Disabled enables or disables user interaction with the editor. |
| `Frame(frame Frame)` | Frame sets the layout frame of the editor, including size and positioning. |
| `FullWidth()` | FullWidth sets the editor to span the full available width. |
| `InputValue(state *core.State[string])` | InputValue binds the editor to an external state for controlled text value updates. |
| `Language(language string)` | Language gives a syntax highlighting hint. Defined are go, html, css, json, xml, markdown but there may be arbitrary support. |
---

## Related
- [Frame](../../layout/frame/)

## Tutorials
- [tutorial-54-codeeditor](../../../examples/tutorial-54-codeeditor)
