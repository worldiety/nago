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
  uc_*.annotation.go          bindet den Use Case an seine Anforderung
  ui/page_books.go            package uilibrary — die Ansicht für Menschen
  ai/tools.go                 package ailibrary — die Ansicht für ein Modell
  cfg/cfg.go                  package cfglibrary — die einzige Stelle, an der sich beides trifft
requirements/
  dec/R-DEC-*.spec.go         die Entscheidungen: warum die Bibliothek so funktioniert
  fun/library/R-LIB-*.spec.go was sie leisten muss
```

Die Abhängigkeiten zeigen ausschließlich nach innen: `ui` und `ai` kennen `library`, `library` kennt keinen von beiden. Ein Kontext, der seine eigene Oberfläche importiert, ist ohne Renderer nicht mehr testbar.

`ai/` steht bewusst neben `ui/`, nicht darin: Ein Modell ist eine Art, diesen Kontext zu erreichen, genau wie ein Bildschirm eine ist. Keines von beidem gehört zur Domäne.

### Ausführen

Weil der Einstiegspunkt unter `cmd/` liegt, ist der Aufruf einen Pfad länger als bei den anderen Tutorials:

```bash
go run go.wdy.de/nago/example/cmd/tutorial-113-ai-assistant/cmd/ai-example@latest
```

### Was hier fehlt

Das Beispiel nutzt `github.com/worldiety/speclink/spec` — ein Modul ohne jede Abhängigkeit, das nur die Deklarationen enthält. Das **Werkzeug** `speclink` selbst (und damit `speclink.json`, `speclink verify` und die statische Prüfung) ist nicht eingebunden: Es ist ein Compiler-Frontend und hat im Modulgraph einer Anwendung nichts zu suchen.

Die Anforderungen werden hier also deklariert und zur Laufzeit gelesen, aber nicht statisch geprüft. In einem echten Projekt kommt das dazu.

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

## Anforderungen als Werkzeug: „warum ist das so?"

Jedes Fachwerkzeug beantwortet eine Frage über **Daten**. Keines beantwortet die Frage, die jemand tatsächlich hat, wenn das System etwas ablehnt:

> „Von *Der Prozess* ist derzeit kein Exemplar frei."
> — *Warum kann ich mich dann nicht vormerken lassen?*

Ohne ein Werkzeug dafür verweigert das Modell nicht etwa die Antwort — es **erfindet eine Begründung**, flüssig und plausibel. Eine erfundene Regel ist schlimmer als Schweigen: Sie klingt wie das System, das über sich selbst spricht.

### nago liefert das fertig mit

Anforderungen werden mit `spec.Declare` deklariert und landen damit in einem Laufzeitkatalog:

```go
// requirements/dec/R-DEC-AVAILABILITY.spec.go
var RDecAvailability = spec.Declare(spec.Requirement{
    ID:           "R-DEC-AVAILABILITY",
    Kind:         spec.Decision,
    Status:       spec.Normative,
    Title:        "Verfügbarkeit wird berechnet, nicht gespeichert",
    Text:         "Die Zahl der freien Exemplare ergibt sich aus Copies minus der Länge von LentTo.",
    Rationale:    "Ein abgeleiteter Wert, der zusätzlich gespeichert wird, kann sich mit sich selbst widersprechen.",
    Consequences: "Jede Anzeige rechnet neu; eine Auswertung über zehntausend Titel geht nicht über einen Index.",
})
```

Die Annotationsdatei verbindet Use Case und Anforderung — und liegt im **normalen Build**, bricht also, wenn der Use Case verschwindet:

```go
// app/library/uc_lend_book.annotation.go
var _ = spec.For[LendBook](
    spec.Satisfies(fun.RLibLend, dec.RDecBorrower),
    spec.Help("Gibt ein Exemplar an eine Person heraus. Ist keines frei, wird die Ausleihe abgelehnt."),
)
```

Und das war die ganze Arbeit. Die Werkzeuge kommen aus dem Framework:

```go
specMod := option.Must(cfgspeclink.Enable(cfg))

tools := append(ailibrary.Tools(lib.UseCases), aispeclink.Tools(specMod.UseCases)...)

SystemPromptFunc: func() string {
    return ailibrary.SystemPrompt + "\n\n" +
        aispeclink.Index(wnd.Subject(), specMod.UseCases) + "\n" +
        uicompletion.WindowContext(wnd)
}
```

`aispeclink.Tools` liefert `list_requirements`, `read_requirement` und `read_capabilities`; `Index` rendert je eine Zeile für den Prompt. **Es gibt in diesem Beispiel keine handgeschriebene Wissensschicht** — und genau darum geht es: Wer den Katalog von Hand nachbaut, pflegt eine Kopie, die abdriftet, ohne dass etwas bricht.

### Zustand ist nicht Existenz

`R-LIB-RESERVATION` steht bewusst auf `planned` und ist an nichts gebunden. speclink verlangt eine Bindung nur für `normative` Anforderungen — eine Sicht, die nur die Bindungen liest, würde sie also **gar nicht sehen**, und der Assistent antwortete „so etwas gibt es nicht" statt „das ist noch nicht umgesetzt".

Genau deshalb liest nago den **Katalog** und nicht die Bindungsregistrierung. Die speclink-Dokumentation nennt diese Falle ausdrücklich.

### Wer welche Anforderung sehen darf

`spec.Requirement` hat ein Feld `Disclosure` (`public`, `internal`, `confidential`, `secret`). speclink sagt dazu klar: *„Nothing enforces it... it is not a control."* Die Durchsetzung ist Sache dessen, der den Text jemandem zeigt — also nagos.

| Stufe | ohne `nago.speclink.requirement.read_internal` | mit |
|---|---|---|
| `public` | sichtbar | sichtbar |
| `internal`, `confidential` | ausgefiltert | sichtbar |
| `secret` | nie | **nur einzeln**, nie in einer Liste |

Letzteres setzt die Definition wörtlich um: *„disclosed individually and never in bulk."* Eine Anforderung, die der Nutzer nicht sehen darf, wird als **nicht vorhanden** gemeldet — die Auskunft „gibt es, darfst du aber nicht" verrät bereits mehr, als das Erraten einer Kennung einbringen sollte.

Die Rolle `nago.speclink.reader` bündelt die drei Lese-Berechtigungen; `read_internal` ist bewusst **nicht** darin, weil das eine Entscheidung über eine Person ist und nicht über eine Funktion.

### Nebenbei: eine Verwaltungsseite

`cfgspeclink.Enable` registriert außerdem `admin/speclink/requirements`. Dieselbe Liste, dieselben Use Cases, dieselben Disclosure-Regeln — was ein Betreiber dort sieht und was der Assistent sagen darf, ist damit konstruktionsbedingt dieselbe Menge und nicht bloß per Absprache.

## Der Knopf

Provider-Auflösung, Modellwahl, Einstellungen, Cache und Diagnose liefert `cfgai` mit:

```go
cfg.SetDecorator(func(wnd core.Window, view core.View) core.View {
    return modAI.Assistant.Decorate(wnd, scaffold(wnd, view), cfgai.AssistantOptions{…})
})
```

Am Decorator statt an einzelnen Seiten, damit der Assistent wirklich überall ist — auch auf den Verwaltungsseiten des Frameworks. Kann er nicht laufen (kein Provider, kein Modell, ausgeblendet, Rolle fehlt), kommt die Ansicht unverändert zurück und der Grund landet **einmal** im Log. Ein fehlender Token darf kein Banner werden, das den Nutzer durch die Anwendung verfolgt.

## Der Bildschirm als Rückkanal

Jedes Werkzeug oben sagt dem Modell, was in den **Daten** steht. Keines sagt ihm, was der Nutzer **sieht** — und genau danach fragt er oft: „Was bedeutet die Zahl da rechts?", oder das Modell hat gerade ausgeliehen und nimmt an, dass die Liste sich aktualisiert hat.

`uicompletion.ScreenTool` ist dafür ein fertiges Werkzeug, eine Art eingebautes Playwright:

```go
Tools: slices.Concat(tools, []completion.Tool{
    uicompletion.ScreenTool(wnd, uicompletion.ScreenToolOptions{}),
}),
```

Es liefert immer einen **Accessibility-Snapshot** des gerade Gerenderten — Rollen, Namen, Werte und Zustände als eingerückter Baum:

```
- heading "Bestand" [level=1]
- textbox "Suche": "Kafka"
- button "Ausleihen"
- checkbox "Nur verfügbare" [checked]
```

Auf Wunsch des Modells (`image: true`) kommt ein PNG dazu. Der Text ist die Voreinstellung, weil er genauer und um ein Vielfaches billiger ist: Ein Bild bleibt Teil des Verlaufs und wird bei jeder folgenden Runde erneut abgerechnet.

Drei Dinge daran sind bewusst so:

- **Es hängt am Fenster.** Anders als die Fachwerkzeuge wird es pro Fenster im Decorator gebaut, weil es genau dieses Fenster ansieht. `slices.Concat` kopiert, damit die geteilte Liste nie von mehreren Fenstern gleichzeitig verlängert wird.
- **Es gibt keine Berechtigung.** Das Modell sieht nur, was dem handelnden Nutzer ohnehin angezeigt wird, und der Entwickler muss das Werkzeug ausdrücklich verdrahten.
- **Ausstehende Änderungen werden vorher gerendert.** `wnd.Screenshot` schickt erst den letzten Stand und dann die Aufnahme; das Frontend verarbeitet in Reihenfolge. Ein Blick direkt nach `lend_book` zeigt also das Ergebnis und nicht den Zustand davor.

Das PNG wird vom Frontend aus dem DOM nachgerendert, nicht vom Bildschirm abfotografiert. Bilder fremder Domains ohne CORS, iframes und Videos können fehlen. Für die Frage „sieht das richtig aus" reicht das; wer die Aufnahme selbst braucht, ruft `wnd.Screenshot(core.ScreenshotOptions{…})` direkt auf.

## Ausprobieren

Bootstrap-Admin: das Passwort steht in `cmd/ai-example/main.go`. Danach unter *Verwaltung → Tresor* einen Provider-Token hinterlegen und den Nutzern die Rollen „Bibliothekar" und „AI Assistant User" zuweisen.

Fragen zum Testen:

- „Was ist von Kafka da?" — eine Leseabfrage
- „Leih Die Verwandlung an Bernd aus." — der Bestätigungsdialog erscheint; einmal ablehnen und beobachten, dass das Modell die Absage aufgreift statt abzustürzen
- In den Einstellungen *Nur lesender Zugriff* setzen und erneut ausleihen lassen — das Modell kennt das Werkzeug dann nicht mehr
- „Warum steht bei den Ausleihern nur ein Name und kein Benutzerkonto?" — das Modell schlägt `R-DEC-BORROWER` nach und nennt auch, was die Entscheidung kostet, statt sich etwas auszudenken
- „Kann ich ein ausgeliehenes Buch vormerken?" — die Antwort ist „noch nicht", nicht „gibt es nicht": `R-LIB-RESERVATION` steht auf `planned`
- „Was kann ich hier eigentlich machen?" — Orientierung über `read_capabilities`
- „Was steht bei mir gerade auf dem Bildschirm?" — `inspect_screen` liefert den Snapshot; „Wie sieht das aus?" holt zusätzlich das Bild

## Example

Die Domäne — Aggregat, Berechtigungen, ein Use Case und das Bündel:

{{< include-code "app/library/model.go" >}}
{{< include-code "app/library/perm.go" >}}
{{< include-code "app/library/uc_lend_book.go" >}}
{{< include-code "app/library/usecases.go" >}}

Eine Anforderung und die Annotation, die sie mit dem Use Case verbindet:

{{< include-code "requirements/dec/R-DEC-AVAILABILITY.spec.go" >}}
{{< include-code "app/library/uc_lend_book.annotation.go" >}}

Die Ansicht für ein Modell:

{{< include-code "app/library/ai/tools.go" >}}

Die Verdrahtung und der Einstiegspunkt:

{{< include-code "app/library/cfg/cfg.go" >}}
{{< include-code "cmd/ai-example/main.go" >}}
