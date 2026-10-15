# gostrict – Übergabe und offene Aufgaben

Dieses Dokument richtet sich an das Modell, das die Arbeit an `pkg/gostrict` fortsetzt. Es fasst den Stand, die
verbindlichen Regeln, die getroffenen Entscheidungen und alle offenen Aufgaben zusammen. Lies es vollständig, bevor
du etwas änderst. Die Paketdokumentation in `doc.go` beschreibt das Subset selbst und ist zugleich der Prompt für
Sprachmodelle, die gostrict-Code schreiben.

---

## 1. Worum es geht

gostrict ist ein streng eingeschränktes Go-Subset, das nach JavaScript übersetzt wird. Ein gostrict-Programm ist
gültiges Go und läuft deshalb auf drei Wegen mit denselben Ergebnissen:

- **nativ** (normales Go, Referenz in Tests, Server),
- **in goja** (JS-Engine in Go, z. B. für Programme, die ein LLM zur Laufzeit schreibt),
- **im Browser** (z. B. three.js- oder Canvas-Programme in nago-Komponenten).

Ziele, in dieser Priorität:

1. **Keine Laufzeitfehler.** Der Checker beweist statisch: kein nil-Zugriff, kein Index außerhalb, keine Division
   durch null, kein `panic`. Die JS-Laufzeit (`runtime.js`) enthält deshalb keine Prüfungen.
2. **LLM-tauglich.** Jede abgelehnte Konstruktion bekommt eine Diagnose mit stabilem Code (`GS000`–`GS038`),
   Quellzeile, Grund (`why`) und konkretem Umbau (`fix`), als Text und als JSON (`gostrict check -json`).
3. **Typsicher über alle Grenzen.** Go-Typen gelten von der Backend-Logik über die Engine bis ins Browser-Programm;
   keine JSON-Roundtrips, keine Reflection in Hot Paths.
4. **Native Performance** hat Vorrang vor bitgleichen Floats.

---

## 2. Verbindliche Regeln für die Arbeit in diesem Repository

Diese Regeln hat der Nutzer festgelegt. Sie gelten ohne Ausnahme.

- **Vorher abstimmen.** Größere Designentscheidungen und alles, was Protokoll oder Frontend von nago berührt, vor
  der Umsetzung mit dem Nutzer abstimmen.
- **Commit und Push nur nach ausdrücklicher Freigabe.** Keine Attributionszeilen in Commit-Nachrichten.
- **Keine Merge-Commits auf main.** Vor dem Push `git pull --rebase`, niemals force-pushen.
- **Frontend-Build nicht pro Änderung committen.** `web/vuejs/dist` wird nicht nach jeder JS-Änderung committet,
  sondern erst ganz am Ende als eigener Commit `chore(frontend): build`, wenn der Nutzer es sagt. Generierter
  Protokollcode (`protonc_gen.go`, `nprotoc_gen.ts`) wird dagegen zusammen mit seiner Spec committet.
- **Nicht anfassen:** die Stashes `flow_chart_pos`, `dnd`, `WIP_DECIDE_picker_translations`; die App-Console (Hub);
  Dateien, die vorher nicht gofmt-sauber waren und nicht zur Aufgabe gehören (u. a.
  `application/ai/provider/anthropic/settings.go`, `application/user/password_test.go`,
  `uc_change_my_password.go`, `uc_requires_password_change.go`).
- **wokoda nicht verändern.** `/Users/tschinke/git/gitlab.worldiety.net/worldiety-erp/wokoda.git` wird von einem
  anderen Agenten umgebaut. Nur lesen.
- **gofmt der Toolchain verwenden:** `"$(go env GOROOT)/bin/gofmt"`. Das `gofmt` aus Homebrew
  (`/opt/homebrew/bin/gofmt`) ist älter als Go 1.27 und hält generische Methoden für einen Fehler.
- **Kommentar- und Codestil** wie im umgebenden Code: Englisch, Lizenzkopf in jeder Datei, Doku-Kommentare in
  ganzen Sätzen.

### Commit-Stand

Commit `0afc3c50` enthält den ursprünglichen Stand von `pkg/gostrict`. Alles aus Abschnitt 5 ist **nicht committet**
(Arbeitskopie im Worktree `twocolumn-right-alignment-e6ecb8`). `web/vuejs/dist` ist lokal gebaut und darf erst am
Ende als `chore(frontend): build` committet werden. Vorschlag für den Zuschnitt: (1) `feat(gostrict)` Compiler und
Bibliothek (`pkg/gostrict`), (2) `feat(proto)` Spec + generierter Code + `core.InputEvent`/`AsyncListen`,
(3) `feat(program)` `presentation/ui/{program,threejs,canvas2}` + Endpunkt + Vue, (4) `docs(example)` Tutorials
117–119, (5) `chore(frontend): build`.

## 3. Getroffene Entscheidungen (nicht neu verhandeln)

| Thema | Entscheidung |
|---|---|
| Sprache | Go-Subset statt TypeScript. Grund: eine Sprache mit dem Backend, natives Ausführen, Strenge selbst bestimmt auf `go/types`. |
| Prüfung | Whitelist auf `go/ast`/`go/types`. Neue Go-Features werden abgelehnt, nie falsch übersetzt. |
| Zahlen | Kein `int`. Ganzzahlen sind `int32`/`uint32` mit Wrap-around (in Go und JS identisch), sonst `float64`. |
| Floats | Nur ungefähr gleich über Engines (`CloseValues`, Toleranz 1e-9). Go fusioniert `a*b+c` auf arm64 zu FMA; das ist akzeptiert, native Performance zählt. |
| Laufzeitfehler | Totalität statt Prüfung: Guards (`if b != 0`, `if ok`, `if err != nil`), totale stdlib, `panic` verboten. |
| nil | Nur `error` darf nil sein. Referenztypen brauchen bei der Deklaration einen Wert. |
| Slices/Maps | Ersetzt durch `list.List[T]`, `list.Frozen[T]`, `dict.Map[K,V]` (deterministische Reihenfolge). |
| Globals | Package-Variablen sind nach der Initialisierung unveränderlich und halten nur unveränderliche Werte. Veränderlicher Zustand ist ein Struct, das eine Export-Funktion als Pointer zurückgibt (Zustands-Token). |
| Summentypen | Versiegelte Interfaces (unexportierte Methode) mit erschöpfendem Type Switch, Semantik nach golang/go#76920; wird #76920 angenommen, kommt dessen Syntax als zweite Schreibweise auf denselben Kern. |
| Enums | Benannter `int32`/`uint32`/`string`-Typ mit Konstanten; `switch` muss erschöpfend sein. |
| Engine | Blackbox hinter `gostrict.Engine`/`Instance`. goja ist Test- und Server-Engine; Alternativen (QuickJS über wazero, Browser) sollen möglich bleiben. |
| Nachrichten | Programme senden nicht selbst: Export-Funktionen geben `msg.Out[T]` zurück, der Host verschickt. Eingehende Nachrichten rufen Handler `func(s *State, m Record\|Union)`. |
| Header | Vertrauenswürdige Pakete (`//gostrict:header`) binden Host-Funktionen; Körper ungeprüft, Grenze geprüft. Handles sind JS-Objekte selbst. |
| Laden | Ohne Go-Toolchain zur Laufzeit: eingebettete Stubs der erlaubten stdlib, Programme aus `fs.FS`. |

---
| Bundle im Browser | `fetch` des Manifests (`/api/nago/v1/gostrict/{hash}.json`, immutable) und indirektes `eval` mit Source Map. Bibliotheken (THREE) als Parameter des Wrappers, nie global. nago setzt keine CSP; eine spätere CSP braucht `'unsafe-eval'`. |
| Vertrauen | Header-Pakete nur aus Modulen mit `Module.Trusted` (Disk über `Load`, `std`, `lib`). Programme von Nutzern oder LLMs sind nie trusted (GS039). `lib/*` ist eingebettet und immer verfügbar. |
| Transport | `AsyncCall`: `GostrictListen` (wiederholbar, Antworten `GostrictReceived`) und `GostrictDeliver` pro Komponenten-ID; Werte positionsweise als `GostrictValue` (Floats binär, -0 als `GostrictNegZero`). |
| Zustellung | Höchstens einmal, FIFO je Richtung; das Frontend puffert bis zu 10 000 Nachrichten für eine noch nicht gemountete Komponente. Programme fragen ihren Zustand im `Init` per Nachricht an (Muster `Hello`). |
| Eingaben | Eine Schnittstelle: `lib/input.Event` spiegelt `core.InputEvent` (Test `TestInputMirrorsCore`), dieselben DOM-Listener (`shared/inputEvents.ts`). Pointer-Moves pro Pointer und Wheel pro Frame zusammengefasst. Wheel: X/Y sind jetzt die Position, die Beträge stehen in `DeltaX/Y/Z`. |

## 4. Stand und Architektur

### Verzeichnisse

```
pkg/gostrict/
  doc.go                 Spezifikation des Subsets (= LLM-Prompt)
  load.go stub.go        Laden aus fs.FS-Modulen (Module.Trusted), eingebettete std/, lib/ und _stubs/
  check*.go guard.go     Checker; check_term.go: Terminierungsprofil (GS040)
  fix.go                 maschinell anwendbare Edits der Diagnosen, ApplyEdits (CLI: gostrict fix)
  sealed.go handle.go    Summentypen (auch *T-Varianten, generisch, vergleichbar), Enums, Handles
  emit*.go shake.go      Emitter; emit_copy.go: Kopien nur gelesener Structs entfallen; sourcemap.go
  reload.go              Zustandsübernahme neuer Versionen (Bundle.States, CanTakeOver)
  boundary.go engine.go  Grenztypen, Records auch an der Header-Grenze
  gen.go                 Generator inkl. TS-Deklaration mit Record-Interfaces
  cache.go pool.go       Cache, Pool (sticky für Zustände)
  std/{list,dict,opt,text,num,msg,fuel,js}
  lib/three lib/canvas   Header mit Handles und nativer Implementierung (Szenengraph bzw. Befehlsprotokoll)
  lib/input lib/asset    Records der Host-Events (Eingaben, geladene Bilder/Texturen/Modelle, Atlas)
  internal/examples/*    u. a. draw (Canvas), trees (opt.Option), geom/host (Records an der Header-Grenze)
  testdata/llm2/*        Programme der zweiten LLM-Messung
presentation/ui/program  Bundle-Endpunkt, Spec, Channel (pro Fenster), Codec, Engine (Programme im Browser)
presentation/ui/threejs  TThreeJS
presentation/ui/canvas2  TCanvas
web/vuejs/src/components/gostrict  session.ts (Laden, Werte, Queue), run.ts (Frame-Loop, Resize, Input),
                                   channels.ts, assets.ts (Bilder), engine.ts
web/vuejs/src/components/{threejs,canvas2}  UiThreeJS.vue, UiCanvas2.vue
web/vuejs/src/shared/inputEvents.ts gemeinsame DOM-Listener für nago-Listener und Programme
example/cmd/tutorial-117-threejs, -118-canvas2-platformer, -119-canvas2-photobook
```

### Tests und Prüfungen

```bash
cd pkg/gostrict
"$(go env GOROOT)/bin/gofmt" -l .
go vet ./...
go test -race -count=1 ./...
go generate .            # regeneriert _stubs und gostrict_gen.go der Beispiele und lib/*
cd ../.. && go test -race ./presentation/ui/program/ ./nagotest/
cd web/vuejs && npx vue-tsc --noEmit -p tsconfig.json --composite false   # Altlasten außerhalb gostrict ignorieren
```

Neu: `TestCanvas` (nativ, goja+Host, goja+Bindings gegen Fake-Context), `TestThreeBindings` (Fake-THREE),
`TestFromFloat` (+ 40 000 Fälle in V8 gegen Go geprüft), `TestFix`, `TestSemantics` (Mehrfachzuweisung,
Schleifenvariablen, gelabelte Sprünge, Pointer-/generische Varianten, Union-Vergleich, Aliasing), `TestTakeOver`,
`TestSourceMap`, `TestStickyPool`, `TestUntrusted*`, `presentation/ui/program`: `TestChannel`, `TestEngine`,
`TestBundleEndpoint`, `TestInputMirrorsCore`; Tutorials: `game.TestPlay`, `book.TestEdit`.

Browser-Nachweise (Browser-Pane): Tutorial 117 mit zwei Instanzen, Klick → Backend → Farbe zurück; Tutorial 118
mit Sprite-Sheet über `asset.Loaded`; Tutorial 119 rendert; Browser-Engine (`Fib(25)`, Zustand über zwei Aufrufe).
Frame-Zeit spin-Szene ~0,14 ms. 10 000 `FillRect` pro Frame: 4,1 ms über Bindings, 4,2 ms direkt.
Hinweis: Ist das Browser-Pane ausgeblendet, drosselt Chrome `requestAnimationFrame` auf ~2 fps.

## 5. Stand der Aufgaben

### Erledigt

- **5.0** Endpunkt, JS-Session, Protokoll, Go-Channel pro Fenster, nagotest, E2E im Browser.
- **A** `TThreeJS` + `UiThreeJS.vue` (three 0.186.1, eigener Chunk), Dispose von Geometrien/Materialien/Texturen
  und `forceContextLoss`, Tutorial 117. Nicht gemessen: Speicher beim Seitenwechsel.
- **B** `lib/three`: Remove, Sichtbarkeit, Namen, Plane/Cylinder/Cone/Torus, BufferGeometry, Line, Basic/Phong/
  Physical/Line-Material, Opazität, Wireframe, Emissive, Textur (`LoadTexture`, `SetRepeat`), glTF (`LoadModel`),
  Point/Spot-Light, Schatten, Ortho-Kamera, `Pick` (Raycast über Namen), Dispose; Fake-THREE-Test.
  OrbitControls bewusst nicht (Steuerung im Programm aus Pointer-Events).
- **C** `lib/canvas` (vollständige Context2D-API inkl. Path2D, Gradienten, Patterns, Text, MeasureText, Bilder),
  native Aufzeichnung, `TCanvas` + `UiCanvas2.vue` (DPR im Host), Tutorials 118 und 119.
- **D** Bilder als Handle + `asset.Loaded`/`asset.Failed`, nur gleicher Origin, `createImageBitmap` (EXIF,
  Herunterskalieren > 4096 px), Cache pro Komponente, Freigabe beim Unmount, `asset.Atlas`/`Grid`, Texturen.
  Fotobuch: Upload per Button; Drag-and-drop von Dateien auf den Canvas fehlt noch.
- **E** siehe Entscheidungen; Pointer Capture, Fokus bei Pointer-Down, `touch-action: none`, Wheel ohne Scrollen.
- **F** Push aus anderen Fenstern (118 Highscore, 119 Layout).
- **I.1** Kopien entfallen bei Parametern/Empfängern, die nur gelesen werden, wenn das Argument eine lokale, nicht
  von Closures erfasste Variable ist (generics 17→13, physics 11→10 `$c(`-Vorkommen).
- **I.2** `opt.Option[T]`, **I.3** `text.FromFloat`, **I.4** Edits + `gostrict fix`, **I.5** Records/Listen/Maps an
  der Header-Grenze (goja: plain objects), **I.7** Terminierungsprofil (`//gostrict:terminating`,
  `Program.Terminating`, `std/fuel`, GS040), **I.10** Source Maps (Importpfade, keine Server-Pfade),
  **I.11** Hot Reload mit Zustandsübernahme (ohne Handles/Interfaces/Funktionen im Zustand), **I.12** Browser-Engine
  (`program.Engine`), **I.13** sticky Pool, **I.14** Pointer-Varianten, generische versiegelte Interfaces,
  Vergleich von Unions, **I.15** GS023 aufgehoben, **I.16** Mehrfachzuweisung und `for`-Init mit Tupel korrigiert.
- **Sicherheit:** Header-Lücke geschlossen (GS039).

### LLM-Messung (I.17, 15.10.2026)

Drei Subagenten, die nur `doc.go` und `go doc` der Bibliotheken lasen (Programme in `testdata/llm2`):

| Aufgabe | Runden bis ok | Diagnosen |
|---|---|---|
| `workflow`: Summentypen, Enums, `opt`, Nachrichten als versiegelte Interfaces | 2 | GS006 ×1 (`math.IsInf(x, 0)`) |
| `paint`: canvas2, Multi-Touch, Undo, Pattern aus Bild, Nachrichten | 1 | – |
| `solar`: three.js, Gruppen, Linien, Pick, Textur, Nachrichten | 1 | – |

Zum Vergleich die erste Messung: mehrere Runden, vor allem `int`, `len` und Formatierung von Zahlen. Umgesetzt aus den
Rückmeldungen: konstante `int`-Argumente erlaubter std-Funktionen (`math.IsInf`, `math.Inf`), Doku zu Handlern
(Signatur statt Name, mehrere `msg.Out`, leere Records, Handles im Zustand, Frozen-Tabellen), zu Feldtypen von
Records und Unions als Nachrichten, Feldtabelle in `lib/input`, Frame-Zustand in `lib/canvas`, `Show`/`Pick`/Linien/
Emissive in `lib/three`, URL der Asset-Events.

### Offen (Abstimmung nötig)

- **G** Use-Case-Vertrag für wokoda: Records an der Header-Grenze sind da; offen sind typisierte DB-Zeilen
  (generische Header-Funktionen über Record-Typen) und der Kontext pro Aufruf (Nutzer, Transaktion), der bei geteilten
  Runtimes nicht in Package-Variablen liegen darf.
- **H** `ui`-Header für nago-Views (laut Plan erst nach Abstimmung).
- **I.6** `go/analysis`-Analyzer (neue Abhängigkeit `golang.org/x/tools`).
- **I.8** Verträge und **I.9** Wertebereiche: Sprachdesign offen (Syntax, Beweisumfang, Property-Tests).
- **I.12** QuickJS über wazero (neue Abhängigkeiten).
- Fotobuch: Drag-and-drop von Dateien direkt auf den Canvas.

### J. Pflege

- `doc.go` bei jeder Spracherweiterung aktualisieren (es ist der LLM-Prompt).
- Neue Diagnosen immer mit Testfall in `testdata/bad/*` (`// want GSxxx`) und konkretem `fix`, wo möglich mit Edits.
- Neue Beispiele unter `internal/examples` brauchen einen generierten `gostrict_gen.go` (`go generate`) und einen
  Differenztest nativ gegen goja.
- `TestNoToolchain` muss grün bleiben: kein `go/build`, `go/importer`, `os/exec` im Compiler.
