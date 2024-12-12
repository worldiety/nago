package tpl

import (
	"go.wdy.de/nago/application/user"
)

type Model struct {
	Subject      user.Subject
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
