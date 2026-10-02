---
title: Persistence Patterns
linkTitle: Persistence patterns
weight: 3
---

How an aggregate is stored is a decision you make **per aggregate**, driven by its requirements, not once for the
whole project. The style knows three patterns, and they are meant to be combined. The underlying Nago APIs are
described in [Persistence](/docs/concepts/persistence/).

## Aggregates

An aggregate is a cluster of data which is changed as a unit and has an identity, e.g. a book with its copies
and borrowers. In Nago, a type becomes an aggregate by implementing `Identity()`:

```go
type BookID string

type Book struct {
	ID     BookID   `json:"id"`
	Title  string   `json:"title"`
	Copies int      `json:"copies"`
	LentTo []string `json:"lentTo,omitempty"`
}

func (b Book) Identity() BookID { return b.ID }
```

Keep values which can be derived out of the aggregate. tutorial-113 computes the free copies as
`Copies - len(LentTo)` instead of storing them, and records that as a decision requirement
(`R-DEC-AVAILABILITY`): a derived value which is also stored can disagree with itself.

## Repository: current state

The repository stores the current state of an aggregate. The context declares the repository it needs as a type
over `data.Repository` and leaves the implementation to `cfg/`:

```go
// app/library/repository.go
type BookRepository = data.Repository[Book, BookID]
```

Depend on `data.ReadRepository[E, ID]` when a use case only reads; the missing write access is then visible in
the signature.

Things which are easy to get wrong:

- `FindByID` returns an `option.Opt[E]`, not a zero value. Unwrap it explicitly.
- `All`, `FindAllByPrefix`, `FindAllByID` and the identifier traversals are `iter.Seq2[E, error]`. Handle the
  error of each element.
- Never call back into the same repository from inside a `yield`; most implementations deadlock.
- Prefix iteration is lexicographic, so prefix queries on integer keys do not mean what they look like.

**Choose a repository when** the aggregate is current state, its history is not a requirement, or the data is
unbounded or owned by someone else.

## Decide/evolve: event sourcing

The write side of event sourcing. A command *decides* against the current state and returns events; each event
*evolves* the aggregate. The events are the stored truth, the aggregate is folded from them.

```go
// The command checks and returns facts. It must not change the aggregate.
func (cmd IntroduceEmployeeCmd) Decide(subject auth.Subject, aggregate *EmployeeAggregate) ([]EmployeeEvent, error) {
	// validate, audit ...
	return []EmployeeEvent{EmployeeIntroduced{EmployeeID: cmd.ID, Name: cmd.Name}}, nil
}

// The event folds itself into the aggregate.
func (e EmployeeIntroduced) Evolve(ctx context.Context, aggregate *EmployeeAggregate) error {
	aggregate.ID = e.EmployeeID
	aggregate.Name = e.Name
	return nil
}

// The discriminator is the tag under which the event is stored.
func (e EmployeeIntroduced) Discriminator() evs.Discriminator {
	return "EmployeeIntroduced"
}
```

The aggregate is a pointer type with `Clone()` and `IsDeleted()`. `cfgevs.NewHandler` from
`go.wdy.de/nago/application/evs/cfg` creates the handler, and a use case passes its command to
`handler.Handle(subject, id, cmd)`. The complete example is
[tutorial-85-eventsourcing-decide-evolve](/docs/examples/tutorial-85-eventsourcing-decide-evolve/).

- **The discriminator is permanent.** The log is replayed forever; changing the tag of an existing event orphans
  every message written under it. Give it a namespace and a version, e.g. `"sales.quote.submitted.v1"`, so it
  cannot collide with another context and survives renaming the Go type. Add a `.v2` type instead of editing
  `.v1`.
- **Choose decide/evolve when** the history itself is a requirement, decisions must be reconstructible, or the
  invariants are per aggregate.
- **The cost:** every schema change becomes a new event version, and the log is replayed at start.

## Projections: the read side

`Evolve` targets exactly one aggregate. A list or dashboard which spans aggregates, or needs a different key,
is a projection: it folds the same events into its own state.

```go
func newQuoteOverview(src evs.Source) *evs.Singleton[*QuoteOverview] {
	p := evs.NewSingleton[*QuoteOverview](src, evs.ProjectionOptions{})
	evs.Project(p,
		func(QuoteSubmitted) evs.Unit { return evs.TheUnit() },
		func(s *QuoteOverview, e QuoteSubmitted) {
			s.Submitted++
			s.LastQuote = e.QuoteNumber
		},
	)
	return p
}
```

Use `evs.NewProjection[K, *S]` when there is a key and `evs.NewSingleton[*S]` when everything folds into one value.
The state type needs a `Clone()` method, because readers get a deep copy. `Run()` replays the history once and then
follows new events; read with `Get` or `All`.

- **A projection has no storage of its own.** It is rebuilt by constructing it again. Persisting it would turn a
  derived view into a second truth.
- **It is not read-your-write.** Folding is asynchronous. If a caller must see its own write, wait for the
  sequence number the write returned with `WaitFor`. For an interactive screen this rarely matters, because a
  re-render follows the event.
- **Choose a projection when** a query crosses aggregates, needs another key, or feeds a list or dashboard. Never
  answer a query by replaying the log inside the request.

If you choose decide/evolve, you almost always need projections as well.

## Mixing them

Common and legitimate in one project:

- an event-sourced core aggregate, because its history is a requirement,
- repositories for reference and master data, where history is noise,
- projections for every screen which spans aggregates.

Not legitimate:

- writing the same aggregate through a repository and through an event log: two truths which will diverge,
- giving a projection its own persistence: it must stay rebuildable,
- reading a repository or a handler from a `ui` package: views call use cases.

Write the choice down as a decision requirement with a rationale and its consequences, so that it does not
become tribal knowledge. See [What you write](../../speclink/files/#requirements).

## Why generic CRUD is not available

Nago can generate repositories, use cases, permissions and admin pages for an entity with `cfgent.Enable`
(see [Persistence](/docs/concepts/persistence/#entities-with-generated-ui)). In this style that is not allowed;
speclink rejects `cfgent.Enable`, `cfgent.EnableUseCases`, `ent.DeclarePermissions`, `ent.NewUseCases` and imports
of `go.wdy.de/nago/application/ent/ui`.

The reason is not the design of these screens but how they are built: the helpers derive permissions, use cases
and routes at run time from a prefix. These facts only exist while the program runs, so no requirement can ever
be traced to the code which satisfies it, and no reviewer can find the use case in a file.

Write the use cases by hand, declare one permission each and bind them to their requirements. The screens which
`uient` renders remain a good reference for how a Nago admin UI should look; take that style into your own views.

## Related

- [Stored shapes](../stored-shapes/): what happens when a persisted type changes.
- [tutorial-103-ndb](/docs/examples/tutorial-103-ndb/): the storage engine below the event log.
