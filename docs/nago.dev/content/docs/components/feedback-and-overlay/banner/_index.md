---
# Content is auto generated
# Manual changes will be overwritten!
title: Banner
---
This component displays a prominent message to the user,
typically used for notifications, warnings, or confirmations. It consists
of a title and message, and can optionally be dismissible and styled
according to intent (e. g. , success, warning, error). It also supports a callback when the banner is closed.

## Constructors
### Banner
```go
	Banner("Nago ist great", "Give it a try.")
```

![](/images/components/feedback-and-overlay/alert/banner.png)

---
## Methods
| Method | Description |
|--------| ------------|
| `AutoCloseTimeoutOrDefault(d time.Duration)` | AutoCloseTimeoutOrDefault either takes the given duration d or timeouts after 5 seconds. |
| `Closeable(presented *core.State[bool])` | Closeable makes the banner dismissible by binding its visibility to the given state. |
| `Frame(frame ui.Frame)` | Frame sets a custom frame (layout constraints) for the banner. |
| `Intent(intent Intent)` | Intent sets the visual intent of the banner (e.g., success, warning, error). |
| `OnClosed(fn func())` | OnClosed sets a callback function that is triggered when the banner is closed. |
---
## Related
- [Banner](../../feedback-and-overlay/banner/)

