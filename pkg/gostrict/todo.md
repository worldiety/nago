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
| Fehlerpositionen | `errors.New` und Bibliotheksfunktionen mit `error` (`text.ParseInt`) bekommen vom Emitter den Index ihrer Position in `Bundle.Positions` (Datei wie in der Source-Map, Zeile), die Laufzeit hält ihn als `$p` im Fehler. goja liefert daraus `*gostrict.ProgramError{Msg, File, Line}` (auch an `Aspect.Exit`), `gostrict.ErrorPos(err)` liest sie. Vergleich nur über die Nachricht; nativ, im Browser, in Bindungen, beim Host und bei eigenen Fehlertypen keine Position. `Bundle.Hash` bleibt der Hash des JS, die Tabelle gehört zum Bundle der Instanz. |
| Schleifen | Ein `range` über eine Liste (`All`, `Values`, auch `msg.Out`) besucht die Elemente, die sie beim Start hat, wie ein `range` über einen Slice: jede solche Schleife endet, auch über Aliase und Aufrufergebnisse, deshalb prüft GS040 sie nicht mehr. |
| Rekursion | `checkRecursion` verbindet dynamische Aufrufe (Funktionswerte, Interface-Methoden) nur mit Paketen derselben Import-Zusammenhangskomponente; Werte wandern nur entlang von Imports. `list.SortFunc` gilt als dynamischer Aufruf seines Arguments. |
| Methodentabellen | Enthalten nur Methoden, deren Namen das Programm über ein Interface aufruft, und `Error`. Builder der Views von `lib/ui` landen so nicht mehr in jedem Bundle (minimale View: 727 → 1 Extern, 197 → 19 KB JS, goja-Heap je Instanz 755 → 122 KB). |
| Interpreter (`engine/vm`) | Register-Bytecode über einer typisierten Baum-IR (Closures waren im Spike 15–50 % langsamer). Drei Bänke: Skalare (`uint64`), Strings, Referenzen (`any`). Structs liegen flach in den Slots ihrer Felder; unveränderliche Structs mit mehr als 8 Slots sind geteilte Objekte. `Emit` übersetzt jedes Bundle zusätzlich mit `internal/vmcode` (`Bundle.VM`), das Bundle hält nur das fertige Programm, nicht Syntax und Typen. Jede Header-Funktion braucht in der VM eine Go-Implementierung; std ist nativ in Go (`std/big` auf `math/big`, geprüft durch `TestBig`), `std/decimal` wird wie Programmcode übersetzt. |
| Werte teilen | Struct-Typen, deren Werte das ganze Programm nie verändert (Feldzuweisung, auch über Ketten und Zeiger, `*p = v`), sind unveränderlich: keine Kopien, ein gemeinsamer Nullwert, Zuweisungen ersetzen statt kopieren (`emit_mut.go`). `Record.Immutable` zusätzlich ohne Listen/Maps; goja konvertiert solche Objekte einmal (Identitäts-Cache, je Aufruf geleert, max. 4096). Records sind Objektliterale statt Klassen (goja: Klasse 1,4 µs, Literal 0,6 µs je Objekt); Felder werden in Quellreihenfolge ausgewertet. |

## Aufbau

```
pkg/gostrict/            Compiler (load, check*, emit*, runtime.js), Grenze (boundary, engine, values, router),
                         Session, Cache, Pool, Generator (nur Header), fix, sourcemap, reload
pkg/gostrict/std/        list, dict, opt, text, num, msg, fuel, js, big, decimal; host.go mit //go:build !gostrict
                         für Reflection
pkg/gostrict/lib/        three, canvas (Header), input, asset (Records), ui (Server-Header, bisher nur Sheet),
                         ui/sheet (Records wie presentation/ui/sheet)
pkg/gostrict/engine/     gojaengine (inkl. CheckBindings), native (Reflect), vm (eigener Interpreter: Engine, Instanz,
                         Callbacks)
pkg/gostrict/internal/vmcode/  Compiler (IR, Lowering, Bytecode) und Interpreter, ohne Abhängigkeit von gostrict
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

### Spike des Interpreters `engine/vm` (Phase 0)

Typisierte Baum-IR aus geprüftem AST und `go/types`, daraus Register-Bytecode (`gen.go`, `interp.go`) und zum
Vergleich Go-Closures (`closures.go`, `Engine.Closures`). Werte liegen in drei Bänken: Skalare (`uint64`), Strings,
Referenzen (`any`). Structs liegen flach in den Slots ihrer Felder (Register, Felder von Objekten, Listen mit
Schrittweite), ohne Allokation; das ist exakt, weil das Subset keine Adressen von Variablen oder Feldern nimmt.
Unveränderliche Structs mit mehr als 8 Slots (`emit_mut.go`) sind ein geteiltes Objekt mit gemeinsamem Nullwert und
einmal umgewandeltem neutralem Wert, z. B. `sheet.Style`. std (text, strings, math, num, big, list) ist nativ in Go,
`std/decimal` wird wie Programmcode übersetzt. Übersetzt werden bisher arena, errs, ints, money, orders, physics und
cells; die Differenztests von ints, errs, money und physics sowie `TestSheet` liefern dieselben Werte wie nativ.

Apple M2 Max, `go test -bench` in `engine/vm` und `lib/ui`:

| Messung | goja | vm Bytecode | vm Closures | nativ |
|---|---|---|---|---|
| `BenchmarkCallback` je Zelle, neutrale Werte | 1,59 µs, 32,5 Alloc. | 0,25–0,29 µs, 7,7 Alloc. | 0,29 µs | – |
| dasselbe, typisierter Adapter (Prototyp `typed_test.go`) | – | 0,20 µs, 2,8 Alloc. | – | – |
| `BenchmarkSheet` je Fensterwechsel | 4,63 ms | 1,27 ms | – | 0,45 ms (Go 0,34 ms) |
| `physics.Simulate(10000, 45)` | 41 ms | 2,7 ms | 3,8 ms | 0,09 ms |
| `ints.Xorshift(…, 100000)` | 46 ms | 2,3 ms | 2,9 ms | 0,21 ms |
| `ints.Hash` (8000 Zeichen) | 7,5 ms | 0,18 ms | 0,27 ms | 0,055 ms |
| `money.Factorial(300)` | 128 µs | 49 µs | 56 µs | 27 µs |
| `money.Invoice` (5 Zeilen) | 178 µs | 11,6 µs | 17,8 µs | 7,5 µs |
| Heap je weiterer Instanz (cells, nach einem Aufruf) | 763 KiB | 0,3 KiB | | |
| Heap je weiterer Instanz (money) | 78 KiB | 1,3 KiB | | |
| Heap der ersten Instanz inkl. Übersetzung (cells) | 1490 KiB | 19 KiB | | |

Ergebnis: Bytecode ist in jedem Fall 15–50 % schneller als Closures, also bleibt es beim Register-Bytecode. Die
Zielwerte für `BenchmarkSheet` (≤ 1,5 ms), Speicher je Instanz und Rechenlast sind schon erreicht. ≤ 3 Allokationen
je Zelle gehen nur mit dem typisierten Pfad: über neutrale Werte kosten Ergebnis-Slice, Record-Slice samt Box und je
String eine Box mindestens 5 Allokationen, wie in goja. Ein Befehl kostet rund 2,8 ns (Xorshift), da ist Spielraum
für Phase 5 (Superinstruktionen, Aufrufe ohne `ensure`/`hi` je Aufruf, schmalere Befehle).

### Interpreter nach Phase 1

Gleiche Messungen nach dem Umbau (Programm im Bundle, Umwandlung über `vmcode.Host`): `BenchmarkCallback` 0,22 µs je
Zelle (goja 1,40 µs), `BenchmarkSheet` 1,20 ms (goja 4,23 ms, nativ 0,43 ms), physics 2,7 ms, Xorshift 2,0 ms,
`money.Invoice` 9,5 µs (goja 130 µs, nativ 6,1 µs). Eine Instanz belegt 0,4–1,3 KiB, die erste nicht mehr, weil das
Programm schon im Bundle ist; das kostet ein Bundle von money rund 48 KiB und weniger als 1 % der Übersetzungszeit.

### Interpreter nach Phase 3

Callbacks der VM sind `gostrict.FastCallback`: die generierten Adapter schreiben die Argumente und lesen die
Ergebnisse über `gostrict.Writer`/`Reader` direkt in den Slots, ein geteilter Record (Style) wird einmal in Go
umgewandelt und behalten. `BenchmarkCallback` misst jetzt die Zellfunktion, wie nago sie ruft (generierter Adapter
und Umwandlung in die Zelle von nago):

| je Zelle | nativ | goja | vm |
|---|---|---|---|
| Zeit | 0,09 µs | 1,53 µs | 0,26–0,27 µs |
| Allokationen | 1,76 | 32,6 | 1,76 |

`BenchmarkSheet/vm` 0,95 ms je Fensterwechsel (goja 4,67 ms, nativ 0,47 ms). Die Allokationen der VM sind nur noch die
Strings von `text.FromInt`, wie nativ. Vom Rest entfallen etwa 80 ns auf den Bytecode (Vergleiche und Multiplikation
mit Konstanten, Structs direkt in die Ergebnisse), 50 ns auf Reader und Writer und 50 ns auf Sperre und Rahmen.

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
- **Interpreter `engine/vm` (I.12):** Phase 1 und 2 sind fertig: alle Ausdrücke und Anweisungen des Subsets, auch
  Funktionsliterale (Zellen nur für eingefangene Variablen, die danach zugewiesen werden, je Iteration neu),
  Methodenwerte, Interfaces (Box mit Methodentabelle je Typ, Type Switch, Vergleich, eigene Fehlertypen), Zeiger
  auf Variablen und Felder für Methoden mit Zeiger-Empfänger (Zelle bzw. Objekt, dessen Slices in das umgebende
  zeigen), Generics (Monomorphisierung) und Range über Iteratoren (Rumpf als `yield`, `return` und Sprünge nach
  außen über Zellen der Funktion). `vmPending` ist leer; neu ist das Beispiel `funcs`. Mit der VM laufen außerdem
  `TestHost`, `TestWidgets`, `TestUnions`, `TestUnionMessages`, `TestSession`, `TestThree`, `TestCanvas`,
  `TestRecordsAndState`, `TestTimeout` und in lib/ui `TestSheet`, `TestPage`, `TestForm`, `TestDataView`,
  `TestAgenda`, `TestExtras`. Variablen von Header-Paketen liefert `Host()` jetzt auch als Funktionen ohne
  Parameter (Generator), die VM liest sie einmal je Programm. Als Nächstes:
  - Phase 3 ist fertig. Typisierte Adapter der Header-Funktionen selbst (VM → Host) sind auf Phase 5 verschoben
    (abgestimmt): dafür müsste das Engine neben `gostrict.Funcs` schnelle Funktionen bekommen, ohne dass ein Host,
    der eine Funktion in `Funcs` ersetzt (z. B. ein Test), stillschweigend umgangen wird; im Sheet dominieren die
    Callbacks, die Header-Funktionen laufen einmal je Render. Erst nach einem Profil entscheiden.
  - Phase 4: ganze Programme (Aspekte um jeden Einstieg, `Abort`), bis dahin lehnt die VM sie beim Übersetzen ab.
  - Package-Variablen werden einmal je Programm mit den Host-Funktionen der ersten Instanz initialisiert. Für
    Initialisierer, die den Host rufen (Handles), muss das geklärt werden.
  - Zeiger auf Felder sind eigene Objekte, also `&x.A == &x.A` falsch, wo Go true liefert; das Subset kann solche
    Zeiger nur über Methoden mit Zeiger-Empfänger bilden, ein Vergleich darin ist unwahrscheinlich.
  - Fehler der VM tragen noch keine Quellposition; `GOSTRICT_VM_PANIC=1` lässt Panics mit Stack durch,
    `VMDUMP=<dir> go test -run TestDump ./engine/vm` zeigt den Bytecode.
- **Fotobuch:** Drag-and-drop von Dateien direkt auf den Canvas.
- **Unions im Browser Ende-zu-Ende:** Codec in Go (`TestCodec`) und Reviver/Replacer in `session.ts` sind einzeln
  geprüft, ein Tutorial mit Union-Nachrichten fehlt.

- **Aufrufe zwischen Instanzen einer Runtime:** Eine Host-Funktion, die eine andere Instanz derselben goja-Runtime
  aufruft, blockiert (`Runtime.mu` ist nicht reentrant). Ohne Kontext in den Host-Funktionen ist der Aufrufer nicht
  von einer anderen Goroutine zu unterscheiden; Skripte, die einander aufrufen, gehören in ein ganzes Programm
  (`LoadRoots`). Für `Pool` liefert `NewPoolFunc` Host-Funktionen je Instanz mit `Slot.Context()`.
- **Gemeinsamer Code je Bundle:** jedes Fenster instanziiert das Bundle in einer eigenen goja-Runtime; ein geteiltes,
  unveränderliches Code-Segment bräuchte eine eigene Engine (I.12).

## Pflege

- goja wandelt BigInt über int64 in Number (`Number(2n**64n)` ist 0); die Bindung von `big.Float64` parst deshalb
  außerhalb von int64 die Ziffern. Neue BigInt-Bindungen mit Werten über 2^64 im Differenztest (`money`) prüfen.

- `doc.go` bei jeder Spracherweiterung aktualisieren.
- Neue Diagnosen mit Testfall in `testdata/bad/*` (`// want GSxxx`) und konkretem `fix`, wo möglich mit Edits.
- Neue Beispiele unter `internal/examples` bekommen einen Eintrag in `exports` (`diff_test.go`) und einen
  Differenztest nativ gegen goja.
- `TestNoToolchain` muss grün bleiben: kein `go/build`, `go/importer`, `os/exec` im Compiler.
