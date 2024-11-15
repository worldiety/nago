package iam

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xiter"
	"go.wdy.de/nago/pkg/xslices"
	"iter"
	"log/slog"
	"sync"
	"time"
)

var (
	bootstrapAdminPermissions []PID = []PID{
		CreateUser,
		ReadUser,
		UpdateUser,
		DeleteUser,
		ReadPermission,

		CreateRole,
		ReadRole,
		UpdateRole,
		DeleteRole,

		CreateGroup,
		ReadGroup,
		UpdateGroup,
		DeleteGroup,
	}
)

type Service struct {
	permissions       *Permissions
	users             UserRepository
	sessions          SessionRepository
	roles             RoleRepository
	groups            GroupRepository
	mutex             sync.Mutex
	bootstrapAdminUid *auth.UID
}

func NewService(permissions *Permissions, users UserRepository, sessions SessionRepository, roles RoleRepository, groups GroupRepository) *Service {
	return &Service{
		permissions: permissions,
		users:       users,
		sessions:    sessions,
		roles:       roles,
		groups:      groups,
	}
}

func (s *Service) Bootstrap() error {
	optUsr, err := s.userByMail("admin@localhost")
	if err != nil {
		return err
	}

	var usr User

	if optUsr.Valid {
		usr = optUsr.Unwrap()
	} else {
		usr.ID = data.RandIdent[auth.UID]() // a random admin user id makes some attacks impossible
		usr.Email = "admin@localhost"
		usr.Firstname = "admin"
		usr.Lastname = "admin"
	}

	// make the bootstrap admin live only 1 hour until it is disabled automatically
	deadline := time.Now().Add(time.Minute * 60)
	usr.Status = AccountStatus{}.WithEnabledUntil(EnabledUntil{ValidUntil: deadline})

	// we are not allowed to have domain specific permissions, only those to bootstrap other users.
	// even admins must not see customers secret domain stuff.
	usr.Permissions = bootstrapAdminPermissions

	// we always generate a new random password at startup, which makes accidental information leak
	// from environment variables impossible. You must compromise the machine and log file, which only involves
	// the actual host and not the entire CI/CD environment.
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Errorf("not enough entropy: %w", err)
	}

	// ensure that we have no copy-paste problems
	randPwd := hex.EncodeToString(buf[:])
	if err := setArgon2idMin(&usr, randPwd); err != nil {
		return fmt.Errorf("cannot set argon2id: %w", err)
	}

	if err := s.users.Save(usr); err != nil {
		return fmt.Errorf("cannot save bootstrap user: %w", err)
	}

	slog.Info("enabled bootstrap admin user with 1 hour lifetime and random password", "login", usr.Email, "password", randPwd, "validUntil", deadline)

	s.bootstrapAdminUid = &usr.ID
	return nil
}

// IsBootstrapAdminAcccount checks whether the given subject is the boostrap admin account
// When error is nil, the check was successful, otherwise error tells you what went wrong
func (s *Service) IsBootstrapAdminAcccount(subject auth.Subject) error {
	if s.bootstrapAdminUid == nil {
		return fmt.Errorf("No boostrap admin was created")
	}

	if subject.ID() != *s.bootstrapAdminUid {
		return fmt.Errorf("id does not match")
	}

	for _, neededPermission := range bootstrapAdminPermissions {
		if err := subject.Audit(neededPermission); err != nil {
			return fmt.Errorf("Not all permissions for the boostrap admin available: %w", err)
		}
	}

	return nil
}

func (s *Service) AllPermissions(subject auth.Subject) iter.Seq2[Permission, error] {
	if err := subject.Audit(ReadPermission); err != nil {
		return xiter.WithError[Permission](err)
	}

	return s.permissions.Each
}

func (s *Service) AllUsers(subject auth.Subject) iter.Seq2[User, error] {
	if err := subject.Audit(ReadUser); err != nil {
		return xiter.WithError[User](err)
	}

	return s.users.All()
}

func (s *Service) AllPermissionsByIDs(subject auth.Subject, ids ...PID) iter.Seq2[Permission, error] {
	if err := subject.Audit(ReadPermission); err != nil {
		return xiter.WithError[Permission](err)
	}

	var tmp []Permission
	for _, id := range ids {
		if p, ok := s.permissions.Get(id); ok {
			tmp = append(tmp, p)
		}
	}

	return xslices.Values2[[]Permission, Permission, error](tmp)
}

// TODO add observer to changed users
