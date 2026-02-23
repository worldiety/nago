---
# Content is auto generated
# Manual changes will be overwritten!
title: Box
---
It lays out its children according to BoxLayout rules. By definition,
the container clips its children. This makes it suitable for overlapping
layouts, but usually requires absolute height and width. Shadows may require
extra padding since clipped children cannot extend beyond the container.

## Constructors
### Box
Box is a container, in which the given children will be layout to the according BoxLayout
rules. Note, that per definition the container clips its children. Thus, if working with shadows,
you need to apply additional padding. Important: this container requires usually absolute height and width
attributes and cannot work properly using wrap content semantics, because it intentionally allows overlapping.

### BoxAlign
BoxAlign creates a new Box with a single child aligned according to
the given alignment position (e.g., Top, Center, Bottom, etc.).

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` | AccessibilityLabel sets a label used by screen readers for accessibility. |
| `BackgroundColor(backgroundColor Color)` | BackgroundColor sets the background color of the box. |
| `Border(border Border)` | Border sets the border styling of the box. |
| `DisableOutsidePointerEvents(disable bool)` | DisableOutsidePointerEvents controls whether pointer events are disabled outside the box's content. |
| `Font(font Font)` | Font sets the font style for text content inside the box. |
| `Frame(fr Frame)` | Frame sets the layout frame of the box, including size and positioning. |
| `FullWidth()` | FullWidth sets the box to span the full available width. |
| `Padding(p Padding)` | Padding sets the inner spacing around the box's children. |
| `Visible(visible bool)` | Visible controls the visibility of the box; setting false hides it. |
| `WithFrame(fn func(Frame) Frame)` | WithFrame applies a transformation function to the box's frame and returns the updated component. |
---

## Related
- [Border](../../utility/border/)
- [Frame](../../layout/frame/)
- [Padding](../../utility/padding/)

