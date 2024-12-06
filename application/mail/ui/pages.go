package uimail

import "go.wdy.de/nago/presentation/core"

type MailPages struct {
	SMTPServer        core.NavigationPath
	OutgoingMailQueue core.NavigationPath
	MailScheduler     core.NavigationPath
	SendMailTest      core.NavigationPath
	Templates         core.NavigationPath
}
