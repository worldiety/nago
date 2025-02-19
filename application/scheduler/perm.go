package scheduler

import "go.wdy.de/nago/application/permission"

var (
	PermConfigure          = permission.Declare[Configure]("nago.scheduler.configure", "Scheduler erstellen", "Träger dieser Berechtigung können beliebige Scheduler hinzufügen.")
	PermViewLogs           = permission.Declare[ViewLogs]("nago.scheduler.viewlogs", "Scheduler Logs einsehen", "Träger dieser Berechtigung können beliebige Scheduler Logs betrachten und ggf. dadurch sensitive Informationen auslesen.")
	PermStatus             = permission.Declare[Status]("nago.scheduler.status", "Scheduler Status auslesen", "Träger dieser Berechtigung können verschiedenen Scheduler Status Informationen auslesen.")
	PermExecuteNow         = permission.Declare[ExecuteNow]("nago.scheduler.executenow", "Scheduler direkt ausführen", "Träger dieser Berechtigung können den Scheduler Job manuell ausführen.")
	PermListSchedulers     = permission.Declare[ListSchedulers]("nago.scheduler.listall", "Scheduler auflisten", "Träger dieser Berechtigung können alle Scheduler Jobs auflisten.")
	PermStart              = permission.Declare[Start]("nago.scheduler.start", "Scheduler starten", "Träger dieser Berechtigung können einen Scheduler per ID starten bzw. händisch ausführen.")
	PermStop               = permission.Declare[Stop]("nago.scheduler.stop", "Scheduler beenden", "Träger dieser Berechtigung können einen Scheduler per ID bis zum nächsten Systemstart beenden.")
	PermUpdateSettingsByID = permission.Declare[UpdateSettings]("nago.scheduler.settings_update", "Scheduler Settings aktualisieren", "Träger dieser Berechtigung können die Scheduler Settings per ID ändern.")
	PermDeleteSettingsByID = permission.Declare[DeleteSettingsByID]("nago.scheduler.settings.delete_by_id", "Scheduler Settings löschen", "Träger dieser Berechtigung können die Scheduler Settings löschen und zurücksetzen.")
	PermFindSettingsByID   = permission.Declare[FindSettingsByID]("nago.scheduler.settings.find_by_id", "Scheduler Settings anzeigen", "Träger dieser Berechtigung können die Scheduler Setting per ID anzeigen.")
)
