// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mail

import "net/mail"

// SendMailRequested can be sent to the event bus and will be issued to the [SendMail] use case.
// This event is serializable and thus must be used with larger streamable parts or attachments.
// Any subject validation is circumvented and anyone can send mails using this request.
type SendMailRequested struct {
	To          []mail.Address    `json:"to,omitempty"`
	CC          []mail.Address    `json:"cc,omitempty"`
	BCC         []mail.Address    `json:"bcc,omitempty"`
	Subject     string            `json:"subject,omitempty"`
	TextBody    string            `json:"textBody,omitempty"`    // If not empty, send this as a text part.
	HTMLBody    string            `json:"HTMLBody,omitempty"`    // If not empty, send this as a html part.
	Attachments map[string][]byte `json:"attachments,omitempty"` // If not empty, add these binary data as Attachments.

	// SmtpHint allows to narrow the wanted mail server, e.g. for specific mail signatures. It can be empty.
	// The hint is matched against the [secret.SMTP.Name] or [secret.Secret.ID] of all secrets available to
	// [group.System].
	// If no match was found, the first found mail secret shared with [group.System] is used.
	SmtpHint string `json:"smtpHint,omitempty"`
}
