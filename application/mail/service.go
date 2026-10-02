// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mail

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.wdy.de/nago/application/mail/nms"
)

// MailServiceID is the id of the Nago Mail Service as a transport. Use it as [Mail.SmtpHint] to send a mail by
// the service, even if SMTP servers are available.
const MailServiceID = "nms"

// MailServiceName is the server name of the Nago Mail Service in attempts, statistics and health.
const MailServiceName = "Nago Mail Service"

// MailService is an HTTP transport, which sends mails without an SMTP server, see [nms.Service].
type MailService interface {
	// Available is true, if the service can be used right now.
	Available() bool
	// Send sends the message. A sender outside the allowed domains is replaced, see [nms.Service.Send].
	Send(ctx context.Context, msg nms.Message) (nms.MessageResult, error)
	Limits() nms.Limits
	Status() nms.Status
}

// sendService delivers the mail by the mail service.
func (s *scheduler) sendService(ctx context.Context, id ID, m Mail) error {
	msg, err := toServiceMessage(id, m)
	if err != nil {
		return &SendError{Phase: PhaseConfig, Err: err}
	}

	res, err := s.opts.MailService.Send(ctx, msg)
	if err != nil {
		return serviceErr(err)
	}

	return messageErr(res)
}

// serviceErr maps a failed request to the phase, which the scheduler understands.
func serviceErr(err error) error {
	if errors.Is(err, nms.ErrNotEnrolled) || errors.Is(err, nms.ErrNoSender) {
		return &SendError{Phase: PhaseConfig, Err: err}
	}

	var e *nms.Error
	if !errors.As(err, &e) {
		return &SendError{Phase: PhaseDial, Err: err} // connection problems and timeouts
	}

	switch {
	case e.Status == http.StatusUnauthorized || e.Status == http.StatusForbidden:
		return &SendError{Phase: PhaseAuth, Code: e.Status, Err: err}
	case e.Status == http.StatusBadRequest || e.Status == http.StatusRequestEntityTooLarge:
		return &SendError{Phase: PhaseData, Code: 554, Err: err}
	default: // rate limits and unavailability
		return &SendError{Phase: PhaseDial, Code: e.Status, Err: err}
	}
}

// messageErr maps the result of a single message.
func messageErr(res nms.MessageResult) error {
	if res.Status == nms.StatusQueued {
		return nil
	}

	if len(res.Errors) == 0 {
		return &SendError{Phase: PhaseData, Err: fmt.Errorf("nago mail service: message has status '%s'", res.Status)}
	}

	first := res.Errors[0]
	err := fmt.Errorf("nago mail service: %s %s %s", first.Code, first.Field, first.Message)
	switch first.Code {
	case nms.CodeDuplicate:
		return nil // has already been sent
	case nms.CodeInvalidRecipient:
		return &SendError{Phase: PhaseRcpt, Code: 550, Err: err}
	case nms.CodeInvalidSender, nms.CodeSenderNotAllowed:
		return &SendError{Phase: PhaseMail, Code: 550, Err: err}
	case nms.CodeRateLimited:
		return &SendError{Phase: PhaseData, Code: 451, Err: err}
	default:
		return &SendError{Phase: PhaseData, Code: 554, Err: err}
	}
}

// toServiceMessage converts the mail. An invalid sender is dropped, so that the service derives one.
func toServiceMessage(id ID, m Mail) (nms.Message, error) {
	msg := nms.Message{Subject: m.Subject, CustomID: string(id)}
	for _, a := range m.To {
		msg.To = append(msg.To, nms.Address{Email: a.Address, Name: a.Name})
	}

	for _, a := range m.CC {
		msg.Cc = append(msg.Cc, nms.Address{Email: a.Address, Name: a.Name})
	}

	for _, a := range m.BCC {
		msg.Bcc = append(msg.Bcc, nms.Address{Email: a.Address, Name: a.Name})
	}

	msg.From = nms.Address{Name: m.From.Name}
	if adr, ok := validSenderAddress(m.From.Address); ok {
		msg.From.Email = adr
	}

	if m.InReplyTo != "" {
		msg.Headers = map[string]string{"In-Reply-To": m.InReplyTo}
	}

	var texts, htmls []string
	for i, p := range m.Parts {
		mediaType, params := partContentType(p)
		encoding := strings.ToLower(strings.TrimSpace(first(p.Header["Content-Transfer-Encoding"])))

		if encoding != "base64" && (mediaType == "text/plain" || mediaType == "text/html") {
			text := strings.TrimSuffix(strings.ReplaceAll(string(p.Encoded), "\r\n", "\n"), "\n")
			if mediaType == "text/html" {
				htmls = append(htmls, text)
			} else {
				texts = append(texts, text)
			}
			continue
		}

		content := base64.StdEncoding.EncodeToString(p.Encoded)
		if encoding == "base64" {
			raw, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(string(p.Encoded)), ""))
			if err != nil {
				return nms.Message{}, fmt.Errorf("part %d: invalid base64 content: %w", i, err)
			}
			content = base64.StdEncoding.EncodeToString(raw)
		}

		name := params["name"]
		if name == "" {
			name = fmt.Sprintf("attachment-%d", i+1)
		}

		if mediaType == "" {
			mediaType = "application/octet-stream"
		}

		msg.Attachments = append(msg.Attachments, nms.Attachment{Filename: name, ContentType: mediaType, Base64Content: content})
	}

	msg.TextPart = strings.Join(texts, "\n\n")
	msg.HTMLPart = strings.Join(htmls, "\n")
	return msg, nil
}

// partContentType parses the Content-Type values of a part, which are stored as separate values like
// "text/plain", "charset=utf-8", see [NewTextPart].
func partContentType(p Part) (string, map[string]string) {
	params := map[string]string{}
	mediaType := ""
	for _, value := range p.Header["Content-Type"] {
		for _, field := range strings.Split(value, ";") {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}

			if k, v, ok := strings.Cut(field, "="); ok {
				params[strings.ToLower(strings.TrimSpace(k))] = strings.Trim(strings.TrimSpace(v), `"`)
			} else if mediaType == "" {
				mediaType = strings.ToLower(field)
			}
		}
	}

	return mediaType, params
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}

	return values[0]
}
