package secret

import "go.wdy.de/nago/application/permission"

var (
	PermFindMySecrets        = permission.Declare[FindMySecrets]("nago.secret.find_my_secrets", "Meine Secrets anzeigen", "Träger dieser Berechtigung können die Secrets anzeigen, die Sie besitzen.")
	PermCreateSecret         = permission.Declare[CreateSecret]("nago.secret.create", "Ein privates Secret erstellen", "Träger dieser Berechtigung können private Secrets erstellen.")
	PermUpdateMySecretGroups = permission.Declare[UpdateMySecretGroups]("nago.secret.groups.update", "Gruppen Secrets zuweisen", "Träger dieser Berechtigung können ihren Secrets Gruppen zuweisen, in denen sie ebenfalls Mitglied sind. Das Entfernen ist mit dieser Berechtigung immer möglich.")
	PermUpdateMySecretOwners = permission.Declare[UpdateMySecretOwners]("nago.secret.owners.update", "Besitzer Secrets zuweisen", "Träger dieser Berechtigung können ihren Secrets Besitzer zuweisen, bei denen sie ebenfalls Besitzer sind.")
	PermUpdateMyCredentials  = permission.Declare[UpdateMyCredentials]("nago.secret.credentials.update", "Meine Secret Credentials aktualisieren", "Träger dieser Berechtigung können die Credentials eines Secrets, der ihnen gehört, aktualisieren.")
	PermDeleteMySecretByID   = permission.Declare[DeleteMySecretByID]("nago.secret.delete", "Mein Secret löschen", "Träger dieser Berechtigung können die von ihr besessenen Secrets entfernen.")
)
