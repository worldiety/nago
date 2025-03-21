package template

import "go.wdy.de/nago/application/permission"

var (
	PermExecute                = permission.Declare[Execute]("nago.template.execute", "Template rendern", "Träger dieser Berechtigung können beliebige Templates rendern.")
	PermFindAll                = permission.Declare[FindAll]("nago.template.find_all", "Templates anzeigen", "Träger dieser Berechtigung können grundsätzlich für sie zugreifbare Templates anzeigen.")
	PermFindByID               = permission.Declare[FindByID]("nago.template.find_by_id", "Template anzeigen", "Träger dieser Berechtigung können ein grundsätzlich für sie zugreifbares Template anzeigen.")
	PermLoadProjectBlob        = permission.Declare[LoadProjectBlob]("nago.template.project.blob.load", "Projektdatei anzeigen", "Träger dieser Berechtigung können eine einzelne Datei aus einem Projekt anzeigen.")
	PermUpdateProjectBlob      = permission.Declare[UpdateProjectBlob]("nago.template.project.blob.update", "Projektdatei aktualisieren", "Träger dieser Berechtigung können eine einzelne Datei aus einem Projekt aktualisieren.")
	PermDeleteProjectBlob      = permission.Declare[DeleteProjectBlob]("nago.template.project.blob.delete", "Projektdatei löschen", "Träger dieser Berechtigung können eine einzelne Datei aus einem Projekt löschen.")
	PermRenameProjectBlob      = permission.Declare[RenameProjectBlob]("nago.template.project.blob.rename", "Projektdatei umbenennen", "Träger dieser Berechtigung können eine einzelne Datei aus einem Projekt umbenennen.")
	PermCreateProjectBlob      = permission.Declare[CreateProjectBlob]("nago.template.project.blob.create", "Projektdatei erstellen", "Träger dieser Berechtigung können eine einzelne Datei zu einem Projekt hinzufügen.")
	PermAddRunConfiguration    = permission.Declare[AddRunConfiguration]("nago.template.project.runcfg.add", "RunConfiguration aktualisieren", "Träger dieser Berechtigung können eine RunConfiguration hinzufügen.")
	PermRemoveRunConfiguration = permission.Declare[RemoveRunConfiguration]("nago.template.project.runcfg.remove", "RunConfiguration entfernen", "Träger dieser Berechtigung können eine RunConfiguration entfernen.")
	PermExportZip              = permission.Declare[ExportZip]("nago.template.project.export", "Projekt exportieren", "Träger dieser Berechtigung können ein Projekt als Zipdatei exportieren.")
	PermImportZip              = permission.Declare[ImportZip]("nago.template.project.import", "Projekt importieren", "Träger dieser Berechtigung können ein Projekt aus einer Zipdatei importieren.")
	PermCreate                 = permission.Declare[FindAll]("nago.template.create", "Template erstellen", "Träger dieser Berechtigung können neue Templates erstellen.")
	PermDelete                 = permission.Declare[Delete]("nago.template.delete", "Template löschen", "Träger dieser Berechtigung können Templates löschen.")
	PermEnsureBuildIn          = permission.Declare[FindAll]("nago.template.ensure_build_in", "Standard Template erstellen", "Träger dieser Berechtigung können neue Standard-Templates erstellen.")
)
