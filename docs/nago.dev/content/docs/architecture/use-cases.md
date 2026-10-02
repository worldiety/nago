---
title: Use Cases
weight: 2
---

A use case is one thing the system does for somebody: lend a book, list the stock, submit a quote. In this style
every use case looks the same, so a reader, a reviewer and a tool find it in the same place and can see at once
who may call it.

## The shape

A use case is a named function type whose first parameter is the acting `auth.Subject`:

```go
type LendBook func(subject auth.Subject, req LendRequest) (LendResult, error)
```

- **The subject is a parameter.** The caller has to decide who acts. A `context.Context` could be passed along
  without deciding anything.
- **The error comes last, always.** Every use case can fail authorization, so there is no use case without an
  error.
- **A streaming read returns only `iter.Seq2[T, error]`**, without a second error beside it. A sequence decides
  nothing before it is pulled, so the error is delivered through the sequence:

  ```go
  type FindAllBooks func(subject auth.Subject, filter BookFilter) iter.Seq2[Book, error]
  ```

A read is a use case like any other: its own file, constructor, permission and place in the bundle. Showing data
is a promise too, namely that this person may see it.

The same shape is what Nago's AI tools take unchanged: `completion.NewUseCaseTool` accepts
`func(auth.Subject, In) (Out, error)` and `completion.NewSeqTool` the streaming form, see
[tutorial-113-ai-assistant](/docs/examples/tutorial-113-ai-assistant/).

## One file per use case

The type and its constructor live together in `uc_<snake_case_name>.go`. The constructor is called
`New<Name>`, returns `<Name>` and receives all dependencies:

```go
// app/library/uc_lend_book.go

// LendBook hands one copy of a book to a person.
type LendBook func(subject auth.Subject, req LendRequest) (LendResult, error)

// NewLendBook builds the lending use case.
func NewLendBook(repo BookRepository) LendBook {
	return func(subject auth.Subject, req LendRequest) (LendResult, error) {
		if err := subject.Audit(PermLendBook); err != nil {
			return LendResult{}, err
		}

		optBook, err := repo.FindByID(req.Book)
		if err != nil {
			return LendResult{}, err
		}
		// ... check availability, append the borrower, save
	}
}
```

Dependencies enter through the constructor and are captured by the closure. The implementation must not read
package-level variables; constants and permission declarations are the exception.

## One permission per use case

Each use case gets its own permission, declared with the use case as type parameter and checked in the
implementation:

```go
// app/library/perm.go
var (
	PermFindAllBooks = permission.DeclareFindAll[FindAllBooks]("tutorial.library.book.find_all", "Buch")
	PermLendBook     = permission.DeclareUpdate[LendBook]("tutorial.library.book.lend", "Buch")
	PermReturnBook   = permission.DeclareUpdate[ReturnBook]("tutorial.library.book.return", "Buch")
)
```

One permission per use case is what makes authorization assignable: an administrator can grant lending without
granting returns. A permission which is declared but never checked is worse than none, because it appears in the
role editor and promises a protection that does not exist.

The name and description of a permission appear in the role editor, where a non-developer decides who may do
what. They must therefore be translatable: use a `permission.Declare<Verb>` helper (`DeclareCreate`,
`DeclareFindByID`, `DeclareFindAll`, `DeclareUpdate`, `DeclareDeleteByID`, ...), which derives English and German
texts from the entity name, or wrap your own texts in `i18n.MustString`. See
[Localization](/docs/concepts/localization/).

## Checking the subject

The implementation must consult the subject. The usual way is `subject.Audit(perm)`. Also accepted are
`AuditResource`, `HasPermission`, `HasResourcePermission`, `HasRole`, `HasGroup`, returning an error which wraps
`user.PermissionDeniedErr`, or passing the subject on to another use case which checks.

When the permission depends on the instance, e.g. one customer or one document, use `AuditResource`:

```go
if err := subject.AuditResource(Namespace, rebac.Instance(customer), PermFindQuoteOverview); err != nil {
	return QuoteOverview{}, err
}
```

See [tutorial-88-resource-based-access](/docs/examples/tutorial-88-resource-based-access/) for resource
permissions.

## Checklist

- [ ] The type is a named func type with `auth.Subject` first and `error` last, or a single
  `iter.Seq2[T, error]` result.
- [ ] Type and `New<Name>` are in `uc_<snake_case_name>.go`; `New<Name>` returns `<Name>`.
- [ ] The implementation checks the subject.
- [ ] A permission is declared with the use case as type parameter and is actually used.
- [ ] The permission texts come from i18n.
- [ ] Dependencies come through the constructor, not from package-level variables.
- [ ] The use case is a field of `UseCases` and is set in `NewUseCases`.
- [ ] An annotation file `uc_<snake_case_name>.annotation.go` names the requirements it satisfies, see
  [speclink](../../speclink/files/#annotations).

## Related

- [Use cases and permissions](/docs/concepts/use-cases-and-permissions/) explains subjects, permissions and
  auditing in Nago.
- [Project layout](../project-layout/) shows where the bundle and the views live.
