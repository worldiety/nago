---
# Content is auto generated
# Manual changes will be overwritten!
title: Table Column
---
It defines the configuration for a table column header and its cell defaults. Columns can define width, alignment, background color, padding, borders,
and cell-specific actions (e. g. , sorting).

## Constructors
### TableColumn
TableColumn creates a new table column with the given header content.

---
## Methods
| Method | Description |
|--------| ------------|
| `Action(action func())` | Action sets an optional click/tap action for the column's cells. |
| `Alignment(alignment Alignment)` | Alignment sets the content alignment within the column cell. |
| `BackgroundColor(backgroundColor Color)` | BackgroundColor sets the background color for the column cell. |
| `Border(border Border)` | Border sets the border for the column cell. |
| `HoveredBackgroundColor(backgroundColor Color)` | HoveredBackgroundColor sets the background color when the column cell is hovered. |
| `Padding(padding Padding)` | Padding sets the padding for the column cell. |
| `Span(span int)` | Span sets how many columns this header should span. |
| `Width(width Length)` | Width sets the column width. |
---

## Related
- [Alignment](../../layout/alignment/)
- [Border](../../utility/border/)
- [Padding](../../utility/padding/)

