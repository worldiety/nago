package group

import "go.wdy.de/nago/application/permission"

var (
	PermFindByID = permission.Declare[FindByID]("nago.group.find_by_id", "Eine Gruppe anzeigen", "Träger dieser Berechtigung können eine Gruppe über ihre EID anzeigen.")
	PermFindAll  = permission.Declare[FindAll]("nago.group.find_all", "Alle Gruppen anzeigen", "Träger dieser Berechtigung können alle vorhandenen Gruppen anzeigen.")
	PermCreate   = permission.Declare[Create]("nago.group.create", "Gruppen erstellen", "Träger dieser Berechtigung können neue Gruppen anlegen.")
	PermUpdate   = permission.Declare[Update]("nago.group.update", "Gruppen aktualisieren", "Träger dieser Berechtigung können vorhandene Gruppen aktualisieren.")
	PermDelete   = permission.Declare[Delete]("nago.group.delete", "Gruppen löschen", "Träger dieser Berechtigung können vorhandene Gruppen löschen.")
)
