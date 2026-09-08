---
title: Tutorial 113
---

Ein KI-Assistent, der die Use Cases einer Anwendung bedient — mit den Berechtigungen des angemeldeten Nutzers und nicht mehr.

Tutorial 77 zeigt die *Mechanik* der Tool-Schleife (Rechner, Dateien). Dieses Tutorial zeigt die *Konventionen*: wie eine echte Fachanwendung ihre Use Cases anbindet, wie die Autorisierung ausgeliefert wird und was verhindert, dass ein Modell etwas ändert, das niemand verlangt hat.

## Aufbau

Das Beispiel folgt dem Profil `go_nago_ddd1` aus [speclink](https://github.com/worldiety/speclink) — dem Traceability-Werkzeug, mit dem worldiety nago-Projekte prüft. Es weicht damit bewusst von den übrigen Tutorials ab, die alles in eine `main.go` legen: Genau diese Struktur ist der Grund, warum der Assistent weiter unten ohne einen einzigen Adapter auskommt.

```
cmd/ai-example/main.go        Einstiegspunkt: Bootstrap und Verdrahtung, keine Fachlichkeit
app/library/                  der Bounded Context — was das System tut
  model.go                    Aggregat, Filter, Request, Result
  perm.go                     eine Berechtigung je Use Case
  repository.go               was der Kontext zum Speichern braucht, nicht wie
  usecases.go                 das UseCases-Bündel
  uc_find_all_books.go        ein Use Case je Datei, Typ und Konstruktor zusammen
  uc_lend_book.go
  uc_return_book.go
  ui/page_books.go            package uilibrary — die Ansicht für Menschen
  ai/tools.go                 package ailibrary — die Ansicht für ein Modell
  cfg/cfg.go                  package cfglibrary — die einzige Stelle, an der sich beides trifft
```

Die Abhängigkeiten zeigen ausschließlich nach innen: `ui` und `ai` kennen `library`, `library` kennt keinen von beiden. Ein Kontext, der seine eigene Oberfläche importiert, ist ohne Renderer nicht mehr testbar.

`ai/` steht bewusst neben `ui/`, nicht darin: Ein Modell ist eine Art, diesen Kontext zu erreichen, genau wie ein Bildschirm eine ist. Keines von beidem gehört zur Domäne.

### Ausführen

Weil der Einstiegspunkt unter `cmd/` liegt, ist der Aufruf einen Pfad länger als bei den anderen Tutorials:

```bash
go run go.wdy.de/nago/example/cmd/tutorial-113-ai-assistant/cmd/ai-example@latest
```

### Was hier fehlt

Ein Projekt, das speclink tatsächlich einsetzt, legt daneben noch zwei Dinge:

```
app/library/uc_lend_book.annotation.go        bindet den Use Case an eine Anforderung
requirements/fun/library/R-LIB-LEND.spec.go   die Anforderung selbst
```

Dieses Tutorial zieht die Abhängigkeit nicht, weil sie nichts über KI-Tools lehrt. Die Struktur ist aber dieselbe, und wer sie hier abschaut, schaut nichts ab, was später umgebaut werden müsste.

## Warum die Tools ohne Adapter funktionieren

Das ist der Bogen, um den es geht. Ein Use Case unter `go_nago_ddd1` hat die Form

```go
func(subject auth.Subject, cmd In) (Out, error)
```

Das ist **keine Bequemlichkeit für die KI**, sondern die Architekturregel des Projekts — das Subject ist ein Parameter, damit der Aufrufer entscheiden *muss*, wer da handelt, statt einen Context durchzureichen, der nichts entscheidet.

Und es ist zugleich exakt die Form, die ein Tool braucht. `completion.NewUseCaseTool` nimmt sie unverändert entgegen:

```go
lend := completion.NewUseCaseTool("lend_book", "…", uc.LendBook)
```

Kein Adapter, kein Subject am Konstruktor, kein Neuaufbau pro Anfrage. Wer seine Anwendung anders schneidet, schreibt für jedes Tool einen Wrapper — und muss die Autorisierung ein zweites Mal von Hand hinschreiben.

## Die sieben Regeln

Diese Regeln sind der eigentliche Inhalt. Alles andere folgt aus ihnen.

### 1. Ein Tool geht durch einen Use Case, nie durch ein Repository

Das ist kein Ordnungsprinzip, sondern das gesamte Autorisierungsmodell. Ein nago-Use-Case prüft das übergebene Subject — dasselbe, das auch ein Klick in der UI prüfen würde. `completion.NewUseCaseTool` reicht das Subject des fragenden Nutzers durch, also kann der Assistent nichts lesen, was dieser Nutzer nicht lesen darf, und nichts schreiben, was er nicht schreiben darf.

Ein Tool, das am Use Case vorbei ins Repository greift, hebelt das aus — und zwar unbemerkt, weil es funktioniert.

### 2. Use Cases brauchen keinen Wrapper

Ein nago-Use-Case hat die Form `func(auth.Subject, Request) (Response, error)`. Genau diese Form nimmt `NewUseCaseTool` entgegen:

```go
lend := completion.NewUseCaseTool("lend_book",
    "Leiht ein Exemplar eines Buches an eine Person aus.",
    uc.LendBook)
```

Kein Adapter, kein Subject-Parameter am Konstruktor, kein Neuaufbau pro Anfrage. Die Tools werden einmal beim Start gebaut und von jedem Fenster und jedem Nutzer geteilt.

Für die zweite übliche Form — `func(auth.Subject, Filter) iter.Seq2[T, error]` — gibt es `NewSeqTool` (siehe Regel 4). Wer wirklich eine eigene Funktion braucht, nimmt `NewSubjectTool`.

### 3. Schreibende Tools werden markiert, nicht beschrieben

```go
lend := completion.NewUseCaseTool(...).
    AsMutating("gibt ein Exemplar heraus und trägt die Person als Ausleiher ein")
```

Ein Satz im System-Prompt („frag immer nach, bevor du schreibst") ist eine **Bitte an das Modell**. Die Markierung ist ein **Gatter davor**: `ConfirmMutations` hält den Aufruf an und zeigt Werkzeug, Wirkung und die vom Modell gewählten Argumente, bevor irgendetwas passiert. `ReadOnly` — vom Betreiber in den Einstellungen umlegbar — entfernt schreibende Tools ganz, das Modell erfährt nicht einmal, dass es sie gibt.

Der Text in `AsMutating` ist das, was der Mensch vor dem Bestätigen liest. Ein leerer Text macht den Dialog wertlos.

### 4. Listen werden begrenzt

```go
list := completion.NewSeqTool("list_books", "…", uc.FindAllBooks)
```

`NewSeqTool` ergänzt das Schema um `limit` (Standard 200, hart 1000) und liefert `truncated` samt Klartext-Hinweis. Ohne das reicht ein `FindAll` über ein paar tausend Datensätze, um das Kontextfenster zu sprengen — und der Fehler sieht dann aus wie ein Provider-Problem, nicht wie eine fehlende Grenze.

Wichtiger noch: eine stillschweigend gekürzte Liste ist schlimmer als eine offen gekürzte. Das Modell soll den Filter enger ziehen, nicht aus einer halben Antwort schließen.

### 5. Namen und Beschreibungen sind Teil der Schnittstelle

Namen sind `lower_snake_case`, Beschreibungen sind für jemanden geschrieben, der die Anwendung nie gesehen hat. Beides wird beim Bauen geprüft: ein falscher Name oder eine leere Beschreibung ist ein Panic beim Start, kein ratloses Modell zur Laufzeit.

Dasselbe gilt für die `desc`-Tags am Request-Typ — sie sind die Dokumentation, die das Modell liest:

```go
type BookFilter struct {
    Query         string `json:"query" desc:"optional; case-insensitive substring matched against title and author"`
    OnlyAvailable bool   `json:"onlyAvailable" desc:"when true, only books with at least one free copy are returned"`
}
```

### 6. Der Domain-Typ ist der Tool-Typ

Ein eigenes DTO neben dem Aggregat ist meistens überflüssig. Die `desc`-Tags gehören an den Domain-Typ — genauso wie `label`-Tags dort stehen, wo `form.Auto` sie liest:

```go
type Book struct {
    ID     BookID   `json:"id" desc:"stable identifier, used when lending or returning"`
    Copies int      `json:"copies" desc:"total number of copies owned, lent out ones included"`
    LentTo []string `json:"lentTo,omitempty" desc:"names of the people currently holding a copy; …"`
}
```

Das funktioniert technisch bereits vollständig: benannte String-Typen (`type BookID string`) werden als `string` abgebildet, unexportierte Felder übersprungen, `time.Time` als `date-time`.

**Ein DTO ist dann richtig**, wenn das Modell etwas braucht, das der Domain-Typ nicht hat: einen für Menschen formatierten Wert (`time.Duration` → „8 Stunden"), ein über mehrere Datensätze berechnetes Aggregat, oder eine bewusst reduzierte Sicht. Nicht, um Felder umzubenennen.

**Pflicht vs. optional:** Ein Feld ist standardmäßig Pflicht. Ein Filter ist das Gegenteil davon — dort gehört `optional:"true"` an jedes Feld:

```go
type BookFilter struct {
    Query string `json:"query" optional:"true" desc:"…"`   // Filter: engt ein
}
type LendRequest struct {
    Book BookID `json:"book" desc:"…"`                     // Request: befiehlt
}
```

Ein Filter, dessen Felder als Pflicht gemeldet werden, zwingt das Modell, bei jedem Aufruf Werte für alle zu erfinden — und es tut es, weil das Schema es so verlangt.

### 7. Rückgabeform nur beschreiben, wo sie nicht offensichtlich ist

**Kein Provider akzeptiert ein Output-Schema für Tools.** Anthropic und die OpenAI-kompatiblen APIs kennen nur `name`, `description` und ein Input-Schema. Die einzige Stelle, an der eine Beschreibung der Rückgabe ankommen kann, ist der Beschreibungstext — und der wird bei *jeder* Anfrage für *jedes* Tool mitgeschickt.

Deshalb ist `.WithResultDoc()` ein bewusster Opt-in:

```go
list := completion.NewSeqTool("list_books", "…", uc.FindAllBooks).
    WithResultDoc()   // wegen truncated: das kann das Modell den Daten nicht ansehen

lend := completion.NewUseCaseTool("lend_book", "…", uc.LendBook).
    AsMutating("…")   // kein WithResultDoc: summary + available sprechen für sich
```

Wo Feldnamen für sich sprechen, lernt das Modell die Form am ersten echten Ergebnis — kostenlos.

**Konsequenz, die man kennen muss:** Ohne `WithResultDoc()` erreichen `desc`-Tags am Rückgabetyp das Modell **nie**. Sie sind dann toter Code.

## Berechtigungen: zwei Rollen, nicht eine

Der Assistent braucht drei Framework-Berechtigungen (Provider auflisten, Modelle auflisten, Sitzung anlegen). Diese **nicht** in die Fachrolle schreiben. `cfgai.Enable` deklariert dafür eine eigene Systemrolle:

```go
cfgai.RoleAssistantUser // "nago.ai.assistant.user"
```

Rolle zuweisen → Knopf erscheint. Die Fachrolle bleibt davon unberührt, und der Assistent kann trotzdem nur, was die Fachrolle erlaubt.

Die eigene Fachrolle liefert man genauso aus:

```go
cfg.DeclareSystemRole(role.Role{ID: RoleLibrarian, Name: "Bibliothekar", …}, LibrarianPermissions()...)
```

Eine Systemrolle lässt sich weder löschen noch in ihren Berechtigungen bearbeiten — beides würde beim nächsten Start ohnehin rückgängig gemacht. Name und Beschreibung darf der Betreiber ändern, und ein Neustart überschreibt das nicht.

**Warum das einen Test verdient:** Das Bootstrap-Konto hat konstruktionsbedingt jede Berechtigung. Es fällt also niemandem auf, wenn eine ausgelieferte Rolle keine davon gewährt — bis der erste echte Nutzer eine leere Seite sieht. `app/library/cfg/roles_test.go` schreibt die Berechtigungs-IDs deshalb als Literale aus. Ein Test, der seine Erwartung aus dem geprüften Code ableitet, stimmt jeder Änderung zu, auch der falschen.

## Kontext: wo der Nutzer steht

Routen beschreiben sich bei der Registrierung selbst:

```go
// app/library/cfg/cfg.go
cfg.RootViewWithDecoration(pages.Books, func(wnd core.Window) core.View {
    return uilibrary.PageBooks(wnd, uc)
}, application.Purpose(
    "Den Bestand der Bibliothek durchsehen: welche Titel es gibt, wie viele Exemplare frei sind …"))
```

`uicompletion.WindowContext(wnd)` macht daraus den situativen Teil des Prompts — Route und Zweck, die Parameter, mit denen die Seite geöffnet wurde, und die Berechtigungen des Nutzers. Alles aus dem, was das Framework ohnehin weiß, also driftet es nicht weg wie eine handgepflegte Liste von Bildschirmen.

Der fachliche Teil des Prompts bleibt davon getrennt:

```go
SystemPromptFunc: func() string {
    return ailibrary.SystemPrompt + "\n\n" + uicompletion.WindowContext(wnd)
}
```

## Der Knopf

Provider-Auflösung, Modellwahl, Einstellungen, Cache und Diagnose liefert `cfgai` mit:

```go
cfg.SetDecorator(func(wnd core.Window, view core.View) core.View {
    return modAI.Assistant.Decorate(wnd, scaffold(wnd, view), cfgai.AssistantOptions{…})
})
```

Am Decorator statt an einzelnen Seiten, damit der Assistent wirklich überall ist — auch auf den Verwaltungsseiten des Frameworks. Kann er nicht laufen (kein Provider, kein Modell, ausgeblendet, Rolle fehlt), kommt die Ansicht unverändert zurück und der Grund landet **einmal** im Log. Ein fehlender Token darf kein Banner werden, das den Nutzer durch die Anwendung verfolgt.

## Ausprobieren

Bootstrap-Admin: das Passwort steht in `cmd/ai-example/main.go`. Danach unter *Verwaltung → Tresor* einen Provider-Token hinterlegen und den Nutzern die Rollen „Bibliothekar" und „AI Assistant User" zuweisen.

Fragen zum Testen:

- „Was ist von Kafka da?" — eine Leseabfrage
- „Leih Die Verwandlung an Bernd aus." — der Bestätigungsdialog erscheint; einmal ablehnen und beobachten, dass das Modell die Absage aufgreift statt abzustürzen
- In den Einstellungen *Nur lesender Zugriff* setzen und erneut ausleihen lassen — das Modell kennt das Werkzeug dann nicht mehr

## Example

Die Domäne — Aggregat, Berechtigungen, ein Use Case und das Bündel:

{{< include-code "app/library/model.go" >}}
{{< include-code "app/library/perm.go" >}}
{{< include-code "app/library/uc_lend_book.go" >}}
{{< include-code "app/library/usecases.go" >}}

Die Ansicht für ein Modell:

{{< include-code "app/library/ai/tools.go" >}}

Die Verdrahtung und der Einstiegspunkt:

{{< include-code "app/library/cfg/cfg.go" >}}
{{< include-code "cmd/ai-example/main.go" >}}
