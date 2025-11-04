---
# Content is auto generated
# Manual changes will be overwritten!
title: Rich Text Editor
---
It provides an interactive editor for creating and modifying rich text content. The editor supports two-way data binding, read-only and disabled states,
and layout configuration via frame settings.

## Constructors
### RichTextEditor
RichTextEditor creates a new rich text editor with the given initial value.

---
## Methods
| Method | Description |
|--------| ------------|
| `Frame(frame Frame)` | Frame sets the layout frame of the editor. |
| `FullWidth()` | FullWidth expands the editor to take the full available width. |
| `InputValue(state *core.State[string])` | InputValue binds the editor's content to a state, enabling two-way data binding. |
---

## Related
- [Frame](../../layout/frame/)

