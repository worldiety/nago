---
# Content is auto generated
# Manual changes will be overwritten!
title: Grid Cell
---
It represents a single cell inside a grid layout, defining its
position, span, alignment, padding, and background color.

## Constructors
### GridCell
GridCell creates a new grid cell containing the given body view.
Rows and columns are 1-based (starting at 1). By default, the cell
uses Stretch alignment, meaning its content will expand to fill the
available space. If another alignment is set, the content uses
wrap-content semantics and the grid background becomes visible.

---
## Methods
| Method | Description |
|--------| ------------|
| `Alignment(a Alignment)` | Alignment sets the alignment of the content within the grid cell. |
| `BackgroundColor(color Color)` | BackgroundColor sets the background color of the grid cell. |
| `ColEnd(colEnd int)` | ColEnd must be always at least +1 of ColStart, even if that column is beyond the defined amount of total columns. |
| `ColSpan(colSpan int)` | ColSpan behavior is unspecified and can sometime make your life easier, because you must not exactly know the layout. However, it may also behave unexpectedly, especially when overlapped. |
| `ColStart(colStart int)` | ColStart must start at 1. |
| `Padding(p Padding)` | Padding sets the inner spacing around the cell content. |
| `RowEnd(rowEnd int)` | RowEnd must be always at least +1 of RowStart, even if that row is beyond the defined amount of total rows. |
| `RowSpan(rowSpan int)` | RowSpan behavior is unspecified and can sometime make your life easier, because you must not exactly know the layout. However, it may also behave unexpectedly, especially when overlapped. |
| `RowStart(rowStart int)` | RowStart must start at 1. |
| `render(ctx core.RenderContext)` |  |
---

## Related
- [Alignment](../../layout/alignment/)
- [Padding](../../utility/padding/)

