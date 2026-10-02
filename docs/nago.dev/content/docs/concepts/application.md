---
title: Application and Configurator
linkTitle: Application
weight: 1
---

Every Nago program has the same shape:

```go
func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("com.example.myapp")
		cfg.Serve(vuejs.Dist())

		// enable systems, create services, register root views ...
	}).Run()
}
```

## Lifecycle

1. `application.Configure` loads a `.env` file from the working directory, if one exists, reads the
   [environment variables](../deployment/) and then calls your function with a fresh
   `*application.Configurator`.
2. In your function you configure everything. This happens exactly once, before the server starts, so it is the
   place to create stores, enable systems and wire your use cases.
3. `Run` applies pending migrations, starts the web server and blocks until the process receives `SIGINT` or
   `SIGTERM`. Then it shuts the application down and calls the functions registered with `cfg.OnDestroy`.

Errors during configuration are usually fatal. The tutorials therefore unwrap results with `std.Must` from
`go.wdy.de/nago/pkg/std`, or `option.Must` and `option.MustZero` from `github.com/worldiety/option`, which panic on
an error.

## What you configure

| Concern                     | Configurator methods                                                                      |
|-----------------------------|-------------------------------------------------------------------------------------------|
| identity                    | `SetApplicationID`, `SetName`, `SetVersion`, `SetSemanticVersion`, `AppIcon`              |
| frontend                    | `Serve(vuejs.Dist())`                                                                     |
| pages                       | `RootView`, `RootViewWithDecoration`, `SetDecorator`, `NewScaffold`, see [Navigation](../navigation/) |
| static resources            | `Resource`, e.g. for embedded images                                                      |
| data                        | `DataDir`, `EntityStore`, `FileStore`, `NDB`, see [Persistence](../persistence/)          |
| theme                       | `ColorSet`, `ThemeManagement`, see [Theming](../theming/)                                 |
| systems                     | `StandardSystems`, `UserManagement`, `SessionManagement`, ..., see [Systems](/docs/systems/) |
| services                    | `AddContextValue`, `Context`, `EventBus`                                                  |
| server                      | `SetHost`, `SetFPS`, `SetContextPath`, `Debug`                                            |

### Static resources

Embed files with `go:embed` into an `application.StaticBytes` and register them with `cfg.Resource`. You get a
URI which you can pass to views such as `Image().URI(...)`:

```go
//go:embed logo.jpg
var logo application.StaticBytes

// within Configure
logoURI := cfg.Resource(logo)
```

See [tutorial-02-combining-views](/docs/examples/tutorial-02-combining-views/).

## Systems

Nago's built-in features are organized as systems. Each one is enabled by a method of the configurator or by an
`Enable` function in a `cfg` package below `application/`:

```go
option.MustZero(cfg.StandardSystems())                 // admin center, users, sessions, mail, backup, ...
users := std.Must(cfg.UserManagement())                // a single system, returns its use cases and pages
std.Must(cfgscheduler.Enable(cfg))                     // go.wdy.de/nago/application/scheduler/cfg
```

Enabling a system is idempotent: calling the method again returns the already configured instance. Systems
enable the systems they depend on themselves. `StandardSystems` enables the image, admin, user, backup, mail,
secret, template and session management. Each system returns a struct with its `UseCases`, which you can call
from your own code. The [Systems](/docs/systems/) section describes them one by one.

## Your own services

Prefer to create your use cases in `Configure` and pass them explicitly to the views which need them. A
closure over a local variable is all you need:

```go
sayHello := NewSayHello()

cfg.RootView(".", func(wnd core.Window) core.View {
	return ui.Text(sayHello(wnd.Subject()))
})
```

For loosely coupled components, the configurator also provides a small service registry. Register a value with
`cfg.AddContextValue(core.ContextValue("", myService))` and look it up by its exact type with
`core.FromContext[MyService](wnd.Context(), "")`. The built-in systems use this mechanism, e.g. to find the
optional mail system. Use it sparingly, explicit wiring is easier to follow.

`cfg.EventBus()` returns an application-wide event bus to publish and subscribe to domain events, see
[tutorial-47-eventbus](/docs/examples/tutorial-47-eventbus/).

## Without a web server

`cfg.OneShot()` makes `Run` return right after the configuration instead of starting the server. This is useful
for command line tools which reuse the configuration and data of an application, like the
[admin reset tool](../deployment/#reset-the-admin-password).

## Related

- [tutorial-01-helloworld](/docs/examples/tutorial-01-helloworld/)
- [tutorial-42-service](/docs/examples/tutorial-42-service/) – a background service using the scheduler
- [Configuration and deployment](../deployment/)
