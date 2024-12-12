package permission

var (
	PermFindAll = Declare[FindAll]("nago.permission.find_all", "Alle Berechtigungen anzeigen", "Träger dieser Berechtigung können alle definierten Berechtigungen anzeigen.")
)
