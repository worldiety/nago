// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package consent

import (
	"time"
)

type ID string

// some predefined consent types.
const (
	// Newsletter consent allows contacting the user by sending advertising mails to the login mail.
	Newsletter ID = "nago.consent.newsletter"
	// SMS consent allows contacting the user by sending SMS to the mobile contact data.
	SMS ID = "nago.consent.sms"
	// GeneralTermsAndConditions must usually be acknowledged if a contract is closed between two parties.
	GeneralTermsAndConditions ID = "nago.consent.gtc"
	// TermsOfUse define the rules to apply to, under which a service can be used without acknowledging a full
	// contract.
	TermsOfUse ID = "nago.consent.termsofuse"
	// DataProtectionProvision is the GDPR provision, e.g. Datenschutzerklärung.
	DataProtectionProvision ID = "nago.consent.dprov"
	// MinAge must be conformed in some legal cases. See also Taschengeldparagraph.
	MinAge ID = "nago.consent.minage"
)

type Consent struct {
	ID      ID       `json:"id"`
	History []Action `json:"history,omitempty"`
}

func (c Consent) Identity() ID {
	return c.ID
}

func (c Consent) Status() Status {
	if len(c.History) == 0 {
		return Revoked
	}

	latest := c.History[0]
	for _, a := range c.History[1:] {
		if a.At.After(latest.At) {
			latest = a
		}
	}

	return latest.Status
}

func (c Consent) IsZero() bool {
	return c.ID == ""
}

type Status int

func (s Status) String() string {
	switch s {
	case Approved:
		return "approved"
	case Revoked:
		return "revoked"
	default:
		return "unknown"
	}
}

const (
	Revoked Status = iota
	Approved
)

type Action struct {
	At     time.Time `json:"at,omitempty"`
	Status Status    `json:"status,omitempty"`
}

func HasApproved(consents []Consent, id ID) bool {
	for _, consent := range consents {
		if consent.ID == id && consent.Status() == Approved {
			return true
		}
	}

	return false
}

func HasRevoked(consents []Consent, id ID) bool {
	return !HasApproved(consents, id)
}

func IsKnown(consents []Consent, id ID) bool {
	for _, consent := range consents {
		if consent.ID == id {
			return true
		}
	}

	return false
}
