package billing

import "go.wdy.de/nago/application/permission"

var (
	PermAppLicenses = permission.Declare[AppLicenses]("nago.billing.license.app", "Abgerechnete Lizenzen anzeigen", "Träger dieser Berechtigung können alle verfügbaren und gebuchten Anwendungslizenzen anzeigen.")
)
