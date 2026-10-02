---
title: Components
weight: 5
sidebar:
  open: false
---

Components are the building blocks of every view. They live in
[`presentation/ui`](https://github.com/worldiety/nago/tree/main/presentation/ui) and its subpackages, and each one
implements `core.View`. The pages are grouped into:

{{< cards >}}
  {{< card link="basic" title="Basic" subtitle="Text, buttons, images and input fields." >}}
  {{< card link="layout" title="Layout" subtitle="Stacks, grids, boxes and other containers." >}}
  {{< card link="composite" title="Composite" subtitle="Ready-made combinations like forms, tables and charts." >}}
  {{< card link="feedback-and-overlay" title="Feedback & Overlay" subtitle="Dialogs, banners and alerts." >}}
  {{< card link="utility" title="Utility" subtitle="Styling values, conditional rendering and helpers." >}}
{{< /cards >}}

## Fluent builders

A constructor like `VStack` or `Text` returns a component value, a struct prefixed with `T` (`TStack`, `TText`).
You configure it by chaining methods. Every method works on a copy and returns the modified copy:

```go
func greeting(name string) core.View {
	return VStack(
		Text("Hello").Font(HeadlineMedium),
		Text(name),
	).Alignment(Leading).
		Gap(L8).
		Padding(Padding{}.All(L16))
}
```

Because components are plain values, a method never changes the receiver. Assign the result if you build a
component step by step:

```go
stack := HStack(Text("A"), Text("B"))
stack.Gap(L16)         // no effect, the result is discarded
stack = stack.Gap(L16) // correct
```

The same applies to styling values like [`Padding`](utility/padding/), [`Frame`](layout/frame/) and
[`Border`](utility/border/): `Padding{}.All(L16)` starts with the zero value and returns a new one.

## Where a chain ends

Some methods return the interface `DecoredView` instead of the concrete type, typically `Padding`, `Frame`,
`WithFrame`, `Border`, `Visible` and `AccessibilityLabel`. After such a call only the methods of `DecoredView` are
left, so call them last:

```go
VStack(Text("A")).
	Gap(L8).                     // TStack
	BackgroundColor(M4).         // TStack
	Padding(Padding{}.All(L16)). // DecoredView from here on
	Border(Border{}.Radius(L8))  // DecoredView
```

`VStack(...).Padding(...).Gap(L8)` does not compile, because `DecoredView` has no `Gap` method. The method
tables on the component pages show the return type where it is not the component itself.

{{< callout type="info" >}}
Every component also has a `Render(core.RenderContext) core.RenderNode` method which implements `core.View`. Nago
calls it for you, so the method tables leave it out.
{{< /callout >}}
