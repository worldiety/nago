package billing

import "go.wdy.de/nago/application/permission"

var (
	PermAppLicenses  = permission.Declare[AppLicenses]("nago.billing.license.app", "Abgerechnete App-Lizenzen anzeigen", "Träger dieser Berechtigung können alle verfügbaren und gebuchten Anwendungslizenzen anzeigen.")
	PermUserLicenses = permission.Declare[UserLicenses]("nago.billing.license.user", "Abgerechnete User-Lizenzen anzeigen", "Träger dieser Berechtigung können alle verfügbaren und gebuchten Nutzerlizenzen anzeigen.")
)
