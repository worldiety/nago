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
	Subject string `label:"Betreff"`
	Parts   []Part
	// SmtpHint allows to narrow the wanted mail server, e.g. for specific mail signatures.
	// The hint is matched against the [secret.SMTP.Name] or [secret.Secret.ID] of all secrets available to
	// [group.System].
	// If no match was found, the first found mail secret shared with [group.System] is used.
	SmtpHint string
}

type Status string

const (
	StatusUndefined   Status = ""
	StatusQueued      Status = "queued"
	StatusSendSuccess Status = "send_success"
	StatusError       Status = "send_error"
)

type Outgoing struct {
	ID         ID `visible:"false"`
	Mail       Mail
	Subject    string `label:"Betreff" disabled:"true"`
	Receiver   string `label:"Empfänger" disabled:"true"`
	Status     Status `values:"[\"undefined=Undefiniert\",\"queued=wartet auf Versand\",\"send_success=erfolgreich versendet\",\"send_error=Versandfehler\"]"`
	LastError  string `label:"Letzter Fehler"`
	ServerName string `label:"Versendet über" disabled:"true" table-visible:"true"`
	QueuedAt   time.Time
	SendAt     time.Time
}

func (o Outgoing) WithIdentity(id ID) Outgoing {
	o.ID = id
	return o
}

func (o Outgoing) Identity() ID {
	return o.ID
}
