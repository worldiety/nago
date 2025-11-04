---
# Content is auto generated
# Manual changes will be overwritten!
title: Tool Window
---
This component displays a window with an icon, title, and optional
top, content, and bottom sections. It can be positioned and toggled visible.

## Constructors
### ToolWindow
ToolWindow creates a new TVToolWindow with the given icon and name.
By default, the window is visible.

### ToolWindowList
ToolWindowList creates a TVToolWindow that displays a list of items
based on the provided configuration. It supports selection, deletion,
creation, and optional actions for each item.

---
## Methods
| Method | Description |
|--------| ------------|
| `Bottom(v ui.DecoredView)` | Bottom sets the bottom section of the tool window to the given view. |
| `Content(v ui.DecoredView)` | Content sets the main content section of the tool window to the given view. |
| `Top(v ui.DecoredView)` | Top sets the top section of the tool window to the given view. |
| `Visible(b bool)` | Visible sets the visibility state of the tool window. |
---

