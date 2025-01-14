package user

import (
	"golang.org/x/text/language"
	"time"
)

type Created struct {
	ID                ID
	Firstname         string
	Lastname          string
	Email             Email
	PreferredLanguage language.Tag
	NotifyUser        bool
}

type MFACodeCreated struct {
	ID                ID
	Firstname         string
	Lastname          string
	Email             Email
	PreferredLanguage language.Tag
	ValidUntil        time.Time
	Code              string
}
