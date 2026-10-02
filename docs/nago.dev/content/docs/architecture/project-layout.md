---
title: Project Layout
linkTitle: Project layout
weight: 1
---

A project in the `go_nago_ddd1` style is one Go module with this layout:

```
speclink.json                               {"profile": "go_nago_ddd1"}
cmd/<name>/main.go                          the entry point
app/<context>/                              a bounded context: the domain
app/<context>/model.go                      aggregates, value types, requests and results
app/<context>/perm.go                       one permission per use case
app/<context>/repository.go                 what the context needs to store, not how
app/<context>/usecases.go                   the UseCases bundle and NewUseCases
app/<context>/uc_<use_case>.go              one use case per file
app/<context>/uc_<use_case>.annotation.go   which requirements it satisfies
app/<context>/ui/                           package ui<context>: the views
app/<context>/cfg/                          package cfg<context>: the wiring
pkg/, foundation/                           infrastructure without domain knowledge
requirements/                               the requirements, see speclink
```

The file names `model.go`, `perm.go` and `repository.go` and the package name of `cfg/` are conventions. The use
case files, the bundle, the package name of `ui/`, the import direction and the location of `main` are checked by
speclink.

## Entry point: `cmd/`

Every `main` package lives below `cmd/`, and the module needs at least one. `main.go` only bootstraps: it calls
`application.Configure`, enables the systems and calls the `Enable` function of each context. It contains no
business logic. In tutorial-113, `cmd/ai-example/main.go` does exactly that.

## Bounded contexts: `app/<context>/`

A bounded context is a part of the business with its own language, e.g. `library`, `sales` or `billing`. Its
package declares the model and the use cases and exports them as one bundle:

```go
// app/library/usecases.go
type UseCases struct {
	FindAllBooks FindAllBooks
	LendBook     LendBook
	ReturnBook   ReturnBook
}

func NewUseCases(repo BookRepository) UseCases {
	return UseCases{
		FindAllBooks: NewFindAllBooks(repo),
		LendBook:     NewLendBook(repo),
		ReturnBook:   NewReturnBook(repo),
	}
}
```

Every use case of the context must be a field of `UseCases`. Callers depend on the bundle, not on the internals,
and `NewUseCases` is the one place where shared dependencies like the repository are passed in.

The domain package must not import a user interface package: nothing below Nago's `presentation/` and no package
whose name starts with `ui`. Pass what the domain needs as a parameter instead.

## Views: `app/<context>/ui/`

The `ui` directory declares the package `ui<context>`, e.g. `uilibrary`. The directory is always called `ui`, so
the package name is what tells you and the imports which context it belongs to. Packages nested below `ui/` may be
named freely.

Views call use cases from the bundle, they never read a repository. Showing or hiding a button with
`HasPermission` is fine; the check that protects the data is the `Audit` in the use case.

## Other ways in, e.g. `app/<context>/ai/`

tutorial-113 puts its AI tools into `app/library/ai`, package `ailibrary`, because a model is one more way to reach
the context, like a screen. speclink only knows `ui/` and `cfg/` as special directories; any other subpackage of a
context is checked like domain code, so it must not import a `ui*` package either.

## Wiring: `app/<context>/cfg/`

The `cfg` package is the only place where storage, domain and views meet, and it is exempt from the import rule.
A typical `Enable` function opens the store, builds the use cases, declares roles and registers the pages:

```go
// app/library/cfg/cfg.go
func Enable(cfg *application.Configurator) (Module, error) {
	store, err := cfg.EntityStore("tutorial.library.book")
	if err != nil {
		return Module{}, fmt.Errorf("cannot open book store: %w", err)
	}

	repo := library.BookRepository(json.NewSloppyJSONRepository[library.Book, library.BookID](store))
	uc := library.NewUseCases(repo)

	// declare roles, register pages ...
	return Module{UseCases: uc, Pages: pages}, nil
}
```

The `Module` it returns is what the rest of the application, e.g. `main.go` or another context's `cfg`, depends
on.

## Infrastructure: `pkg/` and `foundation/`

Code which is useful independent of the business, e.g. a parser or a client for an external API, lives in `pkg/`
or `foundation/`. Infrastructure must not import a bounded context and must not declare a use case: the domain
builds on infrastructure, never the other way round.

## Requirements: `requirements/`

The requirements live next to the code, as Go files. Their layout is described in
[What you write](../../speclink/files/).

## Deviating from the defaults

`speclink.json` only states deviations from the profile. A project which follows the layout contains nothing but
the profile name. For a different layout, the profile understands `contextRoot` (default `app`), `cmdRoot`
(default `cmd`) and `infraRoots` (default `pkg` and `foundation`), plus `sourceRoots`, `scope` and `exclude`.
Unknown keys are refused, not ignored.

## Related

- [tutorial-113-ai-assistant](/docs/examples/tutorial-113-ai-assistant/) shows the complete layout.
- [Application and configurator](/docs/concepts/application/) explains `application.Configure` and the
  configurator.
