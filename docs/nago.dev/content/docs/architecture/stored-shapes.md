---
title: Stored Shapes
linkTitle: Stored shapes
weight: 4
---

A type which is written to disk is a promise. An event is replayed from the log forever, a JSON document is read
back by the next release. If a field is renamed or removed, the code still compiles and the tests still pass,
because they only see new data. What breaks is the data written before, and you notice in production.

speclink therefore treats every persisted type as **frozen by default** and reports incompatible changes as
errors. This page explains what counts as persisted, what you may change, and how a shape grows.

## What is persisted

Two kinds of types, and only these:

| Type                  | Why it is persisted                                                         |
|-----------------------|-----------------------------------------------------------------------------|
| an event              | it implements `Evolve` and `Discriminator`; the struct is the stored form  |
| a persistence model   | the type a JSON repository was built over                                   |

Nothing in a struct says that it is stored; the repository constructor does. Nago offers two:

```go
// Two models, mapped. Only CustomerEntity is promised; Customer stays free to change.
json.NewJSONRepository[Customer, CustomerID, CustomerEntity, CustomerID](store, intoDomain, intoPersistence)

// One model. The domain type IS the stored form and is promised as it stands.
json.NewSloppyJSONRepository[Book, BookID](store)
```

With the sloppy form every rename in the domain type is a change to stored data. tutorial-113 uses it because it
is a small example. For anything which outlives a prototype, prefer the mapped form: the domain model can then be
restructured without touching a byte on disk.

## Drafts

While a type is still in development, mark it as a draft in an annotation file. Everything persisted is frozen
unless it is marked:

```go
var _ = spec.ForPackage(spec.Draft())          // every persisted type of the package
var _ = spec.For[QuoteWithdrawn](spec.Draft()) // one type
var _ = spec.ForField[Quote]("Note", spec.Draft()) // one field
```

`spec.Draft()` means one thing: *we are willing to delete every stored value of this type.* Remove it as soon as
nobody would do that any more. That moment has nothing to do with going live; a development database can already
hold data somebody minds losing.

## What you may change once frozen

| Change                                         | Allowed? |
|------------------------------------------------|----------|
| change the discriminator of an event           | no, it orphans every stored message |
| two persisted types with the same discriminator| no, not even as drafts |
| remove a field                                 | no |
| change the stored name (JSON tag) of a field   | no; renaming the Go field is fine |
| change a field's shape incompatibly, e.g. `int` to `string` | no |
| `string` to a named string type, or another integer width | yes |
| add a field                                    | yes, marked `spec.Optional()` |
| make an optional field required again          | no |

A new field must be optional, because values written before it existed do not carry it:

```go
var _ = spec.ForField[QuoteSubmitted]("Channel",
	spec.Optional(),
)
```

## speclink.lock

The current source cannot tell what a field used to be. The promise is therefore recorded in `speclink.lock`,
written by `speclink freeze` and never edited by hand, similar to `go.sum`. Committing to a shape is two steps:
delete the `spec.Draft()` term, then run

```bash
speclink freeze -n ./...   # show what would be recorded
speclink freeze ./...      # record it
```

The diff of `speclink.lock` in the merge request is what a reviewer reads. `freeze` refuses to record a shape
which already breaks a promise, and a recorded shape cannot be turned back into a draft.

## Related

- [Persistence patterns](../persistence/)
- [The speclink workflow](../../speclink/workflow/)
