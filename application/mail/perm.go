package mail

import "go.wdy.de/nago/application/permission"

var (
	PermSendMail             = permission.Declare[SendMail]("nago.mail.send", "Mail Senden", "Träger dieser Berechtigung können Mails versenden.")
	PermInitDefaultTemplates = permission.Declare[SendMail]("nago.mail.init_default_templates", "Standard Templates setzen", "Träger dieser Berechtigung können die Standard Mail templates aktivieren.")
)
