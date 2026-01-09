---
# Content is auto generated
# Manual changes will be overwritten!
title: Grid
---
It arranges its children into a grid of rows and columns. The grid supports flexible sizing, spacing, alignment, borders,
accessibility labels, and visibility control. For complex cases
(e. g. , overlapping cells), row/column boundaries must be explicitly defined.

## Constructors
### Grid
Grid creates a new grid containing the given cells.
For simple cases, only rows or columns need to be set.
For complex layouts (like overlapping), rows, columns,
and explicit start/end positions must be defined.

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` | AccessibilityLabel sets the accessibility label of the grid, used by screen readers. |
| `Append(cells ...)` | Append adds one or more cells to the grid. |
| `BackgroundColor(backgroundColor Color)` | BackgroundColor sets the background color of the grid. |
| `Border(border Border)` | Border sets the border styling of the grid. |
| `ColGap(g Length)` | ColGap sets the horizontal spacing between columns in the grid. |
| `Columns(cols int)` | Columns sets the amount of columns explicitly. If not set, the result is undefined. This component tries its best to calculate the right amount cells, however, when used with areas, this cannot work properly. |
| `Font(font Font)` | Font sets the font style applied to text inside the grid. |
| `Frame(fr Frame)` | Frame sets the layout frame of the grid, including size and positioning. |
| `FullWidth()` | FullWidth sets the grid to span the full available width. |
| `Gap(g Length)` | Gap sets RowGap and ColGap equally. |
| `Heights(rowHeights ...)` | Heights are optional row heights from top to bottom. |
| `Padding(padding Padding)` | Padding sets the inner spacing around the grid. |
| `RowGap(g Length)` | RowGap sets the vertical spacing between rows in the grid. |
| `Rows(rows int)` | Rows sets the amount of rows explicitly. If not set, the result is undefined. This component tries its best to calculate the right amount cells, however, when used with areas, this cannot work properly. |
| `Visible(visible bool)` | Visible controls the visibility of the grid; setting false hides it. |
| `Widths(colWidths ...)` | Widths are optional column width from left to right. If not all width are defined, the rest of widths are equally distributed based on the remaining amount of space. |
| `WithFrame(fn func(Frame) Frame)` | WithFrame applies a transformation function to the grid's frame and returns the updated component. |
| `cellCount()` | cellCount calculates the number of grid cells based on spans and column positions. It accounts for cells with explicit colSpan or colStart/colEnd values, falling back to counting as one cell if no span information is provided. |
---

## Related
- [Border](../../utility/border/)
- [Frame](../../layout/frame/)
- [Padding](../../utility/padding/)

