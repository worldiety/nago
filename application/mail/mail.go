// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mail

import (
	"net/mail"
	"time"
)

type ID string

// A Mail contains the specific high level parts of a mail
type Mail struct {
	To      []mail.Address
	CC      []mail.Address
	BCC     []mail.Address
	From    mail.Address
	Subject string
	Parts   []Part
	// SmtpHint allows to narrow the wanted mail server, e.g. for specific mail signatures.
	// The hint is matched against the [secret.SMTP.Name] or [secret.Secret.ID] of all secrets available to
	// [group.System].
	// If no match was found, the first found mail secret shared with [group.System] is used.
	SmtpHint string

	// InReplyTo sets the RFC 5322 In-Reply-To header and acts as a thread anchor. If empty, no threading
	// header is written at all and the resulting mail is byte-identical to the behavior without this field.
	// References and Message-ID are intentionally not set, because a constant subject together with a
	// constant In-Reply-To value is sufficient for mail clients to group messages into a conversation.
	InReplyTo string
}

type Status string

const (
	StatusUndefined   Status = ""
	StatusQueued      Status = "queued"
	StatusSendSuccess Status = "send_success"
	StatusError       Status = "send_error"

	// StatusFailed denotes, that the scheduler gave up after [ScheduleOptions.MaxAttempts]. Such mails are only
	// retried if explicitly requested, see [RetryOutgoing]. Older framework versions do not know this state
	// and will just retry it like [StatusError].
	StatusFailed Status = "send_failed"
)

// Phase describes the step of the SMTP conversation, in which an [Attempt] failed.
type Phase string

const (
	PhaseUndefined Phase = ""
	PhaseConfig    Phase = "config" // e.g. no smtp server available
	PhaseDial      Phase = "dial"
	PhaseTLS       Phase = "tls"
	PhaseAuth      Phase = "auth"
	PhaseMail      Phase = "mail"
	PhaseRcpt      Phase = "rcpt"
	PhaseData      Phase = "data"
	PhaseQuit      Phase = "quit"
)

// Attempt documents a single send try of the scheduler.
type Attempt struct {
	At       time.Time     `json:"at"`
	Server   string        `json:"server,omitempty"`
	Success  bool          `json:"success,omitempty"`
	Phase    Phase         `json:"phase,omitempty"`
	Code     int           `json:"code,omitempty"` // SMTP reply code, if available
	Message  string        `json:"message,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
}

// MaxAttemptHistory limits the amount of [Attempt] entries kept per [Outgoing] mail.
const MaxAttemptHistory = 20

// Outgoing is a mail within the outgoing queue. Note, that the persisted format must stay backwards compatible,
// thus only add optional fields.
type Outgoing struct {
	ID         ID
	Mail       Mail
	Subject    string
	Receiver   string
	Status     Status
	LastError  string
	ServerName string
	QueuedAt   time.Time
	SendAt     time.Time // time of the last send attempt

	// Attempts contains the latest send attempts, the oldest first. Mails queued by older versions have no attempts.
	Attempts []Attempt `json:",omitempty"`
	// AttemptCount is the total amount of attempts, which may be larger than len(Attempts).
	AttemptCount int `json:",omitempty"`
	// SentAt is the time of the successful delivery to the SMTP server.
	SentAt time.Time `json:",omitzero"`
	// NextAttemptAt is the earliest time for the next attempt. Zero means as soon as possible.
	NextAttemptAt time.Time `json:",omitzero"`
}

// LastAttempt returns the latest attempt, if any.
func (o Outgoing) LastAttempt() (Attempt, bool) {
	if len(o.Attempts) == 0 {
		return Attempt{}, false
	}

	return o.Attempts[len(o.Attempts)-1], true
}

// Attempted returns the amount of send attempts, also considering mails from older versions.
func (o Outgoing) Attempted() int {
	if o.AttemptCount > 0 {
		return o.AttemptCount
	}

	if !o.SendAt.IsZero() {
		return 1
	}

	return 0
}

func (o *Outgoing) addAttempt(a Attempt) {
	o.Attempts = append(o.Attempts, a)
	if len(o.Attempts) > MaxAttemptHistory {
		o.Attempts = o.Attempts[len(o.Attempts)-MaxAttemptHistory:]
	}

	if o.AttemptCount == 0 && !o.SendAt.IsZero() && len(o.Attempts) == 1 {
		o.AttemptCount = 1 // legacy attempt without history
	}

	o.AttemptCount++
	o.SendAt = a.At
	o.ServerName = a.Server
}

func (o Outgoing) WithIdentity(id ID) Outgoing {
	o.ID = id
	return o
}

func (o Outgoing) Identity() ID {
	return o.ID
}
