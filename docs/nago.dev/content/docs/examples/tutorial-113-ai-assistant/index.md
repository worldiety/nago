---
title: AI Assistant
weight: 113
---

An AI assistant that operates the use cases of an application, with the permissions of the signed-in user and
nothing more.

[Tutorial 77](/docs/examples/tutorial-77-ai/) shows the mechanics of the tool loop. This tutorial shows the
conventions: how a real business application exposes its use cases to a model, how it ships its authorization,
and what keeps a model from changing something nobody asked for. The example is a small library app; the
[AI system](/docs/systems/ai_management/) and the [requirement catalogue](/docs/systems/speclink_management/)
come from the framework.

The assistant needs a configured AI provider, e.g. an Anthropic token in the admin center, and the roles
"Bibliothekar" and "AI Assistant User". Without them the app runs without the assistant button.

## Structure

Unlike the other tutorials, which put everything into one `main.go`, this one follows the `go_nago_ddd1` profile
of [speclink](https://github.com/worldiety/speclink), the traceability tool worldiety uses to check Nago
projects. This structure is the reason why the assistant below needs no adapter at all. See
[Architecture](/docs/architecture/) and [speclink](/docs/speclink/) for the style and the tool.

```
cmd/ai-example/main.go          entry point: bootstrap and wiring, no business logic
app/library/                    the bounded context: what the system does
  model.go                      aggregate, filter, request, result
  perm.go                       one permission per use case
  repository.go                 what the context needs to store, not how
  usecases.go                   the UseCases bundle
  uc_find_all_books.go          one use case per file, type and constructor together
  uc_lend_book.go
  uc_return_book.go
  uc_*.annotation.go            binds a use case to its requirements
  ui/page_books.go              package uilibrary: the view for humans
  ai/tools.go                   package ailibrary: the view for a model
  cfg/cfg.go                    package cfglibrary: the only place where both meet
requirements/
  dec/R-DEC-*.spec.go           the decisions: why the library works the way it does
  fun/library/R-LIB-*.spec.go   what it must do
```

Dependencies point inwards only: `ui` and `ai` import `library`, `library` imports neither. A context which
imports its own UI cannot be tested without a renderer. `ai/` sits next to `ui/` on purpose: a model is one way
to reach the context, just as a screen is, and neither belongs to the domain.

Because the entry point lives under `cmd/`, the run command is one path segment longer than for the other
tutorials:

```bash
go run go.wdy.de/nago/example/cmd/tutorial-113-ai-assistant/cmd/ai-example@latest
```

The example uses `github.com/worldiety/speclink/spec`, a module without dependencies which only contains the
declarations. The `speclink` tool itself (`speclink.json`, `speclink verify`, the static checks) is not part of
it, because a compiler frontend does not belong into the module graph of an application. So the requirements are
declared and read at runtime, but not checked statically. A real project adds that.

## Why the tools need no adapter

A use case in this profile has the shape

```go
func(subject auth.Subject, in In) (Out, error)
```

This is not a convenience for the AI but the architecture rule of the project: the subject is a parameter, so
the caller has to decide who is acting. And it is exactly the shape a tool needs, so
`completion.NewUseCaseTool` takes the use case unchanged:

```go
lend := completion.NewUseCaseTool("lend_book",
	"Leiht ein Exemplar eines Buches an eine Person aus. Nenne vorher Titel und Person, damit klar ist, was passiert.",
	uc.LendBook)
```

No adapter, no subject at construction time, no rebuild per request. The tools are built once at start-up and
shared by every window and every user. An application cut differently has to write a wrapper for every tool,
and has to write the authorization a second time by hand.

## Seven rules

These rules are what the example is about; `app/library/ai/tools.go` follows them.

### 1. A tool goes through a use case, never through a repository

This is the whole authorization model. A Nago use case audits the given subject, the same one a click in the UI
would be checked against. `NewUseCaseTool` passes the subject of the asking user through, so the assistant
cannot read what this user may not read, and cannot write what they may not write. A tool which bypasses the use
case and reads the repository directly breaks this unnoticed, because it works.

### 2. Use cases need no wrapper

`NewUseCaseTool` takes `func(auth.Subject, In) (Out, error)`. For the second common shape,
`func(auth.Subject, Filter) iter.Seq2[T, error]`, there is `completion.NewSeqTool` (see rule 4). If you really
need a function of your own, use `completion.NewSubjectTool`, which also receives the acting subject.

### 3. Mark writing tools, do not describe them

```go
lend := completion.NewUseCaseTool(...).
	AsMutating("gibt ein Exemplar heraus und trägt die Person als Ausleiher ein")
```

A sentence in the system prompt ("always ask before you write") is a request to the model. The mark is a gate in
front of it: unless the operator switched the confirmation off in the assistant settings, the chat holds the
call and shows the tool, its effect and the arguments the model chose before anything happens. The operator
setting *Nur lesender Zugriff* (read-only) removes writing tools completely, so the model does not even learn
that they exist.

The text passed to `AsMutating` is what the user reads before confirming. An empty text makes the dialog
worthless.

### 4. Bound every listing

```go
list := completion.NewSeqTool("list_books", "…", uc.FindAllBooks)
```

`NewSeqTool` adds a `limit` property to the schema (default `completion.DefaultSeqToolLimit` = 200, at most
`completion.MaxSeqToolLimit` = 1000) and wraps the result in `items`, `count`, `truncated` and a plain-text
`note`. Without it, a `FindAll` over a few thousand records exhausts the context window, and the error looks
like a provider problem rather than a missing limit. A list cut short silently is worse than one cut short
openly: the model should narrow its filter instead of drawing conclusions from half an answer.

### 5. Names and descriptions are part of the interface

Tool names are `lower_snake_case`, and descriptions are written for someone who has never seen the application.
Both are validated at construction: an invalid name or an empty description panics at start-up instead of
confusing the model at runtime.

The same holds for the `desc` tags of the input type; they are the documentation the model reads:

```go
type BookFilter struct {
	Query         string `json:"query" optional:"true" desc:"case-insensitive substring matched against title and author"`
	OnlyAvailable bool   `json:"onlyAvailable" optional:"true" desc:"when true, only books with at least one free copy are returned"`
	Borrower      string `json:"borrower" optional:"true" desc:"only books currently lent to this borrower"`
}
```

### 6. The domain type is the tool type

A separate DTO next to the aggregate is usually unnecessary. Put the `desc` tags on the domain type, just as
`label` tags sit where `form.Auto` reads them:

```go
type Book struct {
	ID     BookID   `json:"id" desc:"stable identifier, used when lending or returning"`
	Copies int      `json:"copies" desc:"total number of copies owned, lent out ones included"`
	LentTo []string `json:"lentTo,omitempty" desc:"names of the people currently holding a copy; …"`
	// …
}
```

The schema reflection handles this: named string types (`type BookID string`) become `string`, unexported fields
are skipped, `time.Time` becomes a `date-time` string.

A DTO is right when the model needs something the domain type does not have: a value formatted for humans, an
aggregate computed over several records, or a deliberately reduced view. Not to rename fields.

A field is required by default; `optional:"true"`, a json `omitempty` or a pointer make it optional. A filter is
the opposite of a request, so every filter field carries `optional:"true"`:

```go
type BookFilter struct {
	Query string `json:"query" optional:"true" desc:"…"` // filter: narrows
}

type LendRequest struct {
	Book BookID `json:"book" desc:"…"` // request: commands
}
```

A filter whose fields are reported as required forces the model to invent values for all of them on every call,
and it will, because the schema says so.

### 7. Describe the result only where it is not obvious

No provider accepts an output schema for tools; a tool definition has a name, a description and an input schema.
The only place a description of the result can go is the description text, which is sent with every request for
every tool. So `WithResultDoc()` is an explicit opt-in:

```go
list := completion.NewSeqTool("list_books", "…", uc.FindAllBooks).
	WithResultDoc() // because of truncated: the model cannot tell that from the data

lend := completion.NewUseCaseTool("lend_book", "…", uc.LendBook).
	AsMutating("…") // no WithResultDoc: summary and available speak for themselves
```

Where field names speak for themselves, the model learns the shape from the first real result, for free. Note
the consequence: without `WithResultDoc()`, the `desc` tags of the result type never reach the model.

## Permissions: two roles, not one

The assistant needs three framework permissions: list the providers, list the models, create a session
(`uicompletion.RequiredPermissions`). Do not put them into the business role. `cfgai.Enable` declares a system
role for them:

```go
cfgai.RoleAssistantUser // "nago.ai.assistant.user", shown as "AI Assistant User"
```

Assign the role and the button appears. The business role is unaffected, and the assistant can still only do
what the business role allows. Ship your own business role the same way:

```go
cfg.DeclareSystemRole(role.Role{
	ID:          RoleLibrarian,
	Name:        "Bibliothekar",
	Description: "Darf den Bestand einsehen sowie Exemplare ausleihen und zurücknehmen.",
}, LibrarianPermissions()...)
```

A system role can neither be deleted nor have its permissions replaced in the admin UI, since both would be undone
on the next start. Name and description are written only when the role is created; the operator may change them
and a restart does not overwrite that. See [role management](/docs/systems/role_management/).

This deserves a test: the bootstrap account holds every permission by construction, so nobody notices when a
shipped role grants none of them, until the first real user sees an empty page. `app/library/cfg/roles_test.go`
therefore spells out the permission IDs as literals. A test which derives its expectation from the code under
test agrees with every change, including the wrong one.

## Context: where the user is

Routes describe themselves when they are registered:

```go
// app/library/cfg/cfg.go
cfg.RootViewWithDecoration(pages.Books, func(wnd core.Window) core.View {
	return uilibrary.PageBooks(wnd, uc)
}, application.Purpose(
	"Den Bestand der Bibliothek durchsehen: welche Titel es gibt, wie viele Exemplare frei sind und wer welches ausgeliehen hat."))
```

`uicompletion.WindowContext(wnd)` turns this into the situational part of the prompt: the route and its purpose,
the parameters the page was opened with, and who is asking with their permissions. Everything comes from what the
framework already knows, so it cannot drift like a hand-maintained list of screens. The domain part of the
prompt stays separate, see `SystemPromptFunc` below.

## Requirements as a tool: "why is it like this?"

Every domain tool answers a question about data. None answers the question people actually have when the system
refuses something:

> "Von *Der Prozess* ist derzeit kein Exemplar frei." Then why can't I reserve one?

Without a tool for that, the model does not refuse to answer; it invents a reason, fluent and plausible. An
invented rule is worse than silence, because it sounds like the system talking about itself.

### Nago ships this

Requirements are declared with `spec.Declare` and end up in a runtime catalogue:

```go
// requirements/dec/R-DEC-AVAILABILITY.spec.go
var RDecAvailability = spec.Declare(spec.Requirement{
	ID:           "R-DEC-AVAILABILITY",
	Kind:         spec.Decision,
	Discipline:   spec.Technical,
	Status:       spec.Normative,
	Title:        "Verfügbarkeit wird berechnet, nicht gespeichert",
	Text:         "Die Zahl der freien Exemplare ergibt sich aus Copies minus der Länge von LentTo …",
	Rationale:    "Ein abgeleiteter Wert, der zusätzlich gespeichert wird, …",
	Consequences: "Jede Anzeige rechnet neu. …",
})
```

The annotation file binds use case and requirement. It is part of the normal build, so it breaks when the use
case disappears:

```go
// app/library/uc_lend_book.annotation.go
var _ = spec.For[LendBook](
	spec.Satisfies(fun.RLibLend, dec.RDecBorrower),
	spec.Help("Gibt ein Exemplar an eine Person heraus. Ist keines frei, wird die Ausleihe abgelehnt — eine Vormerkung gibt es noch nicht."),
)
```

That is all the work. The tools come from the framework:

```go
specMod := option.Must(cfgspeclink.Enable(cfg))

tools := append(ailibrary.Tools(lib.UseCases), aispeclink.Tools(specMod.UseCases)...)

SystemPromptFunc: func() string {
	return ailibrary.SystemPrompt + "\n\n" +
		aispeclink.Index(wnd.Subject(), specMod.UseCases) + "\n" +
		uicompletion.WindowContext(wnd)
},
```

`aispeclink.Tools` returns `list_requirements`, `read_requirement`, `read_capabilities` and
`read_source_document`; `aispeclink.Index` renders one line per visible requirement for the prompt. There is no
hand-written knowledge layer in this example, and that is the point: a hand-made copy of the catalogue drifts
without anything breaking.

### State is not existence

`R-LIB-RESERVATION` is deliberately `planned` and bound to nothing. A view which only reads the bindings would not
see it at all, and the assistant would answer "that does not exist" instead of "that is not implemented yet". This
is why Nago reads the catalogue (`spec.Requirements()`), not the binding registry.

### Who may see which requirement

`spec.Requirement` has a `Disclosure` field (`public`, `internal`, `confidential`, `secret`). speclink states that
nothing enforces it; enforcing it is up to whoever shows the text to somebody, so Nago does:

| Disclosure                 | without `nago.speclink.requirement.read_internal` | with                         |
|----------------------------|---------------------------------------------------|------------------------------|
| `public`                   | visible                                           | visible                      |
| `internal`, `confidential` | hidden                                            | visible                      |
| `secret`                   | never                                             | only individually, never in a list |

A requirement the user may not see is reported as not existing. Answering "it exists, but you may not see it"
already reveals more than guessing an ID should.

The system role `nago.speclink.reader` bundles the reading permissions (list and read requirements, list
capabilities, read source documents). `read_internal` is deliberately not part of it, because that is a decision
about a person, not about a feature.

### An admin page on the side

`cfgspeclink.Enable` also registers the page `admin/speclink/requirements`. Same list, same use cases, same
disclosure rules, so what an operator sees there and what the assistant may say are the same set by
construction.

## The button

Provider resolution, model choice, settings, caching and diagnostics come with `cfgai`:

```go
cfg.SetDecorator(func(wnd core.Window, view core.View) core.View {
	return scaffold(wnd, modAI.Assistant.Decorate(wnd, view, cfgai.AssistantOptions{ /* … */ }))
})
```

It hangs on the decorator rather than on single pages, so the assistant is really everywhere, including the
admin pages of the framework. When it cannot run (no provider, no model, hidden by the operator, missing role),
`Decorate` returns the view unchanged and logs the reason once. A missing token must not become a banner that
follows the user through the application.

### Date and time are built in

A model has no clock. Its "today" is its training cut-off, and it calculates with that fluently and wrongly:
"overdue since yesterday", "next Monday", "in two weeks". So the assistant brings the tool `current_time` on its
own: date, time, weekday, ISO week and time zone of the window (`wnd.Location()`). This tutorial does nothing for
it.

Why a tool and not a line in the system prompt: a prompt that changes every second never hits the prompt cache of
the provider, while a tool costs only when the model needs it. If you provide a tool with the same name, the
built-in one steps back; `AssistantOptions.DisableCurrentTime` turns it off.

## The screen as feedback channel

Every tool above tells the model what is in the data. None tells it what the user sees, and that is often the
question: "what does the number on the right mean?", or the model just lent a book and assumes the list was
updated. `uicompletion.ScreenTool` is a ready-made tool for this, a kind of built-in Playwright:

```go
Tools: slices.Concat(tools, []completion.Tool{
	uicompletion.ScreenTool(wnd, uicompletion.ScreenToolOptions{}),
}),
```

The tool is called `inspect_screen`. It always returns an accessibility snapshot of what is currently rendered,
with roles, names, values and states as an indented tree:

```
- heading "Bestand" [level=1]
- textbox "Suche": "Kafka"
- button "Ausleihen"
- checkbox "Nur verfügbare" [checked]
```

If the model asks for it (`image: true`), a PNG is added. Text is the default because it is more precise and much
cheaper: an image stays in the history and is billed again on every following turn.

Three things are deliberate:

- **It is bound to the window.** Unlike the domain tools, it is built per window in the decorator, because it looks
  at exactly this window. `slices.Concat` copies, so the shared slice is never appended to from several windows at
  once. It is also never handed to sub-agents.
- **It needs no permission.** The model only sees what is rendered for the acting user anyway, and the developer
  has to wire the tool explicitly.
- **Pending changes are rendered first.** `wnd.Screenshot` sends the latest state before requesting the capture,
  and the frontend processes messages in order. A look right after `lend_book` shows the result, not the state
  before.

The PNG is re-rendered from the DOM by the frontend, not captured from the screen. Images from foreign domains
without CORS, iframes and videos may be missing. That is good enough for "does this look right"; if you need the
capture yourself, call `wnd.Screenshot(core.ScreenshotOptions{…})` directly.

## Try it

Sign in with the bootstrap admin; its password is in `cmd/ai-example/main.go`. Store a provider token in the
vault of the admin center (see [secret management](/docs/systems/secret_management/)) and assign the roles
"Bibliothekar" and "AI Assistant User" to the users. For the questions about requirements they also need the
role "Requirements Reader" (`nago.speclink.reader`).

Questions to try:

- "Was ist von Kafka da?" A read query.
- "Leih Die Verwandlung an Bernd aus." The confirmation dialog appears. Decline once and watch the model pick up
  the refusal instead of failing.
- Enable *Nur lesender Zugriff* in the assistant settings and ask again. The model no longer knows the tool.
- "Warum steht bei den Ausleihern nur ein Name und kein Benutzerkonto?" The model looks up `R-DEC-BORROWER` and
  also names what the decision costs, instead of making something up.
- "Kann ich ein ausgeliehenes Buch vormerken?" The answer is "not yet", not "does not exist":
  `R-LIB-RESERVATION` is `planned`.
- "Was kann ich hier eigentlich machen?" Orientation via `read_capabilities`.
- "Welcher Wochentag ist heute, und welche Kalenderwoche?" `current_time` instead of a guessed date.
- "Was steht bei mir gerade auf dem Bildschirm?" `inspect_screen` returns the snapshot; "Wie sieht das aus?"
  also fetches the image.

## Code

The entry point bootstraps the framework, enables the library context, the AI system and the requirement
catalogue, and hangs the assistant on the decorator:

{{< example-code name="tutorial-113-ai-assistant" file="cmd/ai-example/main.go" footer="false" >}}

The wiring of the library context: store, use cases, its role and its screen:

{{< example-code name="tutorial-113-ai-assistant" file="app/library/cfg/cfg.go" footer="false" >}}

The view for a model: the tools and the domain half of the system prompt:

{{< example-code name="tutorial-113-ai-assistant" file="app/library/ai/tools.go" footer="false" >}}

The domain: aggregate, permissions, repository, use cases and their annotations:

{{< example-code name="tutorial-113-ai-assistant" file="app/library/*.go" footer="false" >}}

The view for humans:

{{< example-code name="tutorial-113-ai-assistant" file="app/library/ui/*.go" footer="false" >}}

The requirements:

{{< example-code name="tutorial-113-ai-assistant" file="requirements/**.go" pkg="tutorial-113-ai-assistant/cmd/ai-example" >}}
