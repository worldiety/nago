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
| `AccessibilityLabel(label string)` | AccessibilityLabel sets the label used by screen readers for accessibility. |
| `Action(f func())` | Action sets the callback function to be invoked when the stack is clicked or tapped. |
| `Alignment(alignment Alignment)` | Alignment sets how the stack's children are aligned vertically within the horizontal row. |
| `Append(children ...)` | Append adds one or more child views to the horizontal stack. |
| `Background(bg Background)` |  |
| `BackgroundColor(backgroundColor Color)` | BackgroundColor sets the background color of the horizontal stack. |
| `Border(border Border)` | Border sets the border styling of the stack. |
| `Enabled(enabled bool)` | Enabled has only an effect if StylePreset is applied, otherwise it is ignored. |
| `FocusedBackgroundColor(backgroundColor Color)` | FocusedBackgroundColor sets the background color of the stack when it is focused (e.g., via keyboard navigation). |
| `FocusedBorder(border Border)` | FocusedBorder sets the border styling when the stack is focused. |
| `Font(font Font)` | Font sets the font style applied to text content inside the stack. |
| `Frame(fr Frame)` | Frame sets the layout frame of the horizontal stack, including size and positioning. |
| `FullWidth()` | FullWidth sets the stack to span the full available width. |
| `Gap(gap Length)` | Gap sets the spacing between child views in the horizontal stack. |
| `HoveredBackgroundColor(backgroundColor Color)` | HoveredBackgroundColor sets the background color of the stack when the user hovers over it. |
| `HoveredBorder(border Border)` | HoveredBorder sets the border styling when the stack is hovered. |
| `ID(id string)` | ID assigns a unique identifier to the stack, useful for testing or referencing. |
| `NoClip(b bool)` | NoClip toggles whether the stack clips its children. By default, stacks clip their children; setting true disables clipping. |
| `Opacity(opacity float64)` | Opacity sets the visibility of this component. The range is [0..1] where 0 means fully transparent and 1 means fully visible. This also affects all contained children. |
| `Padding(padding Padding)` | Padding sets the inner spacing around the stack's children. |
| `Position(position Position)` | Position sets the position of the horizontal stack within its parent layout. |
| `PressedBackgroundColor(backgroundColor Color)` | PressedBackgroundColor sets the background color of the stack when it is pressed or clicked. |
| `PressedBorder(border Border)` | PressedBorder sets the border styling when the stack is pressed or clicked. |
| `StylePreset(preset StylePreset)` | StylePreset applies a predefined style preset to the stack, controlling its appearance. |
| `Target(target string)` | Target sets the name of the browsing context, like _self, _blank, _ parent, _top. |
| `TextColor(textColor Color)` | TextColor sets the color of text content inside the stack. |
| `URL(url core.URI)` | URL sets the URL that the button navigates to when clicked if no action is specified. If both URL and Action are set, the URL takes precedence. This avoids another render cycle if the only goal is to navigate to a different page. It also avoids issues with browser which block async browser interactions like Safari. In fact, the [core.Navigation.Open] does not work properly on Safari. See also [TButton.Target]. |
| `Visible(visible bool)` | Visible controls the visibility of the stack; setting false hides it. |
| `With(fn func(stack THStack) THStack)` | With applies a transformation function to the stack itself and returns the result. Useful for chaining configuration in a functional style. |
| `WithFrame(fn func(Frame) Frame)` | WithFrame applies a transformation function to the stack's frame and returns the updated component. |
| `Wrap(wrap bool)` | Wrap tries to reproduce the flex-box wrap behavior. This means, that if the HStack has a limited width, it must create multiple rows to place its children. Note, that the text layout behavior is unspecified (it may layout without word-wrap or use some sensible defaults). Each row and each element may have its own custom size, so this must not use a grid-like layouting. |
---

## Related
- [Alignment](../../layout/alignment/)
- [Border](../../utility/border/)
- [Frame](../../layout/frame/)
- [Padding](../../utility/padding/)

## Tutorials
- [tutorial-02-combining-views](../../../examples/tutorial-02-combining-views)
