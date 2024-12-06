package mail

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"iter"
	"net/mail"
	"time"
)

type ID string

// A Mail contains the specific high level parts of a mail
type Mail struct {
	To       []mail.Address
	CC       []mail.Address
	BCC      []mail.Address
	From     mail.Address
	Subject  string `label:"Betreff"`
	Parts    []Part
	SmtpHint SmtpID // an alternative Smtp server can be used e.g. for load balancing or different sender signatures
}

type Status string

const (
	StatusUndefined   Status = ""
	StatusQueued      Status = "queued"
	StatusSendSuccess Status = "send_success"
	StatusError       Status = "send_error"
)

type Outgoing struct {
	ID        ID `visible:"false"`
	Mail      Mail
	Subject   string `label:"Betreff" disabled:"true"`
	Receiver  string `label:"Empfänger" disabled:"true"`
	Status    Status `values:"[\"undefined=Undefiniert\",\"queued=wartet auf Versand\",\"send_success=erfolgreich versendet\",\"send_error=Versandfehler\"]"`
	LastError string `label:"Letzter Fehler"`
	Server    SmtpID `label:"Versendet über" disabled:"true" table-visible:"false"`
	QueuedAt  time.Time
	SendAt    time.Time
}

func (o Outgoing) WithIdentity(id ID) Outgoing {
	o.ID = id
	return o
}

func (o Outgoing) Identity() ID {
	return o.ID
}

type Repository data.Repository[Outgoing, ID]

type FindMailByID func(auth.Subject, ID) (std.Option[Outgoing], error)
type DeleteMailByID func(auth.Subject, ID) error
type FindAllMails func(auth.Subject) iter.Seq2[Outgoing, error]
type SaveMail func(auth.Subject, Outgoing) (ID, error)
