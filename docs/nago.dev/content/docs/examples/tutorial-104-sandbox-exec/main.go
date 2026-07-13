// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Tutorial 104 – sbox: untrusted Prozesse sicher aus einer Nago-Anwendung
// heraus ausführen.
//
// Eine Nago-Anwendung hält Secrets und ihre Datenbank im DataDir(). Wenn sie
// untrusted Prozesse startet (den Go-Compiler, git, go test/vet oder fremde
// Programme, die untersucht werden sollen), darf ein solcher Prozess dieses
// DataDir niemals lesen oder anderweitig ausbrechen. Das Paket pkg/sbox baut
// dafür ausschließlich mit Linux-Kernel-Primitiven (user/mount/pid/ipc/uts/net
// Namespaces, pivot_root, landlock, seccomp) eine engere Sandbox – ohne externe
// Programme wie bubblewrap.
//
// Dieses Beispiel demonstriert an einer Seite:
//
//  1. Einen erlaubten Prozess ("go version") erfolgreich in der Sandbox laufen
//     lassen und seine Ausgabe anzeigen.
//  2. Den Kern-Sicherheitsbeweis: derselbe Sandbox-Prozess versucht, das
//     App-Secret aus dem DataDir zu lesen – und scheitert, weil der Pfad im
//     Mount-Namespace gar nicht existiert und landlock zusätzlich kein Recht
//     darauf gewährt.
//
// WICHTIG: sbox benötigt ein Re-Exec-Trampolin. Deshalb MUSS sbox.Init() die
// allererste Anweisung in main() sein. Auf Nicht-Linux-Systemen (z. B. dem
// macOS-Dev-Host) fällt sbox auf einen UNSANDBOXED-Passthrough zurück, der bei
// jedem Aufruf laut warnt – Produktion ist immer Linux.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/sbox"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

// secretRelPath ist der Ort des (Demo-)Secrets relativ zum DataDir der App.
const secretRelPath = "secrets/db-password.txt"

// demo bündelt die zur Laufzeit benötigten Pfade.
type demo struct {
	dataDir    string // DataDir() der App – enthält Secrets und DB
	secretPath string // absoluter Pfad des Demo-Secrets im DataDir
	workDir    string // eine harmlose, beschreibbare Arbeitskopie für die Sandbox
}

func main() {
	// sbox.Init MUSS als Erstes stehen. Im normalen Prozess ein No-op; im
	// re-exec'ten Sandbox-Kind übernimmt es und kehrt nie zurück.
	sbox.Init()

	// Headless-Selbsttest für die reproduzierbare Docker-Demo (siehe README).
	// Läuft ohne Web-UI, führt beide Sandbox-Läufe aus und beendet sich mit
	// Exit-Code 0, wenn das Secret nachweislich geschützt ist, sonst 1.
	if os.Getenv("SBOX_SELFCHECK") == "1" {
		os.Exit(selfcheck())
	}

	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_104")
		cfg.Serve(vuejs.Dist())

		d, err := setup(cfg)
		if err != nil {
			panic(err)
		}

		cfg.RootView(".", func(wnd core.Window) core.View {
			return page(wnd, d)
		})
	}).Run()
}

// selfcheck führt die beiden Sandbox-Läufe headless aus und prüft, dass das
// Secret aus der Sandbox NICHT lesbar ist. Rückgabe ist der Prozess-Exit-Code.
func selfcheck() int {
	dataDir := os.Getenv("SBOX_SELFCHECK_DATADIR")
	if dataDir == "" {
		dataDir = filepath.Join(os.TempDir(), "tut104-selfcheck")
	}
	d, err := setupDir(dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "setup:", err)
		return 1
	}

	ctx := context.Background()

	// 1) erlaubter Prozess
	if goBin := lookPath("go"); goBin != "" {
		out, err := runInSandbox(ctx, d, goBin, "version")
		fmt.Printf("[go version] err=%v\n%s\n", err, out)
	} else {
		fmt.Println("[go version] übersprungen: kein 'go' im PATH")
	}

	// 2) Sicherheitsbeweis
	catBin := lookPath("cat")
	if catBin == "" {
		fmt.Fprintln(os.Stderr, "kein 'cat' im PATH")
		return 1
	}
	out, err := runInSandbox(ctx, d, catBin, d.secretPath)
	fmt.Printf("[cat secret] err=%v out=%q\n", err, out)

	if err == nil && contains(out, "SECRET") {
		fmt.Println("RESULT: ✘ SECRET LEAKED — Sandbox greift NICHT (Nicht-Linux-Passthrough?)")
		return 1
	}
	fmt.Println("RESULT: ✔ GESCHÜTZT — Secret ist aus der Sandbox nicht erreichbar")
	return 0
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// setup legt das Demo-Secret im DataDir an und eine getrennte, harmlose
// Arbeitskopie, die der Sandbox als beschreibbares WorkDir dient.
func setup(cfg *application.Configurator) (*demo, error) {
	return setupDir(cfg.DataDir())
}

// setupDir ist der DataDir-unabhängige Teil von setup und wird auch vom
// Headless-Selbsttest genutzt.
func setupDir(dataDir string) (*demo, error) {
	secretPath := filepath.Join(dataDir, secretRelPath)

	if err := os.MkdirAll(filepath.Dir(secretPath), 0o700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(secretPath, []byte("SUPER-SECRET-DB-PASSWORD\n"), 0o600); err != nil {
		return nil, err
	}

	workDir := filepath.Join(dataDir, "sandbox-work")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return nil, err
	}

	return &demo{dataDir: dataDir, secretPath: secretPath, workDir: workDir}, nil
}

func page(wnd core.Window, d *demo) core.View {
	// Ausgabe-States für die beiden Läufe.
	allowedOut := core.AutoState[string](wnd)
	secretOut := core.AutoState[string](wnd)

	// Wir suchen ein erlaubtes Programm (go, sonst cat) für die harmlose Demo.
	goBin := lookPath("go")

	runAllowed := func() {
		if goBin == "" {
			allowedOut.Set("kein 'go' im PATH gefunden – Beispiel benötigt die Go-Toolchain")
			return
		}
		out, err := runInSandbox(wnd.Context(), d, goBin, "version")
		allowedOut.Set(formatRun("go version", out, err))
	}

	runSecretAttempt := func() {
		// Wir versuchen, das App-Secret aus dem DataDir zu lesen – aus der
		// Sandbox heraus. Das MUSS scheitern.
		catBin := lookPath("cat")
		if catBin == "" {
			secretOut.Set("kein 'cat' im PATH gefunden")
			return
		}
		out, err := runInSandbox(wnd.Context(), d, catBin, d.secretPath)
		leaked := out
		secretOut.Set(formatRun("cat "+d.secretPath, leaked, err) +
			"\n\n" + verdict(out, err))
	}

	return ui.VStack(
		ui.H1("sbox · untrusted Prozesse sicher ausführen"),
		ui.Text("Diese App hält ihr Secret unter DataDir()/"+secretRelPath+
			". Untrusted Prozesse werden über pkg/sbox in einer engen Sandbox "+
			"gestartet und dürfen das Secret niemals sehen.").Font(ui.BodyLarge),

		platformHint(),

		ui.Space(ui.L16),

		explainerCard(
			"1 · Erlaubter Prozess in der Sandbox",
			"„go version“ läuft vollständig isoliert: eigener User-/Mount-/PID-/"+
				"Netz-Namespace, read-only Systempfade, seccomp-Filter, landlock. "+
				"Nur die harmlose Arbeitskopie ist beschreibbar.",
			ui.VStack(
				ui.PrimaryButton(runAllowed).Title("go version ausführen"),
				outputBox(allowedOut.Get()),
			).Gap(ui.L8).Alignment(ui.Leading).FullWidth(),
		),

		ui.Space(ui.L8),

		explainerCard(
			"2 · Sicherheitsbeweis: Secret ist unerreichbar",
			"Derselbe Sandbox-Prozess versucht, das App-Secret aus dem DataDir "+
				"zu lesen. Der Pfad existiert im Mount-Namespace der Sandbox gar "+
				"nicht (er wurde nie hineingebunden), und landlock verweigert "+
				"zusätzlich jeden Zugriff. Der Leseversuch scheitert daher.",
			ui.VStack(
				ui.SecondaryButton(runSecretAttempt).
					Title("Secret aus der Sandbox lesen (muss scheitern)"),
				outputBox(secretOut.Get()),
			).Gap(ui.L8).Alignment(ui.Leading).FullWidth(),
		),
	).Gap(ui.L8).Alignment(ui.Leading).FullWidth().
		Frame(ui.Frame{MaxWidth: ui.L880}).
		Padding(ui.Padding{}.All(ui.L16))
}

// runInSandbox führt ein Programm über pkg/sbox aus und gibt dessen kombinierte
// stdout/stderr-Ausgabe zurück. Es verwendet ein hermetisches Profil, das nur
// die nötigen read-only Systempfade und eine beschreibbare Arbeitskopie
// exponiert – das DataDir mit den Secrets wird bewusst NICHT gebunden.
func runInSandbox(ctx context.Context, d *demo, path string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// CapBuffer begrenzt die eingefangene Ausgabe eines (potenziell
	// fehlerhaften) untrusted Prozesses, damit dieser nicht beliebig viel
	// Speicher im Host belegen kann.
	buf := sbox.NewCapBuffer(64 * 1024)
	p := sbox.Profile{
		RootFS:   sbox.RootMinimal, // read-only /usr,/bin,/lib,... und CA-Zertifikate
		Binds:    []sbox.Bind{{Host: d.workDir, Writable: true}},
		Env:      []string{"PATH=/usr/bin:/bin:/usr/local/go/bin", "HOME=" + d.workDir},
		WorkDir:  d.workDir,
		Net:      sbox.NetHost, // go/git dürfen ins Netz; Secrets bleiben trotzdem unerreichbar
		Seccomp:  sbox.SeccompStrict,
		Landlock: true,
		Limits:   sbox.Limits{Wall: 30 * time.Second, NoFile: 4096},
	}

	res, err := sbox.Run(ctx, p, sbox.Cmd{
		Path:   path,
		Args:   args,
		Stdout: buf,
		Stderr: buf,
	})
	if err != nil {
		return buf.String(), err
	}
	if res.ExitCode != 0 {
		return buf.String(), fmt.Errorf("exit code %d", res.ExitCode)
	}
	return buf.String(), nil
}

// verdict formuliert das Ergebnis des Sicherheitsbeweises als Klartext.
func verdict(out string, err error) string {
	if err != nil {
		return "✔ GESCHÜTZT: Der Leseversuch ist fehlgeschlagen – das Secret ist " +
			"aus der Sandbox nicht erreichbar."
	}
	if out == "" {
		return "✔ GESCHÜTZT: kein Inhalt gelesen."
	}
	return "✘ ACHTUNG: Es wurde Inhalt gelesen. Läuft dies auf einem " +
		"Nicht-Linux-Dev-Host? Dort nutzt sbox einen UNSANDBOXED-Passthrough " +
		"ohne Isolation – niemals in Produktion verwenden."
}

// platformHint weist darauf hin, wenn wir nicht auf Linux laufen und sbox daher
// nur den unsicheren Passthrough bereitstellt.
func platformHint() core.View {
	if isLinux() {
		return ui.Text("Läuft auf Linux – echte Sandbox aktiv.").
			Font(ui.Small).Color(ui.ColorSemanticGood)
	}
	return ui.Text("Achtung: Nicht-Linux-Dev-Host erkannt. sbox nutzt hier einen " +
		"UNSANDBOXED-Passthrough OHNE jede Isolation (mit Log-Warnung). Der " +
		"Sicherheitsbeweis unten schlägt daher fehl – so ist es gewollt. In " +
		"Produktion (Linux) ist die Sandbox aktiv.").
		Font(ui.Small).Color(ui.ColorSemanticWarn)
}

// ---- kleine Helfer ----------------------------------------------------------

func outputBox(s string) core.View {
	if s == "" {
		return ui.Text("— noch keine Ausgabe —").Font(ui.Small).Color(ui.ColorText)
	}
	return ui.CodeEditor(s).Disabled(true).Frame(ui.Frame{}.FullWidth())
}

func explainerCard(title, text string, body core.View) core.View {
	return ui.VStack(
		ui.Text(title).Font(ui.Title),
		ui.Text(text),
		ui.Space(ui.L8),
		body,
	).Alignment(ui.Leading).
		BackgroundColor(ui.ColorCardBody).
		Padding(ui.Padding{}.All(ui.L16)).
		Frame(ui.Frame{}.FullWidth())
}

func formatRun(cmd, out string, err error) string {
	b := "$ " + cmd + "\n"
	if out != "" {
		b += out
		if out[len(out)-1] != '\n' {
			b += "\n"
		}
	}
	if err != nil {
		b += "[Fehler: " + err.Error() + "]\n"
	}
	return b
}

func lookPath(name string) string {
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	// gängige absolute Fallbacks
	for _, c := range []string{"/usr/bin/" + name, "/bin/" + name, "/usr/local/go/bin/" + name} {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

// isLinux erlaubt eine plattformabhängige Anzeige ohne separate Build-Dateien.
func isLinux() bool { return runtime.GOOS == "linux" }
