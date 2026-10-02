---
title: Use Cases and Permissions
linkTitle: Use Cases
weight: 5
---

Nago separates what your application does from how it is displayed. The business logic lives in use cases,
the views only call them. All built-in systems follow this pattern, and their use cases are available to you in
the `UseCases` field of each system.

## Use cases

A use case is a named function type. Its first parameter is the `auth.Subject` on whose behalf it runs. A
constructor creates the implementation and receives its dependencies, e.g. a repository:

```go
// SayHello greets everyone who is allowed to.
type SayHello func(subject auth.Subject) (string, error)

func NewSayHello() SayHello {
	return func(subject auth.Subject) (string, error) {
		if err := subject.Audit(PermSayHello); err != nil {
			return "", err
		}

		return "hello " + subject.Name(), nil
	}
}
```

You create the use cases once in `application.Configure` and pass them to the root views which need them. In a
view, `wnd.Subject()` returns the subject of the current user:

```go
sayHello := NewSayHello()

cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
	msg, err := sayHello(wnd.Subject())
	if err != nil {
		return alert.BannerError(err)
	}

	return ui.Text(msg)
})
```

`alert.BannerError` from `go.wdy.de/nago/presentation/ui/alert` renders an error as a banner. Known errors get a
localized, user-friendly message; for all others, the user only sees a generic message and a token, while the
details go to the log.

## Subjects

`auth.Subject` (an alias of `user.Subject`) describes who acts. `wnd.Subject()` never returns nil: if nobody is
logged in, you get an anonymous subject whose `Valid()` returns false. Besides permission checks, a subject
provides `ID()`, `Name()`, `Email()`, `Roles()`, `Groups()`, `HasRole`, `HasGroup` and the language of the user.
Do not keep a subject for long, it changes when the user logs in or out.

For code which runs without a user, e.g. during configuration or in a background job, `user.SU()` returns the
system user, which has all permissions.

## Permissions

A permission protects one use case. You declare it at package level with `permission.Declare`, giving the use
case type, a unique id, a name and a description:

```go
var PermSayHello = permission.Declare[SayHello](
	"com.example.myapp.say_hello",
	"Say hello",
	"Allows to greet everyone.",
)
```

- Ids are lower case and dot separated, like `com.example.myapp.order.create`. An invalid or duplicate id panics
  at startup.
- The type parameter must be a named function type, the use case the permission belongs to.
- Permissions are defined in code only. At runtime, administrators assign them to roles or users in the
  [admin center](/docs/systems/permission_management/); they cannot create new ones.
- For CRUD-like use cases, `permission.DeclareCreate`, `DeclareFindByID`, `DeclareFindAll`, `DeclareUpdate`,
  `DeclareDeleteByID` and others generate English and German names from an entity name.

## Checking permissions

| Method                                        | Use it                                                                   |
|-----------------------------------------------|--------------------------------------------------------------------------|
| `subject.Audit(perm)`                         | in use cases; returns an error if the permission is missing              |
| `subject.HasPermission(perm)`                 | in views, to show or hide elements                                       |
| `subject.AuditResource(ns, instance, perm)`   | in use cases, for a permission on a single resource                      |
| `subject.HasResourcePermission(ns, instance, perm)` | in views, for a single resource                                    |

`Audit` is the authoritative check. Hiding a button with `HasPermission` or a [menu entry](../navigation/) only
improves the user experience; it does not protect anything.

Resource permissions grant a permission for a single instance, e.g. one document, instead of all of them. They
are stored in Nago's relation database (`cfg.RDB()`), see
[tutorial-88-resource-based-access](/docs/examples/tutorial-88-resource-based-access/).

## Trying it out

To log in during development, enable the user system and a bootstrap admin, which can log in as
`admin@localhost` with the given password until the given time:

```go
option.MustZero(cfg.StandardSystems())
std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "Only-4-Development!"))
```

The password must be strong enough, i.e. long and with upper and lower case letters, digits and special
characters; otherwise `EnableBootstrapAdmin` returns an error.

The bootstrap admin only gets the permissions of the built-in systems (ids starting with `nago.`), not yours. In
the admin center, create a role with your permissions and assign it to a user.

{{< callout type="warning" >}}
Never ship a hard-coded bootstrap password. To recover access to a production system, use the
[admin reset](../deployment/#reset-the-admin-password).
{{< /callout >}}

## Related

- [tutorial-26-buildin-iam](/docs/examples/tutorial-26-buildin-iam/)
- [User management](/docs/systems/user_management/), [Role management](/docs/systems/role_management/),
  [Permission management](/docs/systems/permission_management/)
