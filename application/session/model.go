package session

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"time"
)

type ID string

type Session struct {
	ID              ID
	User            std.Option[user.ID]
	CreatedAt       time.Time
	AuthenticatedAt time.Time
}

func (s Session) Identity() ID {
	return s.ID
}

type Repository = data.Repository[Session, ID]
