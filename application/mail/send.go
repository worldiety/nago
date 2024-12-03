package mail

import (
	"go.wdy.de/nago/annotation"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"time"
)

var PermissionSend = annotation.Permission[SendMail]("de.worldiety.nago.mail.send")

// SendMail takes the Mail and will try to publish it into either the given [Smtp] hint or whatever is currently defined
// as primary.
type SendMail func(subject auth.Subject, mail Mail) (ID, error)

func NewSendMail(mails Repository) SendMail {
	return func(subject auth.Subject, mail Mail) (ID, error) {
		if err := subject.Audit(PermissionSend.Identity()); err != nil {
			return "", err
		}

		out := Outgoing{
			ID:       data.RandIdent[ID](),
			Mail:     mail,
			Status:   StatusQueued,
			QueuedAt: time.Now(),
		}

		if err := mails.Save(out); err != nil {
			return "", err
		}

		return out.ID, nil
	}
}
