---
# Content is auto generated
# Manual changes will be overwritten!
title: VStack
---
VStack is a vertical layout container that arranges its child views in a column. It supports alignment, spacing, background styling, borders, and interaction states. The VStack can be interactive if an action is defined and responds to hover, press,
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
| `AccessibilityLabel(label string)` | AccessibilityLabel sets the accessibility label for screen readers. |
| `Action(f func())` | Action assigns an action handler, making the VStack interactive. |
| `Alignment(alignment Alignment)` | Alignment sets the alignment of child views within the column. |
| `Animation(animation Animation)` |  |
| `Append(children ...)` | Append adds additional child views to the VStack. |
| `BackgroundColor(backgroundColor Color)` | BackgroundColor sets the background color. |
| `Border(border Border)` | Border sets the default border styling. |
| `FocusedBackgroundColor(backgroundColor proto.Color)` | FocusedBackgroundColor sets the background color when focused. |
| `FocusedBorder(border Border)` | FocusedBorder sets the border styling when focused. |
| `Font(font Font)` | Font sets the default font for text children. |
| `Frame(f Frame)` | Frame sets the layout frame of the VStack. |
| `FullWidth()` | FullWidth sets the VStack to span 100% of the available width. |
| `Gap(gap Length)` | Gap sets the spacing between child views. |
| `HoveredBackgroundColor(backgroundColor Color)` | HoveredBackgroundColor sets the background color when hovered. |
| `HoveredBorder(border Border)` | HoveredBorder sets the border styling when hovered. |
| `ID(id string)` | ID assigns a unique identifier to the VStack. |
| `NoClip(b bool)` | NoClip disables clipping of child content when true. |
| `Opacity(opacity float64)` | Opacity sets the visibility of this component. The range is [0..1] where 0 means fully transparent and 1 means fully visible. This also affects all contained children. |
| `Padding(padding Padding)` | Padding sets the inner padding of the VStack. |
| `Position(position Position)` | Position sets the positioning of the VStack. |
| `PressedBackgroundColor(backgroundColor Color)` | PressedBackgroundColor sets the background color when pressed. |
| `PressedBorder(border Border)` | PressedBorder sets the border styling when pressed. |
| `StylePreset(preset StylePreset)` | StylePreset applies a predefined style preset. |
| `TextColor(textColor Color)` | TextColor sets the default text color for the VStack. |
| `Transformation(transformation Transformation)` |  |
| `Visible(visible bool)` | Visible controls the visibility of the VStack. |
| `With(fn func(stack TVStack) TVStack)` |  |
| `WithFrame(fn func(Frame) Frame)` | WithFrame modifies the current frame using the provided function. |
---

## Related
- [Alignment](../../layout/alignment/)
- [Border](../../utility/border/)
- [Frame](../../layout/frame/)
- [Padding](../../utility/padding/)

## Tutorials
- [tutorial-01-helloworld](../../../examples/tutorial-01-helloworld)
- [tutorial-02-combining-views](../../../examples/tutorial-02-combining-views)
- [tutorial-54-codeeditor](../../../examples/tutorial-54-codeeditor)
