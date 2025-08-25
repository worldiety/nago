---
# Content is auto generated
# Manual changes will be overwritten!
title: VStack
---
VStack is a vertical layout container that arranges its child views in a column. It supports alignment, spacing, background styling, borders, and interaction states. The VStack is interactive if an action is defined and can respond to hover, press,
and focus states with visual feedback.

## Constructors
### VStack
VStack is a container, in which the given children will be layout in a column according to the applied
alignment rules. Note, that per definition the container clips its children. Thus, if working with shadows,
you need to apply additional padding.

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` |  |
| `Action(f func())` |  |
| `Alignment(alignment Alignment)` |  |
| `Animation(animation Animation)` |  |
| `Append(children ...)` |  |
| `BackgroundColor(backgroundColor Color)` |  |
| `Border(border Border)` |  |
| `FocusedBackgroundColor(backgroundColor proto.Color)` |  |
| `FocusedBorder(border Border)` |  |
| `Font(font Font)` |  |
| `Frame(f Frame)` |  |
| `FullWidth()` |  |
| `Gap(gap Length)` |  |
| `HoveredBackgroundColor(backgroundColor Color)` |  |
| `HoveredBorder(border Border)` |  |
| `ID(id string)` |  |
| `NoClip(b bool)` |  |
| `Padding(padding Padding)` |  |
| `Position(position Position)` |  |
| `PressedBackgroundColor(backgroundColor Color)` |  |
| `PressedBorder(border Border)` |  |
| `StylePreset(preset StylePreset)` |  |
| `TextColor(textColor Color)` |  |
| `Transformation(transformation Transformation)` |  |
| `Visible(visible bool)` |  |
| `With(fn func(stack TVStack) TVStack)` |  |
| `WithFrame(fn func(Frame) Frame)` |  |
---

## Related
- [Frame](../../layout/frame/)

## Tutorials
- [tutorial-01-helloworld](../../../examples/tutorial-01-helloworld)
- [tutorial-02-combining-views](../../../examples/tutorial-02-combining-views)
- [tutorial-54-codeeditor](../../../examples/tutorial-54-codeeditor)
