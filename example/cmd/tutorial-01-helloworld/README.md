## Das obligatorische _hello world_ Beispiel.

Sämtliche Tutorialbeispiele befinden sich auch als ausführbare Packages im [Nago-Projekt](https://gitlab.worldiety.net/group/ora/nago/-/tree/main/example/cmd?ref_type=heads). Um ein Tutorial Paket auszuführen, reicht ein Aufruf wie `go run go.wdy.de/nago/example/cmd/tutorial-helloworld@latest`.


Da es sich beim Nago-Projekt um ein firmeninternes Repository handelt, kann es nicht automatisch durch das Go-Modulsystem aufgelöst werden. Einmalig musst du also deine Konfiguration anpassen:

1. Du musst deinen public ssh-key in deinem worldiety Gitlab Account hinzugefügt haben
2. Konfiguriere die folgende git-replace Regel:
```bash 
git config --global url."ssh://git@gitlab.worldiety.net/".insteadOf "https://gitlab.worldiety.net/" 
```
3. Nun muss das _go buildsystem_ noch wissen, dass es sich um ein privates Repository handelt und damit die öffentliche _notary sum database_ deaktiviert wird:
```bash
# note the \* escaping for zsh
go env -w GOPRIVATE=go.wdy.de/\*,gitlab.worldiety.net/\* 
```
4. Vergiss das initale `go mod tidy` bei deinem eigenen Projekt nicht, damit sich dein lokaler Modulecache die Abhängigkeiten zieht. Für ein `go run` ist das aber nicht erforderlich.


Somit sollten sich nun alle Beispiele bauen und ausführen lassen.
Beim späteren Bauen in der CI/CD-Pipeline musst du diese Konfiguration in deiner `.gitlab-ci.yaml` entsprechend nachvollziehen.
Alternativ kannst du sämtliche Abhängigkeiten auch mittels `go mod vendor` in deinem Projekt hinzufügen und kannst fortan offline reproduzierbare Builds erzeugen.


Eine kommentierte _hello world_ Version
```go
// main denotes an executable go package. If you don't know, what that means, go through the Go Tour first.
package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

// the main function of the program, which is like the java public static void main.
func main() {
	// we use the applications package to bootstrap our configuration
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {
			return ui.VStack(ui.Text("hello world")).
				Frame(ora.Frame{}.MatchScreen())
		})
	}).
		// don't forget to call the run method, which starts the entire thing and blocks until finished
		Run()
}

```


Minified _hello world_ zum Copy-Pasten:
```go
package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			return ui.VStack(ui.Text("hello world")).
				Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}

```