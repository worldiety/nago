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
	"net/mail"
	"testing"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/mail/nms"
	"go.wdy.de/nago/application/mail/nms/nmstest"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
)

func newServiceScheduler(t *testing.T, origin string, servers ...secret.SMTP) (*scheduler, Repository, *nmstest.Server, *sentLog) {
	t.Helper()
	srv := nmstest.NewServer(t)
	svc := nms.NewService(nms.Options{Endpoint: srv.URL, Origin: func() string { return origin }, Nonces: nms.NewNonces()})

	repo := json.NewSloppyJSONRepository[Outgoing, ID](mem.NewBlobStore("outgoing"))
	s := newScheduler(ScheduleOptions{MailService: svc}, repo, func() user.Subject { return user.SU() }, testSecrets(servers...))
	log := &sentLog{}
	s.send = func(credentials secret.SMTP, m Mail) error {
		log.add(m)
		return nil
	}
	s.sleep = func(time.Duration) {}
	return s, repo, srv, log
}

func TestSchedulerUsesMailServiceWithoutSmtp(t *testing.T) {
	s, repo, srv, _ := newServiceScheduler(t, "https://wokoda.apps.example.com")
	o := queue(t, repo, "hello", "torben@example.com", time.Now())
	o.Mail.From = mail.Address{Name: "Wokoda", Address: "info@wokoda.de"}
	if err := repo.Save(o); err != nil {
		t.Fatal(err)
	}

	s.runOnce(context.Background())

	sent := srv.Sent()
	if len(sent) != 1 {
		t.Fatalf("expected one message by the service, got %d", len(sent))
	}

	msg := sent[0]
	if msg.From.Email != "noreply@wokoda.apps.example.com" || msg.From.Name != "Wokoda" || msg.ReplyTo == nil || msg.ReplyTo.Email != "info@wokoda.de" {
		t.Fatalf("the sender has not been replaced: from=%+v replyTo=%+v", msg.From, msg.ReplyTo)
	}

	if msg.CustomID != string(o.ID) || msg.TextPart != "body hello" || msg.To[0].Email != "torben@example.com" {
		t.Fatalf("unexpected message %+v", msg)
	}

	got := option.Must(repo.FindByID(o.ID)).Unwrap()
	if got.Status != StatusSendSuccess {
		t.Fatalf("expected success, got %s %s", got.Status, got.LastError)
	}

	if a, _ := got.LastAttempt(); a.Server != MailServiceName {
		t.Fatalf("the attempt names %q", a.Server)
	}

	if st := currentSchedulerStatus(); st.NoSmtpServer {
		t.Fatal("the mail service counts as a server")
	}
}

func TestSchedulerPrefersSmtp(t *testing.T) {
	s, repo, srv, log := newServiceScheduler(t, "https://wokoda.apps.example.com", secret.SMTP{Name: "a", Host: "localhost", Port: 25})
	queue(t, repo, "by smtp", "torben@example.com", time.Now())
	hinted := queue(t, repo, "by service", "torben@example.com", time.Now().Add(time.Second))
	hinted.Mail.SmtpHint = MailServiceID
	if err := repo.Save(hinted); err != nil {
		t.Fatal(err)
	}

	s.runOnce(context.Background())

	if got := log.subjects(); len(got) != 1 || got[0] != "by smtp" {
		t.Fatalf("expected only the unhinted mail by smtp, got %v", got)
	}

	if sent := srv.Sent(); len(sent) != 1 || sent[0].Subject != "by service" {
		t.Fatalf("expected the hinted mail by the service, got %+v", sent)
	}
}

// A local instance asks the service once, because the service may know it by its address. If the service refuses,
// it is not asked again before the exchange interval passed, and the mail waits for a usable transport.
func TestSchedulerWithLocalInstanceTheServiceRefuses(t *testing.T) {
	s, repo, srv, _ := newServiceScheduler(t, "http://localhost:3000")
	srv.VerifyOrigin = func(origin, nonce string) bool { return false }
	o := queue(t, repo, "hello", "torben@example.com", time.Now())

	s.runOnce(context.Background())
	s.runOnce(context.Background())

	if exchanges, _, sends := srv.Calls(); exchanges != 1 || sends != 0 {
		t.Fatalf("expected exactly one exchange and no send, got %d exchanges and %d sends", exchanges, sends)
	}

	if s.opts.MailService.Available() {
		t.Fatal("after a refused exchange, the service is not usable until the exchange interval passed")
	}

	got := option.Must(repo.FindByID(o.ID)).Unwrap()
	if got.Status == StatusFailed || got.Attempted() > 1 {
		t.Fatalf("the mail must wait, got %s after %d attempts", got.Status, got.Attempted())
	}
}

// A local instance the service knows by its address sends like any other.
func TestSchedulerWithLocalInstanceTheServiceAdmits(t *testing.T) {
	s, repo, srv, _ := newServiceScheduler(t, "http://localhost:3000")
	queue(t, repo, "hello", "torben@example.com", time.Now())

	s.runOnce(context.Background())

	if sent := srv.Sent(); len(sent) != 1 {
		t.Fatalf("expected one message by the service, got %d", len(sent))
	}
}

func TestSchedulerServiceRejectsRecipientPermanently(t *testing.T) {
	s, repo, _, _ := newServiceScheduler(t, "https://wokoda.apps.example.com")
	o := queue(t, repo, "hello", "not-an-address", time.Now())

	s.runOnce(context.Background())

	got := option.Must(repo.FindByID(o.ID)).Unwrap()
	if got.Status != StatusFailed {
		t.Fatalf("expected a permanent failure, got %s %s", got.Status, got.LastError)
	}
}

func TestToServiceMessage(t *testing.T) {
	data := []byte("%PDF-1.7 binary \x00\x01\x02")
	m := Mail{
		From:      mail.Address{Address: "noreply@app.example.com"},
		To:        []mail.Address{{Address: "a@example.com", Name: "A"}},
		CC:        []mail.Address{{Address: "c@example.com"}},
		BCC:       []mail.Address{{Address: "b@example.com"}},
		Subject:   "subject",
		InReplyTo: "<thread@app.example.com>",
		Parts:     []Part{NewTextPart("line 1\nline 2"), NewHtmlPart("<p>hello</p>"), NewAttachmentPart("rechnung.pdf", data)},
	}

	msg, err := toServiceMessage("id1", m)
	if err != nil {
		t.Fatal(err)
	}

	if msg.From.Email != "noreply@app.example.com" || msg.ReplyTo != nil {
		t.Fatalf("the sender must be kept: %+v %+v", msg.From, msg.ReplyTo)
	}

	if msg.TextPart != "line 1\nline 2" || msg.HTMLPart != "<p>hello</p>" {
		t.Fatalf("unexpected bodies %q %q", msg.TextPart, msg.HTMLPart)
	}

	if len(msg.Cc) != 1 || len(msg.Bcc) != 1 || msg.Headers["In-Reply-To"] != "<thread@app.example.com>" || msg.CustomID != "id1" {
		t.Fatalf("unexpected message %+v", msg)
	}

	if len(msg.Attachments) != 1 || msg.Attachments[0].Filename != "rechnung.pdf" || msg.Attachments[0].ContentType != "application/octet-stream" {
		t.Fatalf("unexpected attachments %+v", msg.Attachments)
	}

	raw, err := base64.StdEncoding.DecodeString(msg.Attachments[0].Base64Content)
	if err != nil || string(raw) != string(data) {
		t.Fatalf("the attachment has not survived: %q %v", raw, err)
	}

	if msg, _ := toServiceMessage("id2", Mail{From: mail.Address{Name: "App", Address: "API-KEY-123"}, Parts: m.Parts}); msg.From.Email != "" || msg.From.Name != "App" {
		t.Fatalf("an invalid sender must be dropped: %+v", msg.From)
	}
}
