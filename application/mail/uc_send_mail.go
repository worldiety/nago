package mail

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"strings"
	"time"
)

func NewSendMail(mails Repository) SendMail {
	return func(subject auth.Subject, mail Mail) (ID, error) {
		if err := subject.Audit(PermSendMail); err != nil {
			return "", err
		}

		var tmp []string
		for _, rec := range mail.To {
			tmp = append(tmp, rec.String())
		}

		for _, address := range mail.CC {
			tmp = append(tmp, address.String())
		}

		for _, address := range mail.BCC {
			tmp = append(tmp, address.String())
		}

		out := Outgoing{
			ID:       data.RandIdent[ID](),
			Mail:     mail,
			Receiver: strings.Join(tmp, ", "),
			Subject:  mail.Subject,
			Status:   StatusQueued,
			QueuedAt: time.Now(),
		}

		if err := mails.Save(out); err != nil {
			return "", err
		}

		return out.ID, nil
	}
}
