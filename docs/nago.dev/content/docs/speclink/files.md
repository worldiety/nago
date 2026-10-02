---
title: What You Write
linkTitle: What you write
weight: 2
---

A speclink project consists of ordinary Go code plus a few kinds of files. All Go files below are part of the
normal build, so the compiler checks every reference before speclink even runs.

| File                                     | Content                                               |
|------------------------------------------|-------------------------------------------------------|
| `speclink.json`                          | the profile, e.g. `{"profile": "go_nago_ddd1"}`       |
| `requirements/_sources/*.md`             | the source documents people wrote                     |
| `requirements/**/<ID>.spec.go`           | one requirement per file                              |
| `<base>.annotation.go`                   | which requirements the code in `<base>.go` satisfies  |
| `*_test.go` with `spec.Verified`         | which requirements a test demonstrates                |
| `speclink.lock`                          | written by speclink, never by hand                    |

## Source documents

The requirement tree is not the top of the chain. Above it are the documents people wrote: Markdown files and
mockup images, by default in `requirements/_sources/`. A Markdown file is divided into segments by its headings,
and a requirement points at a heading by its slug, e.g. `## 8. Abgabe` becomes `8-abgabe`.

Every segment has to lead to at least one requirement. A section without obligations, e.g. an introduction, says so
where it is written:

```markdown
# Einleitung

<!-- speclink:informative -->
```

PDF is not supported. Convert it to Markdown; the conversion is then visible and diffable in the merge request.
Mockups are PNG or JPEG images with a sidecar `<image>.speclink.json` which names regions, so that a requirement
can point at one button of a screen.

## Requirements

One requirement per file, named after its ID, declared with `spec.Declare` from
`github.com/worldiety/speclink/spec`:

```go
// requirements/fun/library/R-LIB-LEND.spec.go, from tutorial-113 with a source added
package library

import "github.com/worldiety/speclink/spec"

var RLibLend = spec.Declare(spec.Requirement{
	ID:         "R-LIB-LEND",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Exemplar ausleihen",
	Text:       "Ein freies Exemplar MUSS an eine namentlich genannte Person ausgeliehen werden können; ist keines frei, MUSS die Ausleihe abgelehnt werden.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/library.md", Anchor: "ausleihe"},
	},
})
```

Directory, package name and ID prefix are one fact and must agree: `R-LIB-LEND` belongs into `fun/lib/` (package
`lib`), or its ID must be `R-LIBRARY-LEND` when it stays in `fun/library/`. tutorial-113 does not follow this rule
yet, and `speclink verify` reports it as `SPEC-V5-035`.

`spec.Declare` returns its argument and registers it in a catalogue which the running program can read with
`spec.Requirements()`. That is how Nago's [requirement catalogue](/docs/systems/speclink_management/) works.

Directory, ID prefix and `Kind` must agree:

| Directory                                   | ID                  | Kind                 |
|---------------------------------------------|---------------------|----------------------|
| `requirements/fun/<domain>/`                | `R-<DOMAIN>-<NAME>` | `spec.Functional`    |
| `requirements/dec/`                         | `R-DEC-<NAME>`      | `spec.Decision`      |
| `requirements/nfr/`                         | `R-NFR-<NAME>`      | `spec.NonFunctional` |
| `requirements/cst/`                         | `R-CST-<NAME>`      | `spec.Constraint`    |

The most important fields:

- `Text` is the normative statement in one sentence. Longer explanations go into a Markdown file named by `Detail`.
- `Status`: only `spec.Normative` requirements must be implemented and tested. `spec.Planned` ones are known but
  not built yet, `spec.Informative` ones explain, `spec.Superseded` ones must no longer be satisfied.
- `Sources` is mandatory for normative requirements: a `Doc` with an `Anchor` in your source documents, or
  `Extern` for a law or standard without a document in the repository, e.g. `"HGB §§ 383 ff."`.
- A `spec.Decision` must have a `Rationale` **and** `Consequences`: why it was decided, and what it costs.
- `DerivedFrom` and `Supersedes` reference other requirements by their Go variable, so moving a file breaks
  nothing.
- `Disclosure` (`Public`, `Internal`, `Confidential`, `Secret`, default `Public`) states who may see the text.
  speclink does not enforce it; Nago's requirement catalogue does.

Write texts longer than one line as a raw string, not as a concatenation with `+`, so that a sentence can still be
found with grep.

## Annotations

An annotation file sits next to the file it annotates, `uc_lend_book.go` gets `uc_lend_book.annotation.go`, in the
same package:

```go
var _ = spec.For[LendBook](
	spec.Satisfies(fun.RLibLend, dec.RDecBorrower),
	spec.Help("Gibt ein Exemplar an eine Person heraus. Ist keines frei, wird die Ausleihe abgelehnt."),
)
```

Because it is part of the normal build, an annotation for a use case which no longer exists breaks the build. An
annotation file may contain only imports and `var _ = spec.X(...)` terms: no functions, no types, no computation.
This keeps it readable without running it.

| Binding                         | Target                                                   |
|---------------------------------|----------------------------------------------------------|
| `spec.For[T](...)`              | a named type: use case, event, aggregate, projection     |
| `spec.ForDecl(ref, ...)`        | a declared function, variable or constant                |
| `spec.ForField[T]("Name", ...)` | one struct field                                         |
| `spec.ForPackage(...)`          | the whole package                                        |

The assertions you will use most:

| Assertion                  | Meaning                                                            |
|----------------------------|--------------------------------------------------------------------|
| `spec.Satisfies(reqs...)`  | this code was written for these requirements                       |
| `spec.Help(text)`          | an end-user explanation, e.g. for help texts and the AI assistant  |
| `spec.Rationale(text)`     | why a decision is implemented this way here                        |
| `spec.Draft()`             | this persisted shape is not promised yet, see [Stored shapes](../../architecture/stored-shapes/) |
| `spec.Optional()`          | this field may be missing in stored data                           |
| `spec.Waive(rule, reason)` | suspend one rule here; the reason is mandatory                     |

### What must name a requirement

speclink finds the constructs itself. You only bind those which carry business meaning:

| You write                                                     | speclink sees | Needs `Satisfies`? |
|---------------------------------------------------------------|---------------|--------------------|
| named func type with `auth.Subject` first                     | use case      | yes                |
| a type with `Decide(auth.Subject, *Agg) ([]Evt, error)`       | command       | yes                |
| a type with `Evolve` and `Discriminator`                      | event         | yes                |
| the state of `evs.NewProjection` or `evs.NewSingleton`        | projection    | yes                |
| a type with `Identity()`                                      | aggregate     | no                 |
| `permission.Declare...[UseCase](...)`                         | permission    | no                 |
| a type over `data.Repository` or `data.ReadRepository`        | repository    | no                 |

Aggregates, permissions and repositories are covered through the use cases which guard, write or read them.

## Tests

At the **end** of a test, state which requirement it demonstrated:

```go
func TestLendRefusesWhenNothingIsFree(t *testing.T) {
	// ... lend the last copy, then try again and expect an error ...

	spec.Verified(t, fun.RLibLend)
}
```

The position matters: `spec.Verified` writes a line when it runs, so at the end it says the test got there.
speclink reads the call from the source, which makes a missing one reportable, and reads the line from the test
output, which makes a present one believable.

A requirement which cannot be tested, e.g. a structural decision, is waived on a construct which satisfies it:

```go
var _ = spec.For[Customer](
	spec.Satisfies(dec.RDecCustomerState),
	spec.Waive("K14-REQ-UNVERIFIED", "The decision is that a customer is stored as state, which the type itself shows."),
)
```

## speclink.lock

`speclink.lock` records what has been promised and what has happened: the wording of requirements and source
sections, the shapes of persisted types, which tests demonstrated which requirement, and who reviewed what. It is
written by `speclink freeze`, `speclink evidence` and `speclink attest`, and you commit it. Its diff is what a
reviewer reads.

## Processes and topology

Larger projects can also describe their business processes (`<name>.process.go` with `spec.Process`) and the
systems and channels around them (`<name>.topology.go` with `spec.Actor`, `spec.Foreign` and `spec.Channel`).
speclink checks them and draws them as diagrams in the generated document. They are optional; see the
[speclink README](https://github.com/worldiety/speclink) when you need them.
