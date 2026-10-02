---
title: Example
weight: 4
---

[tutorial-113-ai-assistant](/docs/examples/tutorial-113-ai-assistant/) is a small library application in the
`go_nago_ddd1` style. This page follows one requirement through it: lending a book.

## 1. The requirement

The functional requirement lives in the requirement tree, in the directory of its domain:

```go
// requirements/fun/library/R-LIB-LEND.spec.go
var RLibLend = spec.Declare(spec.Requirement{
	ID:         "R-LIB-LEND",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Exemplar ausleihen",
	Text:       "Ein freies Exemplar MUSS an eine namentlich genannte Person ausgeliehen werden können; ist keines frei, MUSS die Ausleihe abgelehnt werden.",
})
```

## 2. The decision behind it

Lending records the borrower as a plain name, not as a user account. That is not obvious, so it is written down as
a decision, together with what it costs:

```go
// requirements/dec/R-DEC-BORROWER.spec.go
var RDecBorrower = spec.Declare(spec.Requirement{
	ID:           "R-DEC-BORROWER",
	Kind:         spec.Decision,
	Discipline:   spec.Business,
	Status:       spec.Normative,
	Title:        "Ausleiher werden als Name geführt, nicht als Benutzerkonto",
	Text:         "LentTo enthält Klartextnamen; es gibt keine Verknüpfung zur Benutzerverwaltung.",
	Rationale:    "Eine Bibliothek leiht auch an Menschen aus, die kein Konto in diesem System haben …",
	Consequences: "Zwei Personen gleichen Namens sind nicht unterscheidbar, und es gibt keine automatische Erinnerung …",
})
```

Three years later, somebody who wants to link borrowers to accounts reads first why it was not done, and what
reversing the decision would change.

## 3. The code

The use case follows the [architecture](../../architecture/use-cases/): its own file, a constructor, its own
permission:

```go
// app/library/uc_lend_book.go
type LendBook func(subject auth.Subject, req LendRequest) (LendResult, error)

func NewLendBook(repo BookRepository) LendBook {
	return func(subject auth.Subject, req LendRequest) (LendResult, error) {
		if err := subject.Audit(PermLendBook); err != nil {
			return LendResult{}, err
		}
		// ... refuse if no copy is free, otherwise record the borrower
	}
}
```

## 4. The link

The annotation file next to it states the one fact speclink cannot infer, which requirements this use case was
written for, plus a help text for users:

```go
// app/library/uc_lend_book.annotation.go
var _ = spec.For[LendBook](
	spec.Satisfies(fun.RLibLend, dec.RDecBorrower),
	spec.Help("Gibt ein Exemplar an eine Person heraus. Ist keines frei, wird die Ausleihe abgelehnt — eine Vormerkung gibt es noch nicht."),
)
```

That `LendBook` is a use case, that `PermLendBook` guards it and that `BookRepository` is a repository, speclink
finds out by itself.

## 5. What is planned, but not built

Reserving a lent-out book is known, but not built yet. The requirement exists with `Status: spec.Planned` and is
bound to nothing:

```go
// requirements/fun/library/R-LIB-RESERVATION.spec.go
var RLibReservation = spec.Declare(spec.Requirement{
	ID:     "R-LIB-RESERVATION",
	Kind:   spec.Functional,
	Status: spec.Planned,
	Title:  "Vormerkung auf ein ausgeliehenes Exemplar",
	// ...
})
```

speclink does not demand code for it. In the running application, the AI assistant can therefore answer "not
implemented yet" instead of "does not exist".

## 6. At run time

The requirement packages are imported by the annotation files and thus linked into the binary. With
`cfgspeclink.Enable(cfg)`, Nago shows them in the admin center, and the assistant of tutorial-113 reads them
through tools. Ask it why borrowers are names and not accounts, and it answers with `R-DEC-BORROWER` and its
consequences, see [Requirements](/docs/systems/speclink_management/).

## What a checked project adds

The tutorial declares requirements and annotations, but the Nago repository does not run `speclink verify` on its
examples. Running it in the tutorial directory shows what is still missing:

```bash
cd example/cmd/tutorial-113-ai-assistant
speclink verify -profile go_nago_ddd1 ./...
```

It ends with 24 findings:

- `SPEC-V5-020`: a normative requirement names no source. Fixed by a source document in `requirements/_sources/`
  and `Sources` on every normative requirement (`SPEC-V5-027` reports the missing directory itself).
- `SPEC-V6-120`: no test demonstrates the requirement. Fixed by tests which end with `spec.Verified(t, …)`, handed
  to `speclink evidence`.
- `SPEC-V5-035`: `R-LIB-*` lies in `fun/library/`. Fixed by moving the files to `fun/lib/` or renaming the IDs to
  `R-LIBRARY-*`.
- `SPEC-V6-021`: the aggregate `Book` rests on no recorded decision about how it is stored. Fixed by writing a
  decision requirement which records the choice, e.g. a new `R-DEC-STORAGE`, and binding it with
  `spec.For[Book](spec.Satisfies(dec.RDecStorage))`.
- `SPEC-V6-022`: stored fields like `Book.Title` trace to no requirement. Fixed with
  `spec.ForField[Book]("Title", spec.Satisfies(…))` for each field.
- `SPEC-V6-090`: `Book` is persisted, but its shape was never recorded. Fixed by `speclink freeze`, or by marking
  it `spec.Draft()` while it is still in flux, see [stored shapes](/docs/architecture/stored-shapes/).

The use cases themselves pass: each names a requirement, and each requirement is implemented. A real project also
keeps `speclink.json` with `{"profile": "go_nago_ddd1"}` and `speclink.lock` in version control.
