# gostrict – Stand und offene Aufgaben

gostrict ist ein streng eingeschränktes Go-Subset, das nach JavaScript übersetzt wird. Ein Programm ist gültiges Go
und läuft nativ, in goja und im Browser (Komponenten `threejs` und `canvas2`). `doc.go` beschreibt das Subset und ist
zugleich der Prompt für Sprachmodelle, die gostrict-Code schreiben.

Ziele in dieser Reihenfolge: keine Laufzeitfehler (statisch bewiesen), LLM-taugliche Diagnosen (`GSxxx` mit `why`,
`fix` und Edits), Typsicherheit in Go auf beiden Seiten der Grenze, native Performance vor bitgleichen Floats.

## Regeln für die Arbeit

- Größere Designentscheidungen und alles, was Protokoll oder Frontend von nago berührt, vorher abstimmen.
- Commit und Push nur nach Freigabe, ohne Attributionszeilen. Vor dem Push `git pull --rebase`, keine Merge-Commits,
  nie force-pushen.
- `web/vuejs/dist` nicht pro Änderung committen, sondern erst am Ende als `chore(frontend): build`, wenn es verlangt
  wird. Generierter Protokollcode (`protonc_gen.go`, `nprotoc_gen.ts`) gehört zum Commit seiner Spec.
- Nicht anfassen: die Stashes `flow_chart_pos`, `dnd`, `WIP_DECIDE_picker_translations`; die App-Console; wokoda
  (nur lesen).
- `gofmt` der Toolchain verwenden (`"$(go env GOROOT)/bin/gofmt"`), nicht das ältere aus Homebrew.

## Entscheidungen

| Thema | Entscheidung |
|---|---|
| Sprache | Go-Subset statt TypeScript, Prüfung per Whitelist auf `go/ast`/`go/types`. |
| Zahlen | Kein `int`; `int32`/`uint32` mit Wrap-around, sonst `float64`. Floats sind über Engines nur ungefähr gleich (FMA, GPU). Exakt: `std/big.Int` (Header, BigInt bzw. `math/big`, `==` verboten mit GS041) und `std/decimal.Dec{M big.Int; S int32}` (Subset-Code, Skala 0..`MaxScale` = 1000, Rundung explizit: HalfUp, HalfEven, Down, Up, Floor, Ceiling). |
| Totalität | Guards statt Laufzeitprüfungen, nur `error` darf nil sein, kein `panic`, Package-Variablen unveränderlich. |
| Typen | `list`/`dict`/`opt` statt Slices/Maps/nil; Summentypen als versiegelte Interfaces (Semantik nach golang/go#76920); erschöpfende Enums. |
| Engine | Blackbox `gostrict.Engine`/`Instance`: goja, nativ, Browser. |
| Generierung | Nur Header-Pakete (`lib/three`, `lib/canvas`, Beispiel `host`) haben ein `gostrict_gen.go` mit `Host()`, weil Reflection weder generische Header-Funktionen instanziieren noch Handles bauen kann. Programme und Anwendungen brauchen keinen generierten Code. |
| Go-Seite | Reflection: `gostrict.ToNeutral`/`FromNeutral`, `gostrict.Call` (typisierte Aufrufe), `native.Reflect` (native Engine), `gostrict.Router` + `Handle[T]` (Empfang), `Session.Send`/`Channel.Send` (Senden). Unions, die Go dekodiert, einmal mit `gostrict.Variants[T](...)` registrieren. Eingebettete Programme kompiliert `program.MustCompileFS(sources, "dir")`, das den Modulpfad aus den Imports (`gostrict.InferModulePath`) oder dem `main`-Paket ableitet; er muss dem Go-Importpfad entsprechen, weil Nachrichten nach Go-Typnamen zugeordnet werden. |
| Transport | Nago-Protokoll trägt nur Komponenten-ID, Typname und JSON (`GostrictMessage.Json`, `GostrictProgram.Args`, `GostrictCall.Args`, `GostrictReturn.Values`). Records sind Objekte mit Go-Feldnamen, Listen Arrays; Sonderfälle mit `$`-Schlüssel: `$m` Map, `$` Variante, `$f` NaN/±Inf/-0, `$b` big.Int als Dezimalstring, `$e` Fehler, `$r` Zustand der Browser-Engine. Im Browser Reviver/Replacer ohne Schema. Keine anwendungsspezifischen Typen im nago-Protokoll. |
| Nachrichten | Programme senden über `msg.Out[T]`-Rückgaben, Handler sind Exporte `func(s *State, m Record\|Union)`. Nachrichtenpakete heißen `messages`, nicht `proto`. Zustellung höchstens einmal, FIFO je Richtung; Programme fragen ihren Zustand im `Init` an. |
| Vertrauen | Header-Pakete nur aus vertrauenswürdigen Modulen (GS039); Programme von Nutzern oder LLMs nie. |
| Callbacks | Header-Funktionen nehmen Funktionen des Programms (`ValueFunc`, in Go ein `gostrict.Callback`, generierter Code macht daraus eine Go-Funktion). Aufruf während eines Aufrufs (Host-Funktion) oder später zwischen zwei Aufrufen (Fenster rendert), immer auf der Goroutine der Instanz (Event-Loop des Fensters); innerhalb eines Aufrufs ohne erneutes Sperren. Fehler der Engine: Nullwerte bzw. der Fehler als `error`-Ergebnis, die Instanz liefert ihn beim nächsten Aufruf einmal. Keine Bereichs-Callbacks: in goja ist eine Schleife in JS langsamer als Aufrufe aus Go. |
| Server-Header | `//gostrict:host` statt `//gostrict:js`: nur ein Go-Host implementiert die Funktion (`lib/ui`), Engines laden das Bundle sonst nicht, der Browser nie. Exporte dürfen Handles direkt nehmen und zurückgeben (`func View() ui.View`), nicht in Records, Listen oder Maps. |
| lib/ui | Spiegelt `presentation/ui` von nago 1:1 (Namen, Builder, Enums mit gleichen Werten, `int` → `int32`), jede Funktion `//gostrict:host`. Views sind Handles, die das versiegelte Interface `ui.View` vereint; Builder geben ihren eigenen Typ zurück, auch wo nago `DecoredView` liefert. Werte (Records, Enums, Konstanten, Variablen mit JS-Bindung) liegen im Header-Paket; `TestVars` vergleicht die Bindungen mit Go. Zustände nur über `StringStateOf(wnd, id)` u. a. (kein `AutoState`, der die Aufrufstelle in Go nutzt). Ein Server rendert mit `ui.Render(ctx, instance, "Render", wnd)`. |
| Ganzes Programm | `LoadRoots`/`CompileWhole`: alle Skripte als ein Programm, einmal geprüft; `Whole.Bundle(entries...)` übersetzt je Ziel nur Erreichbares. Einstiegspunkte sind alle exportierten Funktionen der Wurzeln (GS042: nicht generisch, nur Daten, keine Unions). Der Compiler ruft `$enter`/`$exit` (`gostrict.Aspect`, `Funcs.WithAspect`) um jeden Einstieg vom Host oder aus einem anderen Paket, auch über Funktionswerte; im eigenen Paket nicht. Ablehnung oder Fehler des Hosts: `gostrict.Abort` beendet den ganzen Aufruf (für das Programm nicht abfangbar, wie R-SCRIPT-UNCATCHABLE in wokoda), alle Frames erhalten `Exit` mit Fehler, der Aufruf liefert den Fehler selbst (`errors.Is`). Imports sind zyklenfrei, also gibt es kein A→B→A und keine Aufrufe nach Namen zur Laufzeit; die Tiefe ist statisch begrenzt. Nativ gibt es keinen Aspekt (Go ruft direkt). |
| Werte teilen | Struct-Typen, deren Werte das ganze Programm nie verändert (Feldzuweisung, auch über Ketten und Zeiger, `*p = v`), sind unveränderlich: keine Kopien, ein gemeinsamer Nullwert, Zuweisungen ersetzen statt kopieren (`emit_mut.go`). `Record.Immutable` zusätzlich ohne Listen/Maps; goja konvertiert solche Objekte einmal (Identitäts-Cache, je Aufruf geleert, max. 4096). Records sind Objektliterale statt Klassen (goja: Klasse 1,4 µs, Literal 0,6 µs je Objekt); Felder werden in Quellreihenfolge ausgewertet. |

## Aufbau

```
pkg/gostrict/            Compiler (load, check*, emit*, runtime.js), Grenze (boundary, engine, values, router),
                         Session, Cache, Pool, Generator (nur Header), fix, sourcemap, reload
pkg/gostrict/std/        list, dict, opt, text, num, msg, fuel, js, big, decimal; host.go mit //go:build !gostrict
                         für Reflection
pkg/gostrict/lib/        three, canvas (Header), input, asset (Records), ui (Server-Header, bisher nur Sheet),
                         ui/sheet (Records wie presentation/ui/sheet)
pkg/gostrict/engine/     gojaengine (inkl. CheckBindings), native (Reflect)
presentation/ui/program  Bundle-Endpunkt, Channel, JSON-Codec, Browser-Engine
presentation/ui/{threejs,canvas2}, web/vuejs/src/components/{gostrict,threejs,canvas2}
example/cmd/tutorial-117-threejs, -118-canvas2-platformer, -119-canvas2-photobook
```

Prüfen:

```bash
cd pkg/gostrict && "$(go env GOROOT)/bin/gofmt" -l . && go vet ./... && go test -race -count=1 ./...
go generate .   # nur _stubs und gostrict_gen.go der Header-Pakete
cd ../.. && go test -race ./presentation/ui/program/ ./example/cmd/tutorial-11{8,9}*/...
cd web/vuejs && npx vue-tsc --noEmit -p tsconfig.json --composite false   # Fehler in nprotoc_gen.ts sind Altlasten
```

## Messungen

`go test -bench . ./lib/ui/` (Apple M-Serie): Fensterwechsel einer Tabelle mit 90 × 29 Zellen und Kopfzeile, also rund
2700 Aufrufe der Zellfunktion, ganz durch nago (`nagotest`):

| Quelle | je Fensterwechsel | je 1000 Zellen |
|---|---|---|
| Go direkt | 0,38 ms | 0,15 ms |
| gostrict nativ | 0,47 ms | 0,18 ms |
| gostrict in goja | 4,8 ms | 1,8 ms |

Ein Callback in goja kostet rund 1,7 µs, davon etwa zwei Drittel für JS selbst (Objekt anlegen, Zahl zu Text) und ein
Drittel für die Umwandlung. Unter 1 ms je 1000 Zellen geht nur mit einer schnelleren Engine (I.12, eigener
Interpreter), eine direkte Umwandlung goja → Go-Struct spart höchstens ein Viertel und lohnt den Aufwand nicht.

## Offen

- **Use-Case-Vertrag für wokoda (G):** typisierte DB-Zeilen über generische Header-Funktionen mit Record-Typen. Den
  Kontext je Aufruf (Nutzer, Rechte, Transaktion) hält der Host über `Aspect` als Stapel von Frames, siehe
  `whole_test.go`; die wokoda-Header (db, files, search, events, user, integration) entstehen downstream.
- **Rest von `lib/ui`:** Hauptpaket und Unterpakete von `presentation/ui` sind gespiegelt (Vorlage von Hand, Rest von
  haiku in Batches, Zeit/Kalender/Flowchart von sonnet). Lücken geschlossen über `IndexState`, `ColorState`,
  `DateState`, `TimeFrameState`, `DurationState`, `ui.Time`/`ui.Duration` (Unix-Millisekunden), Records mit
  `list.List`-Feldern statt Slices, Handles mit Buildern für Records mit Funktionen (z. B. `CalendarEvent`), sowie
  Wrapper für Dataview (`DataViewFromSlice`, Zeilen per Index, ohne Repository), Charts (`ui.Chart`,
  `ChartSeries`) und Dropdown (`string`-IDs). Es fehlen: `crud`, `form`, `treeview`, `dropdown.FromSlice`, alles mit
  `auth.Subject`/Repositories/i18n-Diensten, `Box(BoxLayout)`, `SignatureField`, QR-Leser, die Icons von hero und
  material (Flowbite: `lib/ui/icons/flowbite/{outline,solid}`, erzeugt von `TestIcons -update`). Picker über Indizes
  (`PickerFromSlice` mit `IndicesState`), `ui.Background()`, `Image.Embed(SVG)`/`Adaptive(ImageID)` und
  `Colors(...)` der Charts sind ergänzt. Die Baumvergleiche der Tests sind tief (JSON), Zeiger und erzeugte IDs
  normalisiert.
- **`opt.Option` an der Grenze:** fehlt, deshalb hat `sheet.Cell` keine `Options` (Select-Editor) und keine `View`
  (Handles in Records).
- **`go/analysis`-Analyzer (I.6):** neue Abhängigkeit `golang.org/x/tools`, abstimmen.
- **Verträge und Wertebereiche (I.8, I.9):** Sprachdesign offen.
- **QuickJS über wazero (I.12):** Engine mit Speicherlimit, neue Abhängigkeiten.
- **Fotobuch:** Drag-and-drop von Dateien direkt auf den Canvas.
- **Unions im Browser Ende-zu-Ende:** Codec in Go (`TestCodec`) und Reviver/Replacer in `session.ts` sind einzeln
  geprüft, ein Tutorial mit Union-Nachrichten fehlt.

## Pflege

- goja wandelt BigInt über int64 in Number (`Number(2n**64n)` ist 0); die Bindung von `big.Float64` parst deshalb
  außerhalb von int64 die Ziffern. Neue BigInt-Bindungen mit Werten über 2^64 im Differenztest (`money`) prüfen.

- `doc.go` bei jeder Spracherweiterung aktualisieren.
- Neue Diagnosen mit Testfall in `testdata/bad/*` (`// want GSxxx`) und konkretem `fix`, wo möglich mit Edits.
- Neue Beispiele unter `internal/examples` bekommen einen Eintrag in `exports` (`diff_test.go`) und einen
  Differenztest nativ gegen goja.
- `TestNoToolchain` muss grün bleiben: kein `go/build`, `go/importer`, `os/exec` im Compiler.
