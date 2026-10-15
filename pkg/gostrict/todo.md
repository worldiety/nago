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

Es ist **noch nichts committet.** `pkg/gostrict/` ist komplett neu (untracked), `go.mod`/`go.sum` enthalten die neue
Abhängigkeit `github.com/dop251/goja v0.0.0-20260917113740-793a2a65c13b`. Vor dem ersten Commit beim Nutzer nach
Freigabe und gewünschtem Zuschnitt fragen. Vorschlag für den Zuschnitt: (1) Abhängigkeit, (2) Compiler-Kern mit
stdlib und Stubs, (3) Engines, Generator, Session, (4) `lib/three` und Beispiele.

---

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

## 4. Stand und Architektur

### Verzeichnisse

```
pkg/gostrict/
  doc.go                 Spezifikation des Subsets (= LLM-Prompt)
  load.go stub.go        Laden aus fs.FS-Modulen, Stubs von Header-Paketen, eingebettete std/ und _stubs/
  check*.go guard.go     Checker: Whitelist, Totalität, Guards, Globals, Header-Grenze, Switches
  sealed.go handle.go    Summentypen/Enums, Handles und Generics an der Header-Grenze
  diag.go                Diagnosen (Codes, Text, JSON)
  emit*.go shake.go      Emitter nach JS (Monomorphisierung, Tree-Shaking der Laufzeit)
  runtime.js             JS-Laufzeit (nur Semantik, keine Prüfungen)
  boundary.go engine.go  Grenztypen (Type, ValueKind, Record, Union, Token, Variant), Engine-Interface
  gen.go                 Generator: Host()/Exports()/Client/Backend/Dispatch/XMessage, TS-Deklaration
  cache.go pool.go       Übersetzungscache (LRU nach Quell-Hash), Pool, Bundle.Expect
  session.go             Nachrichten: Message, Session, Handler-Zuordnung
  engine/gojaengine      goja-Engine (Programm-Cache, gemeinsame Runtimes, CheckBindings-Fuzz)
  engine/native          native Engine über generierte Exports
  std/{list,dict,text,num,msg,js}   Bibliothek des Subsets (Header-Pakete, total)
  lib/three              three.js-Header mit Handles und nativem Szenengraph (native.go, //go:build !gostrict)
  internal/examples/*    Beispiele (je mit generiertem gostrict_gen.go)
  testdata/bad/*         Diagnosetests mit // want GSxxx
  testdata/llm/*         drei Programme, die ein Subagent nur mit doc.go geschrieben hat
  _stubs/                generierte Stubs von math, strings, errors, iter (go generate)
  cmd/gostrict           CLI: check [-json], build, gen, run
```

### Wichtige Mechanismen

- **Übersetzung:** jede Funktion einmal pro Kombination von Typargumenten (Monomorphisierung). Structs sind
  JS-Klassen mit `$z` (Nullwert), `$c` (Kopie), `$s` (in-place zuweisen), `$e` (Vergleich). Interfaces sind Boxen
  `{v, m}` mit Methodentabelle; ein Type Switch vergleicht `box.m`.
- **Bundle:** IIFE `(function($host){...})`, liefert `{Exports..., $new, $unions, $externs}`. Enthält nur die
  genutzten Teile der Laufzeit. `Bundle.Hash` identifiziert es (Engine-Cache, unveränderliche URL).
- **Grenzwerte** (`ValueKind`): bool, int32, uint32, float64, string, error, handle, list, record, state, out, map,
  union. Neutrale Go-Werte: `[]any` für Listen/Records (Felder in Schema-Reihenfolge), `[]any{k, v}`-Paare für Maps,
  `gostrict.Token` für Zustände, `gostrict.Variant{Tag, Fields}` für Unions, `js.Ref` für Handles.
- **Generator:** `gostrict gen <dir>` bzw. `go generate` in `pkg/gostrict` (Tests `TestStubs -update`,
  `TestGenerated -update`). Generierte Dateien tragen `//go:build !gostrict` und werden vom Loader ignoriert.
- **Header-Pakete:** Funktionen mit `//gostrict:js <Funktionsausdruck>`; Parameter nur Basistypen und Handles,
  Ergebnisse zusätzlich `error`; Generics nur über endliche Unions von Grenztypen (`[T Mesh | Group]`), je ein
  nativer Adapter pro Instanziierung. Go-API nur für native Hosts gehört in Dateien mit `//go:build !gostrict`.
- **Engines:** `gojaengine.Engine` (Programm-Cache per Hash, `NewRuntime()` für viele Bundles in einer Runtime,
  `ctx`-Abbruch), `native.Engine{Exports: pkg.Exports()}`. `gostrict.NewPool` für zustandslose Parallelität.
- **Session:** koppelt Instanz und Zustand an einen Sende-Callback, serialisiert Aufrufe, verschickt Ausgänge nach
  dem Aufruf (Backend darf direkt aus dem Callback antworten).

### Tests und Prüfungen

```bash
cd pkg/gostrict
"$(go env GOROOT)/bin/gofmt" -l .
go vet ./...
go test -race -count=1 ./...
go generate .            # regeneriert _stubs und gostrict_gen.go der Beispiele
```

Wichtige Tests: `TestDifferential` (nativ gegen goja), `TestDiagnostics` (`testdata/bad`, `// want`),
`TestThree` (Szenengraph über beide Engines), `TestSession` (Nachrichten), `TestUnions`, `TestRecordsAndState`,
`TestNoToolchain` (keine Toolchain-Abhängigkeit zur Laufzeit), `TestStubs`/`TestGenerated` (Generate aktuell).

Browser-Nachweis: `spin` lief mit three.js 0.160 (jsdelivr UMD) im Browser, `Frame()` etwa 17 µs bei 13 Meshes.
Die Testseite wurde nur in der Scratchpad gebaut; eine dauerhafte Demo entsteht mit Aufgabe A.

---

## 5. Offene Aufgaben

Reihenfolge = empfohlene Priorität. A, C, D, E und F hängen zusammen und teilen sich die Infrastruktur
(Bundle-Auslieferung, JS-Session, Event-Kanal). Beginne mit dem gemeinsamen Unterbau (Abschnitt 5.0) und stimme das
Protokoll vorher mit dem Nutzer ab.

### 5.0 Gemeinsamer Unterbau für Browser-Programme (Voraussetzung für A, C, D, E, F)

1. **Bundle-Auslieferung in nago:** HTTP-Endpunkt, der ein Bundle per Hash ausliefert, z. B.
   `/api/gostrict/{hash}.js`, mit `Cache-Control: public, max-age=31536000, immutable`. Bundles kommen aus einem
   app-weiten `gostrict.Cache`. Format im Browser: das Bundle ist ein Ausdruck; als Modul oder per
   `(0, eval)(text)` laden (CSP prüfen) – Entscheidung mit dem Nutzer.
2. **JS-Session im Frontend** (`web/vuejs/src/components/gostrict/session.ts` o. ä.): lädt das Bundle, ruft den
   Konstruktor-Export (liefert den Zustand), hält den Zustand, ruft Handler per `Bundle.Handlers`, liest
   `msg.Out`-Ergebnisse (Arrays von Records/Varianten) und schickt sie über das Protokoll. Records werden
   positionsweise nach `Bundle.Records` (de)codiert; Unions über `$unions` (Methodentabelle → Variante, Konstruktor
   für eingehende Varianten). Diese Metadaten müssen dafür mit dem Bundle in den Browser (z. B. als JSON neben dem
   Bundle oder im Protokoll).
3. **Protokollerweiterung** (Spec in `presentation/proto/spec`, Generator `go generate` in `cmd/nprotoc`):
   mindestens Nachrichten Server→Client und Client→Server mit Komponenten-/Instanz-ID, Nachrichtentyp (voller
   Record-Name oder kompakte ID aus dem Schema) und positionsweisen Feldern. Hinweise zu nprotoc: Records haben
   höchstens 30 Felder (sonst Unterrecords), generierte TS-Arrays haben `.value`. Wire-Format ohne JSON wäre ideal;
   Floats müssen NaN/±Inf/-0 überstehen.
4. **Go-Seite:** pro Fenster eine Session (`core.StateOf`-Muster), Zuordnung von Komponenten-ID zu Session,
   Zustellung serialisiert mit der Event-Loop des Fensters. Nachrichten an das Programm können jederzeit kommen
   (Push aus Use Cases, anderen Fenstern, Timern).
5. **Offene Entscheidungen (mit dem Nutzer klären):** Zustellung bei Reconnect (höchstens einmal vs. erneut),
   Zusammenfassen hochfrequenter Nachrichten („das Neueste gewinnt“ pro Schlüssel, z. B. Feldmarkierung im Record),
   Backpressure, Reihenfolge (FIFO pro Instanz ist gesetzt).
6. **Tests:** Ende-zu-Ende im Browser-Pane (Werkzeuge `mcp__Claude_Browser__*`) mit einem Tutorial; nagotest für
   die Go-Seite.

### A. `ui/threejs.TThreeJS` – nago-Komponente für three.js-Programme

- **Go:** Paket `presentation/ui/threejs` mit `TThreeJS` (Builder-Stil wie andere nago-Komponenten, z. B.
  `ui/sheet.TSheet`): Bundle bzw. Programmquelle, Konstruktor-Export und Argumente, Frame-Export, Größe (`Frame`),
  Nachrichten-Backend (generiertes `Backend`-Interface + `Dispatch` des Programms), optional Hintergrund/Antialias.
- **Protokoll:** eigene Komponente (Spec), referenziert Bundle-Hash, Exportnamen, Initialargumente (als
  Grenzwerte), Komponenten-ID für Nachrichten.
- **Vue:** `web/vuejs/src/components/threejs/UiThreeJS.vue`, lazy registriert in `web/vuejs/src/shared/uiComponents.ts`
  (wie `LazyUiSheet`). three.js als npm-Abhängigkeit (Version festlegen, mind. zwei Wochen alt), als eigener Chunk
  lazy geladen (Muster: Vite-Alias wie bei `@material/material-color-utilities/hct/hct.js` in `vite.config.ts`).
  Der Host im Browser stellt `show(scene, camera)` bereit (siehe `lib/three.Show`), rendert per
  `requestAnimationFrame`, ruft den Frame-Export mit der Zeit, reagiert auf Resize (Export für Seitenverhältnis),
  räumt beim Unmount auf (`dispose` von Renderer, Geometrien, Materialien, Texturen).
- **Tutorial:** `example/cmd/tutorial-117-threejs` (z. B. `spin` plus Nachrichten an das Backend).
- **Abnahme:** läuft im Browser ohne Konsolenfehler, Frame-Zeit gemessen, Nachrichten in beide Richtungen,
  mehrere Instanzen pro Fenster, kein Speicherleck beim Wechsel der Seite.

### B. `lib/three` vervollständigen

Aktuell: Scene, Camera, Group, Mesh, Light, Geometry, Material; Box/Sphere/TorusKnot; StandardMaterial;
Directional/Ambient Light; Add, SetPosition/Rotation/Scale, LookAt, SetBackground, SetAspect, Show.

Fehlt (Vorschlag, mit Nutzer priorisieren):

- Objekte entfernen (`Remove`), `Dispose` für Geometrie/Material/Textur, Sichtbarkeit, Namen.
- Weitere Geometrien (Plane, Cylinder, Cone, Torus, BufferGeometry aus Punktlisten), Materialien (Basic, Phong,
  Physical, Linien), Farbe/Opazität/Wireframe/Emissive, Schatten (castShadow/receiveShadow, Light-Shadows).
- Texturen und Bilder (siehe D): `LoadTexture(url) Texture` als Handle, Zustand über Events.
- Orthographische Kamera, Kamera-Steuerung (OrbitControls über `three/examples/jsm` oder eigene Logik im Programm
  aus Pointer-Events – im Programm bevorzugt, da testbar).
- Picking/Raycasting: `Raycast(scene, camera, x, y) (Hit, bool)` – braucht Records über die Header-Grenze
  (siehe I.5) oder mehrere Ergebnisse aus Basistypen.
- Modelle (GLTF-Loader) als asynchrones Laden mit Event.
- Native Implementierung (`native.go`) jeweils mitziehen, damit Programme nativ und in goja testbar bleiben.
- `CheckBindings` überspringt Funktionen mit Handles; für `lib/three` einen Test mit einem minimalen Fake-`THREE`
  in goja ergänzen, der Aufrufe protokolliert und mit dem nativen Szenengraph vergleicht.

### C. `ui/canvas2.TCanvas` – 2D-Canvas mit gleicher Architektur und Performance

Das bestehende `presentation/ui/canvas` (+ `web/vuejs/src/components/canvas/UiCanvas.vue`) streamt einzelne
Zeichenbefehle per `AsyncCall` vom Server (immediate mode mit Display-Listen). `canvas2` ist ein **neues, eigenes**
Paket; das alte bleibt unverändert.

- **`lib/canvas` (Header)** mit vollständiger Context2D-API, Handles `Context`, `Image`, `Gradient`, `Pattern`,
  `Path2D`:
  - Zustand: `Save/Restore`, `SetTransform/Transform/Translate/Rotate/Scale/ResetTransform`, `GlobalAlpha`,
    `GlobalCompositeOperation`, `ImageSmoothing`.
  - Stile: `FillColor/StrokeColor` (uint32 RGBA oder string), Gradienten (linear, radial, conic,
    `AddColorStop`), Patterns, `LineWidth/Cap/Join/MiterLimit`, `SetLineDash/LineDashOffset`, Schatten.
  - Pfade: `BeginPath/ClosePath/MoveTo/LineTo/BezierCurveTo/QuadraticCurveTo/Arc/ArcTo/Ellipse/Rect/RoundRect`,
    `Fill/Stroke/Clip` (Füllregel), `Path2D`, `IsPointInPath/IsPointInStroke`.
  - Rechtecke: `FillRect/StrokeRect/ClearRect`.
  - Text: `Font`, `TextAlign`, `TextBaseline`, `Direction`, `FillText/StrokeText` (mit maxWidth),
    `MeasureText` (braucht mehrere Ergebnisse oder Records über die Header-Grenze, siehe I.5).
  - Bilder: `DrawImage` in allen drei Formen (inkl. Quell-Rechteck für Sprite-Sheets), `ImageData` vorerst nicht.
  - Gerätepixel: `Show(ctx)`/Größe/`devicePixelRatio`-Handhabung im Host, nicht im Programm.
- **Native Implementierung:** Aufzeichnung der Befehle als Liste (für Tests: Vergleich der Befehlsfolgen nativ vs.
  goja; optional später Rasterung).
- **Performance:** Bindings sind dünne Funktionsausdrücke (`(c, x, y, w, h) => c.fillRect(x, y, w, h)`), die V8
  inlinen kann; keine Allokation pro Aufruf. Messen: z. B. 10 000 `FillRect` pro Frame.
- **Komponente** analog zu A: `presentation/ui/canvas2`, `UiCanvas2.vue`, Protokoll-Spec, Frame-Loop, Resize,
  Nachrichten, Events (E), Bilder (D).
- **Tutorials:** `tutorial-118-canvas2-platformer` (2D-Plattformer mit Tilemap, Sprite-Sheet, Kollision,
  Nachrichten an Backend für Highscore/Multiplayer) und `tutorial-119-canvas2-photobook` (Doppelseiten-Editor:
  Bilder per Drag-and-Drop/Upload, Verschieben/Skalieren/Drehen, Snapping, Ebenen, Speichern über Nachrichten).

### D. Spezialisierte Bild-Ladelogik

- Programme laufen synchron und ohne Promises. Laden ist deshalb **Handle + Event**:
  `img := canvas.LoadImage(url)` liefert sofort ein Handle; der Host meldet `ImageLoaded{...}` bzw.
  `ImageFailed{...}` als Event (siehe E) an einen Handler des Programms. Zeichnen eines noch nicht geladenen
  Bildes ist ein No-op (total).
- Quellen: nago-Ressourcen-URIs (Blob-Store, Uploads, eingebettete Assets), keine beliebigen fremden URLs ohne
  Freigabe (CORS/Sicherheit). Größen/Varianten (Thumbnails, Retina) über nago-Image-Endpunkte; Dekodierung im
  Browser mit `createImageBitmap` (off-main-thread), Cache pro Komponente bzw. Fenster, Freigabe beim Unmount.
- Für Sprite-Sheets: Atlas-Beschreibung als Record (Frames als Liste von Rechtecken).
- Für das Fotobuch: Upload über nago (Drag-and-Drop von Dateien → Upload → URL → Nachricht an Programm), EXIF-
  Orientierung beachten, große Bilder herunterskalieren.
- Gleiche Mechanik für three.js-Texturen (B).

### E. Event-Listener-Mechanik (Eingaben des Hosts an das Programm)

- **Modell:** Events des Hosts sind Records, die wie Nachrichten an Handler gehen – nur kommen sie vom Host statt
  vom Backend. Vorschlag: Standard-Records in einer Bibliothek, z. B. `lib/input`:
  `Pointer{Kind, ID, X, Y, Buttons, Pressure, Shift, Ctrl, Alt, Meta}` (down/move/up/cancel, Mehrfach-Touch über
  `ID`), `Wheel{DX, DY, DZ, Mode}`, `Key{Kind, Key, Code, Repeat, Modifiers}`, `Resize{W, H, DPR}`,
  `Visibility{Visible}`, `Focus{}`, `Drop{Files ...}`. Ein Programm deklariert nur die Handler, die es braucht
  (`OnPointer(s *S, e input.Pointer)`); der Host registriert nur Listener, für die es einen Handler gibt
  (`Bundle.Handlers` zeigt das).
- Alternativ ein versiegeltes `input.Event` mit einem Handler; beides ist mit dem vorhandenen Mechanismus möglich.
  Entscheidung mit dem Nutzer.
- **Zusammenfassen:** Pointer-Moves und Wheel pro Frame bündeln (nur das letzte bzw. aufsummiert), damit ein Frame
  nicht von Hunderten Events blockiert wird; Down/Up nie verwerfen.
- **Koordinaten** im Canvas-Koordinatensystem (CSS-Pixel), `DPR` separat.
- **Pointer Capture**, `touch-action: none`, Kontextmenü unterdrücken optional, Tastatur nur bei Fokus.
- **Hit-Testing** gehört ins Programm (rein, testbar): für 2D ein Quadtree/Grid in Go; für 3D Raycasting über den
  Header (B).
- Die JS-Session (5.0) stellt Events und Backend-Nachrichten in **einer** Warteschlange zu, zwischen den Frames.

### F. Nachrichten-Transport über den nago-Websocket

Siehe 5.0 (Protokoll, JS-Session, Go-Session pro Fenster). Zusätzlich:

- Generierter Backend-Code existiert (`XMessage`, `Backend`, `Dispatch`). Ein Programm und sein Backend importieren
  dasselbe Nachrichtenpaket (Beispiel `internal/examples/arena/proto`).
- Ende-zu-Ende-Test: `arena`-ähnliches Programm in einer `TThreeJS`- oder `TCanvas`-Komponente, Backend im Go-Teil
  der App, Push aus einem Use Case.

### G. Use-Case-Vertrag für wokoda (nur in nago entwerfen, wokoda nicht ändern)

Erkenntnisse aus wokoda (`app/database/script_runtime.go`, `views_runtime.go`, `views_bridge*.go`):
Kompilieren und neue goja-Runtime pro Aufruf und pro verschachteltem `ctx.call`, eine Runtime pro Skript und
Fenster (~91 KB) als Schutzgrenze, JSON-Roundtrips an jeder Grenze, Eingabeschema zur Laufzeit, Ausgaben untypisiert,
UI-Baum als `{t, a, m}`-Knoten mit Allowlist-Prüfung in Go.

- Vertrag: `func Main(ctx Ctx, in Input) (Output, error)` mit Records; Schema aus `Bundle.Records`; Prüfung per
  `Bundle.Expect`.
- `Ctx` als Header-Paket mit typisierten Funktionen (DB-Abfragen, Aufruf anderer Use Cases, Nutzerinfo). Für
  typisierte DB-Zeilen braucht es Records über die Header-Grenze (I.5) oder generische Header-Funktionen über
  Record-Typen.
- `gostrict.Cache` + gemeinsame Runtime pro Fenster ersetzen Neukompilieren und eine Runtime pro Skript.

### H. `ui`-Header für nago-Views (Prototyp)

- Handles auf nago-Views; Builder als Header-Funktionen; Typen garantieren gültige Views, keine Allowlist zur
  Laufzeit. Interaktionen über Handler/Events. Erst nach Abstimmung beginnen.

### I. Compiler, Checker, Bibliothek

1. **Unnötige Struct-Kopien vermeiden:** Analyse „Funktion verändert Parameter/Empfänger nicht“ (Reinheit/Effekte),
   dann Kopien bei Aufrufen streichen. Messbar an `generics`/`physics`-Bundles.
2. **`opt.Option[T]`** (z. B. `std/opt`): nötig für rekursive Strukturen (`Node{Next *Node}` ist heute nicht
   baubar, weil `Next` nil wäre). Zugriff nur über `Get() (T, bool)` mit Guard; Nullwert = None ist erlaubt.
3. **`text.FromFloat(f, decimals)`** mit in Go und JS identischer Rundung (Ties-to-even auf der exakten Binärzahl;
   `toFixed` rundet anders). Größter Ergonomie-Wunsch aus der LLM-Messung.
4. **Maschinell anwendbare Fixes** in Diagnosen (Textänderungen im JSON), z. B. `int`→`int32`,
   `len(s)`→`text.RuneCount(s)`.
5. **Records über die Header-Grenze** (Ein- und Ausgaben von Host-Funktionen), z. B. für `MeasureText`, Raycast,
   DB-Zeilen. Die Konverter existieren für Exporte; für Header fehlen Schema im Bundle und Adapter.
6. **Integration als `go/analysis`-Analyzer**, damit Diagnosen in gopls/`go vet` erscheinen (Abhängigkeit
   `golang.org/x/tools` mit Nutzer klären).
7. **Optionales Terminierungsprofil:** nur beschränkte Schleifen (`range` über endliche Mengen, Zähler mit
   Schranke, `num.Fuel`), Rekursion verboten oder mit abnehmendem Argument.
8. **Verträge** (`contract.Requires(...)`) als modulare Guards; Unbeweisbares als generierte Property-Tests.
9. **Wertebereiche** (`//gostrict:range 0..100`) zum statischen Beweis von Indizes.
10. **Source Maps / Positionszuordnung** JS→Go für Debugging im Browser.
11. **Hot Reload mit Zustandsübernahme:** Zustandstyp per Hash vergleichen, Token an neue Version übergeben.
12. **Weitere Engines:** QuickJS über wazero (mit Speicherlimit, das goja fehlt), Browser-Engine (Go-Stub, der über
    nago ausführt).
13. **Pool mit Zuständen:** Zustände sind an ihre Instanz gebunden; ggf. „sticky“ Pool.
14. **Summentypen:** Varianten mit Pointer-Empfängern, generische versiegelte Interfaces, Vergleich von
    Union-Werten (#76920: vergleichbar, wenn alle Varianten es sind) – heute GS018. Bei Annahme von #76920 dessen
    Syntax unterstützen.
15. **GS023** (labeled `break`/`continue` aus Range über Iteratoren) durch Flags im Callback aufheben.
16. **Bekannte Schwächen:** Mehrfachzuweisung wertet linke Seiten nicht exakt nach Go-Spezifikation aus;
    `for`-Init mit Tupel-Aufruf verliert Variablen pro Iteration; Transzendente (`math.Sin` u. a.) weichen zwischen
    Browsern um ULPs ab (akzeptiert).
17. **LLM-Messung wiederholen** mit Unions, Nachrichten, Globals-Regel und Canvas/three-Aufgaben (Subagent, der nur
    `doc.go` und die öffentliche API liest; Runden und Diagnosecodes zählen).

### J. Pflege

- `doc.go` bei jeder Spracherweiterung aktualisieren (es ist der LLM-Prompt).
- Neue Diagnosen immer mit Testfall in `testdata/bad/*` (`// want GSxxx`) und konkretem `fix`.
- Neue Beispiele unter `internal/examples` brauchen einen generierten `gostrict_gen.go` (`go generate`) und einen
  Differenztest nativ gegen goja.
- `TestNoToolchain` muss grün bleiben: kein `go/build`, `go/importer`, `os/exec` im Compiler.

---

## 6. Wo man anfängt

1. `doc.go` lesen, dann `internal/examples/arena` (Nachrichten), `internal/examples/spin` + `lib/three`
   (Handles/Generics), `internal/examples/shapes` (Unions/Enums/Maps), `internal/examples/orders` (Records/Zustand).
2. Tests laufen lassen (Abschnitt 4).
3. Abschnitt 5.0 mit dem Nutzer abstimmen (Protokoll, Bundle-Auslieferung, offene Entscheidungen), danach A oder C.
