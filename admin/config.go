package admin

import "go.wdy.de/nago/presentation/core"

type Pages struct {
	Dashboard         core.NavigationPath
	SMTPServer        core.NavigationPath
	OutgoingMailQueue core.NavigationPath
	MailScheduler     core.NavigationPath
	SendMailTest      core.NavigationPath
}

func (p Pages) SendMailTestOrDefault() core.NavigationPath {
	if p.SendMailTest == "" {
		p.SendMailTest = "admin/mail/test"
	}

	return p.SendMailTest
}

func (p Pages) DashboardOrDefault() core.NavigationPath {
	if p.Dashboard == "" {
		p.Dashboard = "admin/dashboard"
	}
	return p.Dashboard
}

func (p Pages) SMTPServerOrDefault() core.NavigationPath {
	if p.SMTPServer == "" {
		return "admin/mail/smtp"
	}

	return p.SMTPServer
}

func (p Pages) OutgoingMailQueueOrDefault() core.NavigationPath {
	if p.OutgoingMailQueue == "" {
		p.OutgoingMailQueue = "admin/mail/outgoing"
	}

	return p.OutgoingMailQueue
}

func (p Pages) MailSchedulerOrDefault() core.NavigationPath {
	if p.MailScheduler == "" {
		p.MailScheduler = "admin/mail/scheduler"
	}

	return p.MailScheduler
}
