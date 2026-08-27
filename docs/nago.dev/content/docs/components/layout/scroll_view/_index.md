---
# Content is auto generated
# Manual changes will be overwritten!
title: Scroll View
---
It provides a scrollable container for a single child view. The scroll direction can be vertical (default) or horizontal. Supports customization of frame, position, border, background color, and padding.

## Constructors
### ScrollView
A ScrollView can either be horizontal or vertical. By default, it is vertical.

---
## Methods
| Method | Description |
|--------| ------------|
| `Axis(axis ScrollViewAxis)` | Axis sets the scroll direction (vertical or horizontal). |
| `BackgroundColor(color Color)` | BackgroundColor sets the background color of the scroll view. |
| `Border(border Border)` | Border applies a border around the scroll view. |
| `Content(content core.View)` |  |
| `Frame(frame Frame)` | Frame sets the layout frame for the scroll view. |
| `FullWidth()` |  |
| `ListLength(length int)` | ListLength sets the length of the content list. This is optional but should be used if the content is based on a list. |
| `Padding(padding Padding)` | Padding sets the inner padding of the scroll view. |
| `Position(position Position)` | Position sets the alignment of the content inside the scroll view. |
| `ScrollAlignment(alignment ScrollAlignment)` | ScrollAlignment defines how the component should align the scroll target when scrolling into view |
| `ScrollBehavior(behavior ScrollBehavior)` | ScrollBehavior defines how the component should behave when the scrollable content grows |
| `ScrollButtonLabel(label string)` | ScrollButtonLabel sets the label of scroll button when the component asks whether to scroll |
| `ScrollToView(scrollToView string, animation ScrollAnimation)` |  |
---

## Related
- [Border](../../utility/border/)
- [Frame](../../layout/frame/)
- [Padding](../../utility/padding/)

