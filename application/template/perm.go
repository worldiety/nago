package template

import "go.wdy.de/nago/application/permission"

var (
	PermExecute       = permission.Declare[Execute]("nago.template.execute", "Template rendern", "Träger dieser Berechtigung können beliebige Template rendern.")
	PermFindAll       = permission.Declare[FindAll]("nago.template.find_all", "Templates anzeigen", "Träger dieser Berechtigung können grundsätzlich für sie zugreifbare Templates anzeigen.")
	PermCreate        = permission.Declare[FindAll]("nago.template.create", "Template erstellen", "Träger dieser Berechtigung können neue Templates erstellen.")
	PermEnsureBuildIn = permission.Declare[FindAll]("nago.template.ensure_build_in", "Standard Template erstellen", "Träger dieser Berechtigung können neue Standard-Templates erstellen.")
)
