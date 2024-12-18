package license

import (
	"go.wdy.de/nago/application/permission"
)

var (
	PermFindAllAppLicenses  = permission.Declare[FindAllAppLicenses]("nago.license.app.find_all", "App-Lizenzen anzeigen", "Träger dieser Berechtigung können alle verfügbaren Anwendungslizenzen anzeigen.")
	PermFindAppLicenseByID  = permission.Declare[FindAppLicenseByID]("nago.license.app._find_by_id", "App-Lizenz anzeigen", "Träger dieser Berechtigung können eine Lizenz über ihre ID anzeigen.")
	PermCreateAppLicense    = permission.Declare[CreateAppLicense]("nago.license.app.create", "App-Lizenz erstellen", "Träger dieser Berechtigung können eine neue App-Lizenz erstellen.")
	PermUpdateAppLicense    = permission.Declare[UpdateAppLicense]("nago.license.app.update", "App-Lizenz aktualisieren", "Träger dieser Berechtigung können eine App-Lizenz aktualisieren.")
	PermDeleteAppLicense    = permission.Declare[DeleteAppLicense]("nago.license.app.delete", "App-Lizenz löschen", "Träger dieser Berechtigung können eine App-Lizenz löschen.")
	PermFindAllUserLicenses = permission.Declare[FindAllUserLicenses]("nago.license.user.find_all", "Nutzer-Lizenzen anzeigen", "Träger dieser Berechtigung können alle verfügbaren Nutzerlizenzen anzeigen.")
	PermFindUserLicenseByID = permission.Declare[FindUserLicenseByID]("nago.license.user.find_by_id", "Nutzer-Lizenz anzeigen", "Träger dieser Berechtigung können eine Nutzerlizenz anzeigen.")
	PermCreateUserLicense   = permission.Declare[CreateUserLicense]("nago.license.user.create", "Nutzer-Lizenz erstellen", "Träger dieser Berechtigung können eine neue Nutzerlizenz erstellen.")
	PermUpdateUserLicense   = permission.Declare[UpdateUserLicense]("nago.license.user.update", "Nutzer-Lizenz aktualisieren", "Träger dieser Berechtigung können eine vorhandene Nutzerlizenz aktualisieren.")
	PermDeleteUserLicense   = permission.Declare[DeleteUserLicense]("nago.license.user.delete", "Nutzer-Lizenz löschen", "Träger dieser Berechtigung können eine Nutzerlizenz löschen.")
)
