---
title: Views and Modifiers
linkTitle: Views
weight: 2
---

A Nago user interface is a tree of views. A view is any value which implements `core.View`:

```go
type View interface {
	Render(RenderContext) RenderNode
}
```

You rarely implement this interface yourself. The package `go.wdy.de/nago/presentation/ui` and its sub packages
provide the [components](/docs/components/): texts, buttons, stacks, grids, forms, tables, dialogs and many
more. You compose them into a tree and return it from a [root view](../navigation/).

## Composing views

Containers take their children as arguments:

```go
ui.VStack(
	ui.Text("Bumblebee").Font(ui.Title),
	ui.HStack(
		ui.Text("Oldenburg"),
		ui.Spacer(),
		ui.Text("Germany"),
	).FullWidth(),
	ui.PrimaryButton(func() { fmt.Println("clicked") }).Title("Save"),
).Gap(ui.L16).Alignment(ui.Leading)
```

`VStack` arranges its children vertically, `HStack` horizontally. Sizes are expressed as `ui.Length`, e.g. the
predefined `ui.L16` (1rem), `ui.Full` or a CSS value like `"50%"`.

To create your own component, write a function which returns a view:

```go
func CircleImage(uri core.URI) ui.DecoredView {
	return ui.Image().
		URI(uri).
		Border(ui.Border{}.Circle().Width(ui.L4).Color("#ffffff")).
		Frame(ui.Frame{}.Size(ui.L320, ui.L320))
}
```

This is all there is to custom components: no registration, no templates. See
[tutorial-02-combining-views](/docs/examples/tutorial-02-combining-views/).

## Modifiers and value semantics

Components are small structs like `ui.TText` or `ui.TStack` with value receivers. Every modifier, e.g. `Font`,
`Gap` or `Padding`, returns a modified **copy**. That is why you can chain them, and why a modifier on its own
has no effect:

```go
text := ui.Text("hello")
text.Font(ui.Title)          // wrong: the copy is discarded
text = text.Font(ui.Title)   // right
```

Views are cheap to create, because they are plain values and Nago builds a new tree on each render anyway.

### Where the chain ends

Modifiers which are common to all components, i.e. `Padding`, `Frame`, `Border`, `Visible` and
`AccessibilityLabel`, return the interface `ui.DecoredView` instead of the concrete type. After one of them you
can only call other `DecoredView` methods:

```go
ui.VStack(...).Gap(ui.L8).Frame(ui.Frame{}.FullWidth()) // fine
ui.VStack(...).Frame(ui.Frame{}.FullWidth()).Gap(ui.L8) // does not compile
```

Therefore, put type-specific modifiers first and the decorating ones last. Some components offer variants like
`TStack.WithPadding`, which keep the concrete type.

## Conditions and lists

Views are Go values, so you use Go to compute them. For inline cases, the `ui` package provides helpers:

| Helper                     | Purpose                                                                            |
|----------------------------|------------------------------------------------------------------------------------|
| `ui.If(cond, view)`        | returns the view or nil, nil children are ignored                                  |
| `ui.IfElse(cond, a, b)`    | returns one of two views                                                           |
| `ui.IfFunc(cond, fn)`      | like `If`, but creates the view only if the condition is true                      |
| `ui.ForEach(slice, fn)`    | maps a slice to a slice of views, use it with `...` as children                    |
| `ui.Each(seq, fn)`         | the same for an `iter.Seq`                                                         |
| `ui.Lazy(fn)`              | wraps a function, which is called at render time to create the view                |

```go
ui.VStack(
	ui.If(len(people) == 0, ui.Text("nobody here")),
).Append(ui.ForEach(people, func(p Person) core.View {
	return ui.Text(p.Name)
})...)
```

## Text and icons

Use `ui.Text` for text, `ui.Image().Embed(svg)` or `ui.ImageIcon(svg)` for icons. Nago ships icon sets as Go
packages, e.g. `go.wdy.de/nago/presentation/icons/hero/solid`. See
[tutorial-10-icons](/docs/examples/tutorial-10-icons/) and
[tutorial-111-typography](/docs/examples/tutorial-111-typography/).

## Related

- [Components](/docs/components/)
- [State and rendering](../state/) – making views interactive
- [tutorial-06-responsive](/docs/examples/tutorial-06-responsive/) – adapting the tree to the window size
