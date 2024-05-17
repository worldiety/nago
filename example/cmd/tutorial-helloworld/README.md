# hello world

Das obligatorische _hello world_ Beispiel.

Sämtliche Tutorialbeispiele befinden sich auch als ausführbare Packages im Nago-Projekt. Um ein Tutorial Paket auszuführen, reicht ein Aufruf wie `go run go.wdy.de/nago/example/cmd/tutorial-helloworld@latest`.


Da es sich beim Nago-Projekt um ein firmeninternes Repository handelt, kann es nicht automatisch durch das Go-Modulsystem aufgelöst werden. Einmalig musst du also deine Konfiguration anpassen:

1. Du musst deinen public ssh-key in deinem worldiety Gitlab Account hinzugefügt haben
2. Konfiguriere die folgende git-replace Regel:
```bash 
git config --global url."ssh://git@gitlab.worldiety.net/".insteadOf "https://gitlab.worldiety.net/" 
```
3. Nun muss das _go buildsystem_ noch wissen, dass es sich um ein privates Repository handelt und damit die öffentliche _notary sum database_ deaktiviert wird:
```bash
# note the \* escaping for zsh go env -w GOPRIVATE=go.wdy.de/\*,gitlab.worldiety.net/\* 
```
4. Vergiss das initale `go mod tidy` bei deinem eigenen Projekt nicht, damit sich dein lokaler Modulecache die Abhängigkeiten zieht. Für ein `go run` ist das aber nicht erforderlich.


Somit sollten sich nun alle Beispiele bauen und ausführen lassen.
Beim späteren Bauen in der CI/CD-Pipeline musst du diese Konfiguration in deiner `.gitlab-ci.yaml` entsprechend nachvollziehen.
Alternativ kannst du sämtliche Abhängigkeiten auch mittels `go mod vendor` in deinem Projekt hinzufügen und kannst fortan offline reproduzierbare Builds erzeugen.