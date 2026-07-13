---
# Static content
title: "Tutorial 104: Sandbox Exec (sbox)"
weight: 104
---

# Tutorial 104 – untrusted Prozesse sicher ausführen mit `pkg/sbox`

Dieses Beispiel zeigt, wie eine Nago-Anwendung untrusted Prozesse (den
Go-Compiler, `git`, `go test`/`vet` oder fremde Programme, die untersucht werden
sollen) in einer engen Sandbox startet, ohne dass diese Prozesse an die Secrets
und die Datenbank im `DataDir()` der App gelangen.

Die Sandbox nutzt ausschließlich Linux-Kernel-Primitive – User-/Mount-/PID-/
IPC-/UTS-/Netz-Namespaces, `pivot_root`, Landlock und seccomp-bpf – und kommt
ohne externe Programme wie `bubblewrap` aus. Siehe [`pkg/sbox`](../../../pkg/sbox).

Die Seite bietet zwei interaktive Aktionen:

1. **Erlaubter Prozess** – `go version` läuft vollständig isoliert und die
   Ausgabe wird angezeigt.
2. **Sicherheitsbeweis** – derselbe Sandbox-Prozess versucht, das App-Secret aus
   dem `DataDir()` zu lesen, und scheitert: der Pfad existiert im
   Mount-Namespace der Sandbox nicht und Landlock verweigert zusätzlich jeden
   Zugriff.

## Wichtig: `sbox.Init()` muss zuerst laufen

`sbox` verwendet ein Re-Exec-Trampolin (wie `nsjail`/`bubblewrap`). Damit das
funktioniert, **muss** `sbox.Init()` die allererste Anweisung in `main()` sein:

```go
func main() {
    sbox.Init() // No-op im Elternprozess; übernimmt im Sandbox-Kind und kehrt nie zurück
    // ... application.Configure(...).Run()
}
```

## Plattform

Eine echte Sandbox gibt es nur auf **Linux** (Ziel: Ubuntu 24.04+, Kernel 6.8,
Landlock ABI v4+). Auf einem Nicht-Linux-Dev-Host (z. B. macOS) fällt `sbox` auf
einen **UNSANDBOXED-Passthrough** zurück, der bei jedem Aufruf laut warnt. Der
Sicherheitsbeweis schlägt dort bewusst fehl. Produktion ist immer Linux, wo der
unsichere Stub gar nicht erst mitkompiliert wird.

Zusätzlich werden **unprivilegierte User-Namespaces** benötigt. Auf Ubuntu 24.04
können diese per AppArmor gesperrt sein
(`kernel.apparmor_restrict_unprivileged_userns`). Ist das der Fall, liefert
`sbox.Run` einen klaren Fehler (`ErrNoUserNamespace`) statt schwächer zu
isolieren.

## Lokal ausführen (Linux)

```bash
go run go.wdy.de/nago/example/cmd/tutorial-104-sandbox-exec
```

Danach `http://localhost:3000` öffnen und die beiden Buttons ausprobieren.

Auf macOS lässt sich das Beispiel ebenfalls starten (zum UI-Entwickeln), aber
der Sicherheitsbeweis schlägt wegen des Passthroughs fehl – das ist gewollt.

## Reproduzierbar mit Docker

Da die Sandbox Linux-spezifisch ist, lässt sie sich am einfachsten in einem
Container nachvollziehen. Die Sandbox benötigt die Fähigkeit, unprivilegierte
User-Namespaces zu erzeugen sowie `mount`/`pivot_root` durchzuführen; im
Container erreicht man das prod-nah mit `--privileged`.

> Hinweis: `--privileged` wird hier verwendet, damit der Container dieselben
> Kernel-Fähigkeiten hat wie ein realer Linux-Host, auf dem die Nago-App läuft.
> Die Sandbox selbst *entzieht* dem untrusted Prozess diese Fähigkeiten wieder.

### Variante A – schneller, headless Sicherheitsbeweis (empfohlen)

Das Beispiel enthält einen headless Selbsttest (`SBOX_SELFCHECK=1`), der beide
Sandbox-Läufe ohne Web-UI ausführt und mit Exit-Code `0` endet, wenn das Secret
nachweislich geschützt ist.

Vom Repo-Root:

```bash
# 1. Linux-Binary bauen (Architektur des Docker-Hosts wählen: arm64 oder amd64)
GOOS=linux GOARCH=arm64 go build -o /tmp/tut104 \
  ./example/cmd/tutorial-104-sandbox-exec

# 2. Selbsttest im Container ausführen
docker run --rm --privileged \
  -e SBOX_SELFCHECK=1 \
  -v /tmp/tut104:/tut104:ro \
  golang:1.23-bookworm /tut104
```

Erwartete Ausgabe:

```
[go version] err=<nil>
go version go1.23.12 linux/arm64

[cat secret] err=exit code 1 out="/usr/bin/cat: /tmp/tut104-selfcheck/secrets/db-password.txt: No such file or directory\n"
RESULT: ✔ GESCHÜTZT — Secret ist aus der Sandbox nicht erreichbar
```

- `go version` läuft **in der Sandbox** erfolgreich.
- Der Versuch, das Secret zu lesen, scheitert mit *No such file or directory* –
  der Pfad existiert im Mount-Namespace der Sandbox nicht.

Zum Gegenbeweis kann man die Sandbox über den Kill-Switch erzwingen: Setzt man
`SBOX_REQUIRE_ISOLATION=1` und lässt das Binary auf einem Nicht-Linux-Host oder
ohne User-Namespaces laufen, verweigert `sbox.Run` die Ausführung, statt
ungeschützt durchzureichen.

### Variante B – vollständige Web-UI im Container

```bash
GOOS=linux GOARCH=arm64 go build -o /tmp/tut104 \
  ./example/cmd/tutorial-104-sandbox-exec

docker run --rm --privileged \
  -p 3000:3000 \
  -v /tmp/tut104:/tut104:ro \
  golang:1.23-bookworm /tut104
```

Anschließend `http://localhost:3000` im Browser öffnen und die beiden Buttons
betätigen. Der zweite Button zeigt im Ausgabefeld, dass der Leseversuch auf das
Secret scheitert.

### Ganz ohne Host-Toolchain bauen (optional)

Wenn kein lokales Go vorhanden ist, kann auch im Container gebaut werden. Das
kompiliert allerdings den gesamten Nago-Baum und dauert entsprechend; außerdem
muss der Zugriff auf die privaten Module gegeben sein (siehe
[`example/cmd/README.md`](../README.md)). Empfohlen ist daher das Vorbauen des
Binaries auf dem Host (Variante A/B). Zum Bauen im Container:

```bash
docker run --rm --privileged \
  -e SBOX_SELFCHECK=1 \
  -e GOPRIVATE='go.wdy.de/*,gitlab.worldiety.net/*' \
  -v "$PWD":/src -w /src \
  golang:1.23-bookworm \
  sh -c 'go run ./example/cmd/tutorial-104-sandbox-exec'
```

## Was das Beispiel demonstriert

| Mechanismus        | Effekt im Beispiel                                                        |
|--------------------|--------------------------------------------------------------------------|
| Mount-Namespace + `pivot_root` | Nur explizit gebundene Pfade sind sichtbar; das `DataDir()` fehlt vollständig. |
| Landlock           | Zweite FS-Grenze: selbst bei einer Lücke in der Mount-Isolation kein Zugriff aufs Secret. |
| seccomp-bpf        | Gefährliche Syscalls (`ptrace`, `bpf`, `keyctl`, `mount`, `unshare`, …) werden blockiert. |
| User-Namespace     | Rootless-Isolation ohne Host-root; fehlt er, bricht `Run` hart ab.       |
| Netz-Namespace     | Standard hier `NetHost` (go/git dürfen ins Netz); optional `NetLoopback`/`NetNone`. |
| rlimits + Timeout  | Ressourcen-Limits und ein Wall-Clock-Timeout beenden den Prozessbaum.    |

## Verwandte Bausteine

- [`pkg/sbox`](../../../pkg/sbox) – die Sandbox-API (`Profile`, `Run`, `Init`)
  sowie fertige Profile `GoBuild`, `GoTest`, `GoVet`, `Git`, `UntrustedServer`.
