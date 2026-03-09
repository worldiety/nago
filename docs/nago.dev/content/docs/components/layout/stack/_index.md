---
# Content is auto generated
# Manual changes will be overwritten!
title: Stack
---
It is responsive and can switch between [HStack] and [VStack] during rendering.

## Constructors
### HStack
HStack is a container, in which the given children will be layout in a row according to the applied
alignment rules. Note, that per definition the container clips its children. Thus, if working with shadows,
you need to apply additional padding.

### Stack
Stack is a responsive variant which decides between [VStack] and [HStack].

### VStack
VStack is a container, in which the given children will be layout in a column according to the applied
alignment rules. Note, that per definition the container clips its children. Thus, if working with shadows,
you need to apply additional padding.

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` | AccessibilityLabel sets the label used by screen readers for accessibility. |
| `Action(f func())` | Action sets the callback function to be invoked when the stack is clicked or tapped. |
| `Alignment(alignment Alignment)` |  |
| `Animation(animation Animation)` |  |
| `Append(children ...)` |  |
| `Background(bg Background)` |  |
| `BackgroundColor(color Color)` |  |
| `Border(border Border)` |  |
| `Enabled(enabled bool)` | Enabled has only an effect if StylePreset is applied, otherwise it is ignored. |
| `FocusedBackgroundColor(backgroundColor Color)` | FocusedBackgroundColor sets the background color of the stack when it is focused (e.g., via keyboard navigation). |
| `FocusedBorder(border Border)` | FocusedBorder sets the border styling when the stack is focused. |
| `Font(font Font)` | Font sets the font style applied to text content inside the stack. |
| `Frame(frame Frame)` |  |
| `FullWidth()` |  |
| `Gap(gap Length)` |  |
| `HRef(url core.URI)` | HRef sets the URL that the button navigates to when clicked if no action is specified. If both URL and Action are set, the URL takes precedence. This avoids another render cycle if the only goal is to navigate to a different page. It also avoids issues with browser which block async browser interactions like Safari. In fact, the [core.Navigation.Open] does not work properly on Safari. See also [TButton.Target]. |
| `HoveredBackgroundColor(backgroundColor Color)` | HoveredBackgroundColor sets the background color of the stack when the user hovers over it. |
| `HoveredBorder(border Border)` | HoveredBorder sets the border styling when the stack is hovered. |
| `ID(id string)` | ID assigns a unique identifier to the stack, useful for testing or referencing. |
| `Layout(layout StackLayout)` |  |
| `NoClip(b bool)` |  |
| `Opacity(opacity float64)` | Opacity sets the visibility of this component. The range is [0..1] where 0 means fully transparent and 1 means fully visible. This also affects all contained children. |
| `Padding(padding Padding)` |  |
| `Position(position Position)` | Position sets the position of the horizontal stack within its parent layout. |
| `PressedBackgroundColor(backgroundColor Color)` | PressedBackgroundColor sets the background color of the stack when it is pressed or clicked. |
| `PressedBorder(border Border)` | PressedBorder sets the border styling when the stack is pressed or clicked. |
| `Responsive(fn func(wnd core.Windowstack TStack) TStack)` |  |
| `StylePreset(preset StylePreset)` | StylePreset applies a predefined style preset to the stack, controlling its appearance. |
| `Target(target string)` | Target sets the name of the browsing context, like _self, _blank, _ parent, _top. |
| `TextColor(textColor Color)` | TextColor sets the color of text content inside the stack. |
| `Transformation(transformation Transformation)` |  |
| `Visible(visible bool)` | Visible controls the visibility of the stack; setting false hides it. |
| `WithFrame(fn func(Frame) Frame)` | WithFrame applies a transformation function to the stack's frame and returns the updated component. |
| `WithPadding(padding Padding)` |  |
| `Wrap(wrap bool)` | Wrap tries to reproduce the flex-box wrap behavior. This means, that if the HStack has a limited width, it must create multiple rows to place its children. Note, that the text layout behavior is unspecified (it may layout without word-wrap or use some sensible defaults). Each row and each element may have its own custom size, so this must not use a grid-like layouting. |
---

## Related
- [Alignment](../../layout/alignment/)
- [Border](../../utility/border/)
- [Frame](../../layout/frame/)
- [Padding](../../utility/padding/)

