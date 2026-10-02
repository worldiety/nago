---
title: Animation
---

An `Animation` lets a [stack](../../layout/stack/) move continuously, e.g. to draw attention to a new element or to
show that something is loading, or animates changes of its style.

```go
VStack(Text("hello world")).
	BackgroundColor(ColorCardBody).
	Animation(AnimatePulse).
	Padding(Padding{}.All(L16))
```

| Constant | Effect |
|----------|--------|
| `AnimateNone` | No animation, the default. |
| `AnimateTransition` | Animates changes of properties like colors, opacity or the [transformation](../transformation/). |
| `AnimatePulse` | Fades the component in and out. |
| `AnimateBounce` | Lets the component bounce up and down. |
| `AnimatePing` | Scales the component up while fading it out, like a radar ping. |
| `AnimateSpin` | Rotates the component continuously. |

`Animation` has no methods.

## Related

- [Transformation](../transformation/), [Stack](../../layout/stack/)
- Tutorials: [Animation](/docs/examples/tutorial-73-animation/)
