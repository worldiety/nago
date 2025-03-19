package template

import "go.wdy.de/nago/application/permission"

var (
	PermExecute                = permission.Declare[Execute]("nago.template.execute", "Template rendern", "Träger dieser Berechtigung können beliebige Templates rendern.")
	PermFindAll                = permission.Declare[FindAll]("nago.template.find_all", "Templates anzeigen", "Träger dieser Berechtigung können grundsätzlich für sie zugreifbare Templates anzeigen.")
	PermFindByID               = permission.Declare[FindByID]("nago.template.find_by_id", "Template anzeigen", "Träger dieser Berechtigung können ein grundsätzlich für sie zugreifbares Template anzeigen.")
	PermLoadProjectBlob        = permission.Declare[LoadProjectBlob]("nago.template.project.blob.load", "Projektdatei anzeigen", "Träger dieser Berechtigung können eine einzelne Datei aus einem Projekt anzeigen.")
	PermUpdateProjectBlob      = permission.Declare[UpdateProjectBlob]("nago.template.project.blob.update", "Projektdatei aktualisieren", "Träger dieser Berechtigung können eine einzelne Datei aus einem Projekt aktualisieren.")
	PermDeleteProjectBlob      = permission.Declare[DeleteProjectBlob]("nago.template.project.blob.delete", "Projektdatei löschen", "Träger dieser Berechtigung können eine einzelne Datei aus einem Projekt löschen.")
	PermAddRunConfiguration    = permission.Declare[AddRunConfiguration]("nago.template.project.runcfg.add", "RunConfiguration aktualisieren", "Träger dieser Berechtigung können eine RunConfiguration hinzufügen.")
	PermRemoveRunConfiguration = permission.Declare[RemoveRunConfiguration]("nago.template.project.runcfg.remove", "RunConfiguration entfernen", "Träger dieser Berechtigung können eine RunConfiguration entfernen.")
	PermCreate                 = permission.Declare[FindAll]("nago.template.create", "Template erstellen", "Träger dieser Berechtigung können neue Templates erstellen.")
	PermEnsureBuildIn          = permission.Declare[FindAll]("nago.template.ensure_build_in", "Standard Template erstellen", "Träger dieser Berechtigung können neue Standard-Templates erstellen.")
)
