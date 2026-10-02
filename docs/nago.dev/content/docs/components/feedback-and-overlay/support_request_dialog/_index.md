---
title: Support Request Dialog
---

The support request dialog tells the user that an unexpected technical problem occurred, e.g. an infrastructure
error, which the user cannot fix. Report the error with `tracking.RequestSupport(wnd, err)` from anywhere; the
dialog then opens without showing the error details. The user can download an error report with an anonymous error
code, continue or reload the app. The code is written to the log together with the error. Place `tracking.SupportRequestDialog(wnd)` in your view tree; the
scaffold of the configurator already includes it. Both live in package `presentation/ui/tracking`.

```go
VStack(
	tracking.SupportRequestDialog(wnd),
	PrimaryButton(func() {
		if err := export(); err != nil {
			tracking.RequestSupport(wnd, err)
		}
	}).Title("Export"),
)
```

{{< callout type="warning" >}}
The texts of the dialog are currently German only.
{{< /callout >}}

## Constructors

| Constructor | Description |
|---|---|
| `func SupportRequestDialog(wnd core.Window) TSupportRequestDialog` | Creates the dialog which shows the latest error reported with `RequestSupport`; it renders nothing without errors. |
| `func RequestSupport(wnd core.Window, err error)` | Reports an unexpected error and opens the dialog. A nil error is ignored. |

Only use `RequestSupport` if you cannot offer a domain-specific hint. For errors the user can act on, show a
[Banner](../banner/) or a [Dialog](../dialog/).

## Methods

| Method | Description |
|---|---|

## Related

- [Error View](../error_view/)
- [Banner Messages](../banner_messages/) with `alert.ShowBannerError`
- Tutorials: [Download file](/docs/examples/tutorial-19-download-file/)
