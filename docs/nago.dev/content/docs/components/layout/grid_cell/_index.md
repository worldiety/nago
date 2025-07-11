---
# Content is auto generated
# Manual changes will be overwritten!
title: Grid Cell
---


## Constructors
### GridCell
GridCell creates a cell based on the given body. Rows and Columns start at 1, not zero.
Without any alignment rules, the cell will stretch its body automatically to the calculated
cell dimensions. Otherwise, if a cell alignment is set, the size is wrap-content semantics
and the background of the grid will be visible. Thus, the default specification of no-alignment
is different here.

---
## Methods
| Method | Description |
|--------| ------------|
| `Alignment(a Alignment)` |  |
| `BackgroundColor(color Color)` |  |
| `ColEnd(colEnd int)` | ColEnd must be always at least +1 of ColStart, even if that column is beyond the defined amount of total columns. |
| `ColSpan(colSpan int)` | ColSpan behavior is unspecified and can sometime make your life easier, because you must not exactly know
the layout. However, it may also behave unexpectedly, especially when overlapped. |
| `ColStart(colStart int)` | ColStart must start at 1. |
| `Padding(p Padding)` |  |
| `RowEnd(rowEnd int)` | RowEnd must be always at least +1 of RowStart, even if that row is beyond the defined amount of total rows. |
| `RowSpan(rowSpan int)` | RowSpan behavior is unspecified and can sometime make your life easier, because you must not exactly know
the layout. However, it may also behave unexpectedly, especially when overlapped. |
| `RowStart(rowStart int)` | RowStart must start at 1. |
| `render(ctx core.RenderContext)` |  |
---
