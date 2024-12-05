package tpl

import "go.wdy.de/nago/auth"

type Model struct {
	Subject      auth.Subject
	Verification Verification
}

type Verification struct {
	Mail VerificationMail
}

type VerificationMail struct {
	ConfirmLink     string
	LifetimeMinutes int
	Code            string
}
