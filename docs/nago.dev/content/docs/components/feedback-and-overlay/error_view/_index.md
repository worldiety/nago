---
title: Error View
---

An error view replaces your view when it cannot be rendered because of an unexpected error, e.g. a failed
repository read. If the error is a permission error, it shows a permission hint with a back button. For any other
error it shows a generic message with a random code, a button to download an error report and a button to reload
the app; the error itself is only written to the log, together with the code. For a nil error it renders an empty
view. The error view lives in package `presentation/ui/tracking`.

```go
project, err := findProject(wnd.Subject(), id)
if err != nil {
	return tracking.ErrorView(wnd, err)
}
```

{{< callout type="warning" >}}
The texts of the error view are currently German only.
{{< /callout >}}

## Constructors

| Constructor | Description |
|---|---|
| `func ErrorView(wnd core.Window, err error) TErrorView` | Returns a view which shows the error safely instead of your actual view. |

Each call creates a new error code, so do not create error views in loops.

## Methods

| Method | Description |
|---|---|

## Related

- [Support Request Dialog](../support_request_dialog/)
- [Banner Error](../banner_error/)
