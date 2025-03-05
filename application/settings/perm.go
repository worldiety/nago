package settings

import "go.wdy.de/nago/application/permission"

var (
	PermLoadGlobal  = permission.Declare[LoadGlobal]("nago.settings.global.load", "Globale Einstellungen anzeigen", "Träger dieser Berechtigung können alle globalen Einstellungen anzeigen.")
	PermStoreGlobal = permission.Declare[StoreGlobal]("nago.settings.global.store", "Globale Einstellungen speichern", "Träger dieser Berechtigung können alle globalen Einstellungen überschreiben.")
)
