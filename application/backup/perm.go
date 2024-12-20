package backup

import "go.wdy.de/nago/application/permission"

var (
	PermBackup  = permission.Declare[Backup]("nago.backup.backup", "Backup erstellen", "Träger dieser Berechtigung können die gesamten Daten als Backup herunterladen. Dies hat einen Datenabfluss zur Folge. In der Regel sollte dies nur der Systemadministrator aus Datenschutz- und Vertraulichkeitsgründen dürfen.")
	PermRestore = permission.Declare[Restore]("nago.backup.restore", "Backup wiederherstellen", "Träger dieser Berechtigung können die gesamten Daten aus einem Backup ersetzen. Dies hat einen Datenverlust und potentielle Sicherheitsprobleme zur Folge. In der Regel sollte dies nur der Systemadministrator aus Datenschutz- und Vertraulichkeitsgründen dürfen.")
)
