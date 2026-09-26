// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mail

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"go.wdy.de/nago/application/secret"
)

// SendError is returned by send and describes in which phase of the SMTP conversation the error occurred.
type SendError struct {
	Phase Phase
	Code  int // SMTP reply code or 0
	Err   error
}

func (e *SendError) Error() string {
	return fmt.Sprintf("%s: %v", e.Phase, e.Err)
}

func (e *SendError) Unwrap() error {
	return e.Err
}

func sendErr(phase Phase, err error) error {
	if err == nil {
		return nil
	}

	se := &SendError{Phase: phase, Err: err}
	var tpErr *textproto.Error
	if errors.As(err, &tpErr) {
		se.Code = tpErr.Code
	}

	return se
}

// validSenderAddress returns the trimmed address and true, if s is a single bare mail address like
// "info@example.com". Display names, lists or strings like API keys are rejected.
func validSenderAddress(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}

	adr, err := mail.ParseAddress(s)
	if err != nil || adr.Address != s {
		return "", false
	}

	return s, true
}

// resolveSender determines the envelope and header sender. The first valid mail address of the given from
// address, the configured sender address and the login username wins. A display name of from is kept.
// If none of them is a valid mail address, a [SendError] with [PhaseConfig] is returned, so that the mail
// is not sent without a proper sender and the failure becomes visible.
func resolveSender(credentials secret.SMTP, from mail.Address) (mail.Address, error) {
	for _, candidate := range []string{from.Address, credentials.SenderAddress, credentials.Username} {
		if adr, ok := validSenderAddress(candidate); ok {
			return mail.Address{Name: from.Name, Address: adr}, nil
		}
	}

	return mail.Address{}, &SendError{
		Phase: PhaseConfig,
		Err:   fmt.Errorf("no valid sender address: neither the mail sender, the sender address setting of smtp server '%s' nor its username is a valid mail address", credentials.Name),
	}
}

func send(credentials secret.SMTP, m Mail) (err error) {
	from, err := resolveSender(credentials, m.From)
	if err != nil {
		return err
	}
	m.From = from

	// Connect to the SMTP Server
	servername := credentials.Host + ":" + strconv.Itoa(credentials.Port)

	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", credentials.Username, credentials.Password, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	conn, err := net.DialTimeout("tcp", servername, 10*time.Second)
	if err != nil {
		return sendErr(PhaseDial, err)
	}

	defer conn.Close()

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return sendErr(PhaseDial, err)
	}

	defer c.Close()

	err = c.StartTLS(tlsconfig)
	if err != nil {
		return sendErr(PhaseTLS, err)
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		return sendErr(PhaseAuth, err)
	}
	// the from address is usually important for authentication
	if err = c.Mail(m.From.Address); err != nil {
		return sendErr(PhaseMail, err)
	}

	// add all recipients, which is independent of what is in the actual message
	for _, adr := range m.To {
		if err = c.Rcpt(adr.Address); err != nil {
			return sendErr(PhaseRcpt, err)
		}
	}

	for _, adr := range m.CC {
		if err = c.Rcpt(adr.Address); err != nil {
			return sendErr(PhaseRcpt, err)
		}
	}

	for _, adr := range m.BCC {
		if err = c.Rcpt(adr.Address); err != nil {
			return sendErr(PhaseRcpt, err)
		}
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return sendErr(PhaseData, err)
	}

	to := recipients(m.To).String()
	if len(to) == 0 {
		return sendErr(PhaseRcpt, fmt.Errorf("recipient list is empty"))
	}

	// Setup data
	boundaryMultipartMixed := "------------DD1AADA7899159F3F80A4C5A"
	data := &dataWriter{sb: &bytes.Buffer{}}
	data.writeHeader("From", m.From.String())
	data.writeHeader("To", to)
	data.writeHeader("CC", recipients(m.CC).String())
	data.writeHeader("Subject", mime.QEncoding.Encode("UTF-8", m.Subject))
	data.writeHeader("MIME-Version", "1.0")
	data.writeHeader("Content-Type", "multipart/mixed;  boundary=\""+boundaryMultipartMixed+"\"")

	// Threading header: when InReplyTo is set, write In-Reply-To only. References and Message-ID are
	// intentionally omitted, see [Mail.InReplyTo].
	if len(m.InReplyTo) > 0 {
		data.writeHeader("In-Reply-To", m.InReplyTo)
	}

	data.rf()
	data.writeLine(" This is a multi-Part message in MIME format.")
	data.rf()
	data.rf()

	for _, p := range m.Parts {
		data.writeLine("--")
		data.writeLine(boundaryMultipartMixed)
		data.rf()
		err = p.write(data.sb)
		if err != nil {
			return sendErr(PhaseData, err)
		}

		data.rf()
	}

	data.writeLine("--")
	data.writeLine(boundaryMultipartMixed)
	data.writeLine("--")

	//fmt.Println(string(data.sb.Bytes()))
	_, err = w.Write(data.sb.Bytes())
	if err != nil {
		return sendErr(PhaseData, err)
	}

	err = w.Close()
	if err != nil {
		return sendErr(PhaseData, err)
	}

	return sendErr(PhaseQuit, c.Quit())
}

type recipients []mail.Address

func (r recipients) String() string {
	if len(r) == 0 {
		return ""
	}
	sb := &strings.Builder{}
	for i := 0; i < len(r)-1; i++ {
		sb.WriteString(r[i].String())
		sb.WriteString(",")
	}
	sb.WriteString(r[len(r)-1].String())
	return sb.String()
}

type dataWriter struct {
	sb *bytes.Buffer
}

func (d *dataWriter) writeHeader(key string, value string) *dataWriter {
	d.sb.WriteString(fmt.Sprintf("%s: %s\r\n", protect(key), protect(value)))
	return d
}

func (d *dataWriter) rf() *dataWriter {
	d.sb.WriteString("\r\n")
	return d
}

func (d *dataWriter) writeLine(str string) *dataWriter {
	d.sb.WriteString(str)
	return d
}
