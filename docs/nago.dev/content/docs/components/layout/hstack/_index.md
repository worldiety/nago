---
# Content is auto generated
# Manual changes will be overwritten!
title: HStack
---
HStack is a horizontal layout container that arranges its child views in a row. It supports alignment, spacing, background styling, borders, and interaction states. The HStack is interactive if an action is defined and can respond to hover, press,
and focus states with visual feedback.

## Constructors
### HStack
HStack is a container, in which the given children will be layout in a row according to the applied
alignment rules. Note, that per definition the container clips its children. Thus, if working with shadows,
you need to apply additional padding.

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` |  |
| `Action(f func())` |  |
| `Alignment(alignment Alignment)` |  |
| `Append(children ...)` |  |
| `BackgroundColor(backgroundColor Color)` |  |
| `Border(border Border)` |  |
| `Enabled(enabled bool)` | Enabled has only an effect if StylePreset is applied, otherwise it is ignored. |
| `FocusedBackgroundColor(backgroundColor Color)` |  |
| `FocusedBorder(border Border)` |  |
| `Font(font Font)` |  |
| `Frame(fr Frame)` |  |
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
| `Visible(visible bool)` |  |
| `With(fn func(stack THStack) THStack)` |  |
| `WithFrame(fn func(Frame) Frame)` |  |
| `Wrap(wrap bool)` | Wrap tries to reproduce the flex-box wrap behavior. This means, that if the HStack has a limited width, it must create multiple rows to place its children. Note, that the text layout behavior is unspecified (it may layout without word-wrap or use some sensible defaults). Each row and each element may have its own custom size, so this must not use a grid-like layouting. |
---
## Related
- [Frame](../../layout/frame/)
- [HStack](../../layout/hstack/)

## Tutorials
- [tutorial-02-combining-views](../../../examples/tutorial-02-combining-views)
