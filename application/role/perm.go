package role

import "go.wdy.de/nago/application/permission"

var (
	PermFindByID = permission.Declare[FindByID]("nago.role.find_by_id", "Eine Rolle anzeigen", "Träger dieser Berechtigung können eine Rolle über ihre EID anzeigen.")
	PermFindAll  = permission.Declare[FindAll]("nago.role.find_all", "Alle Rollen anzeigen", "Träger dieser Berechtigung können alle vorhandenen Rollen anzeigen.")
	PermCreate   = permission.Declare[Create]("nago.role.create", "Rollen erstellen", "Träger dieser Berechtigung können neue Rollen anlegen.")
	PermUpdate   = permission.Declare[Update]("nago.role.update", "Rollen aktualisieren", "Träger dieser Berechtigung können vorhandene Rollen aktualisieren.")
	PermDelete   = permission.Declare[Delete]("nago.role.delete", "Rollen löschen", "Träger dieser Berechtigung können vorhandene Rollen löschen.")
)
