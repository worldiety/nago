---
title: Conditional Rendering
linkTitle: Conditionals
---

Views in Nago are plain Go values, so you build them with ordinary Go code. A few helpers make conditions and
loops fit into a builder expression. Containers skip `nil` children, which is what `If` returns when the condition
is false.

```go
func orderView(order Order) core.View {
	return VStack(
		Text(order.Title).Font(HeadlineSmall),
		If(order.Shipped, Text("Shipped")),
		IfElse(order.Paid, Text("Paid"), Text("Payment pending")),
		IfFunc(order.HasInvoice, func() core.View {
			return invoiceView(order) // only built if HasInvoice is true
		}),
		VStack(
			ForEach(order.Items, func(item Item) core.View {
				return Text(item.Name)
			})...,
		).Alignment(Leading),
	).Alignment(Leading).Gap(L8)
}
```

## Functions

| Function | Description |
|----------|-------------|
| `If(b bool, t core.View) core.View` | Returns the view if b is true, otherwise nil. |
| `IfElse(b bool, ifTrue, ifFalse core.View) core.View` | Returns one or the other view. |
| `IfFunc(b bool, fn func() core.View) core.View` | Like `If`, but creates the view lazily, only if b is true. |
| `ForEach[T, V any](seq []T, m func(T) V, more ...V) []V` | Maps each element of a slice to a view; `more` is appended at the end. |
| `Each[T, V any](seq iter.Seq[T], m func(T) V, more ...V) []V` | Like `ForEach`, but for an `iter.Seq`. |
| `Each2[K, V, X any](seq iter.Seq2[K, V], m func(K, V) X) []X` | Maps each key-value pair of an `iter.Seq2`. |
| `With[T any](t T, with func(T) T) T` | Passes a value through a function, to intercept a builder chain without a local variable. |

`If`, `IfElse` and `IfFunc` evaluate their arguments like any Go call: the views of `If` and `IfElse` are built
even if they are not shown. Use `IfFunc` or `Lazy` when building a view is expensive or only valid under the
condition.

## Lazy views

`Composable`, also available as `Lazy`, is a `func() core.View` which implements `core.View` itself. It is
evaluated only when it is rendered:

```go
Lazy(func() core.View {
	return expensiveView()
})
```

Components with their own `With` method, like [`TStack.With`](../../layout/stack/), apply a function to themselves
in the same way.

## Related

- [View That Matches](../view_that_matches/), [Stack](../../layout/stack/)
- Tutorials: [Responsive](/docs/examples/tutorial-06-responsive/), [Dialog](/docs/examples/tutorial-07-dialog/),
  [Long list](/docs/examples/tutorial-90-long-list/)
