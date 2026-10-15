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

## Aufbau

```
pkg/gostrict/            Compiler (load, check*, emit*, runtime.js), Grenze (boundary, engine, values, router),
                         Session, Cache, Pool, Generator (nur Header), fix, sourcemap, reload
pkg/gostrict/std/        list, dict, opt, text, num, msg, fuel, js, big, decimal; host.go mit //go:build !gostrict
                         für Reflection
pkg/gostrict/lib/        three, canvas (Header), input, asset (Records)
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

## Offen

- **Use-Case-Vertrag für wokoda (G):** typisierte DB-Zeilen über generische Header-Funktionen mit Record-Typen und
  ein Kontext pro Aufruf (Nutzer, Transaktion), der bei geteilten Runtimes nicht in Package-Variablen liegen darf.
  Nur in nago entwerfen.
- **`ui`-Header für nago-Views (H):** erst nach Abstimmung.
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
