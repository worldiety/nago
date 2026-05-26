---
# Content is auto generated
# Manual changes will be overwritten!
title: Card Layout
---
It organizes child views into a grid-like layout, typically using multiple
columns. The number of columns can be defined globally or customized per
window size class to enable responsive design.

## Constructors
### Layout
Layout creates a new TCardLayout with the given child views.
Nil children are automatically skipped. By default, the layout uses 3 columns.

---
## Methods
| Method | Description |
|--------| ------------|
| `Columns(class core.WindowSizeClass, columns int)` | Columns sets a custom number of columns for a specific WindowSizeClass. This allows the layout to adapt responsively to different screen sizes. |
| `Frame(frame ui.Frame)` | Frame sets the frame (size and positioning) for the card layout. |
| `Padding(padding ui.Padding)` | Padding sets the inner spacing for the card layout. This allows customizing the distance between the layout border and its content. |
---

