---
# Content is auto generated
# Manual changes will be overwritten!
title: Text Layout
---
It arranges multiple text (and text-like) child views with shared typography,
spacing, and alignment. Useful for paragraphs, captions, or any block of text
that needs consistent styling and optional interaction (action callback).

## Constructors
### TextLayout
TextLayout performs an inline layouting of multiple text elements. The alignment properties of each
Text are ignored. Any implementation must support an arbitrary amount of text elements with different
font settings. However, implementations are also open to support images and any other views, as long
as they can be rendered inline.

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` | AccessibilityLabel sets the screen-reader label describing the layout's content or purpose. |
| `Action(f func())` | Action sets a callback function that is executed when the layout is clicked. |
| `Alignment(alignment TextAlignment)` | Alignment defines the text alignment within the layout. |
| `BackgroundColor(backgroundColor Color)` | BackgroundColor sets the background color of the layout. |
| `Border(border Border)` | Border applies the given border (widths, radii, colors, shadow) to the layout. |
| `Font(font Font)` | Font sets the font styling for the text in the layout. |
| `Frame(f Frame)` | Frame sets the dimensions and position of the layout. |
| `FullWidth()` | FullWidth expands the layout to occupy the full available width. |
| `Padding(padding Padding)` | Padding sets the inner spacing of the layout. |
| `Visible(visible bool)` | Visible toggles visibility of the layout. Setting false hides it. |
| `WithFrame(fn func(Frame) Frame)` | WithFrame modifies the current frame using the provided function. |
---

## Related
- [Border](../../utility/border/)
- [Frame](../../layout/frame/)
- [Padding](../../utility/padding/)

