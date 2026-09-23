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

func send(credentials secret.SMTP, m Mail) (err error) {
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
	if len(m.From.Address) == 0 {
		if len(credentials.SenderAddress) != 0 {
			m.From.Address = credentials.SenderAddress
		} else {
			m.From.Address = credentials.Username
		}
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
