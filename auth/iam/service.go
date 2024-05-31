package iam

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/pkg/slices"
	"log/slog"
	"sync"
	"time"
)

type Service struct {
	permissions *Permissions
	users       UserRepository
	sessions    SessionRepository
	roles       RoleRepository
	groups      GroupRepository
	mutex       sync.Mutex
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
	usr.Permissions = []PID{
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

	return nil
}

func (s *Service) AllPermissions(subject auth.Subject) iter.Seq2[Permission, error] {
	if err := subject.Audit(ReadPermission); err != nil {
		slog.Error("insufficient permission", "err", err)
		return iter.Empty2[Permission, error]()
	}

	return s.permissions.Each
}

func (s *Service) AllUsers(subject auth.Subject) iter.Seq2[User, error] {
	if err := subject.Audit(ReadUser); err != nil {
		slog.Error("insufficient permission", "err", err)
		return iter.Empty2[User, error]()
	}

	return s.users.Each
}

func (s *Service) AllPermissionsByIDs(subject auth.Subject, ids ...PID) iter.Seq2[Permission, error] {
	if err := subject.Audit(ReadPermission); err != nil {
		slog.Error("insufficient permission", "err", err)
		return iter.Empty2[Permission, error]()
	}

	var tmp []Permission
	for _, id := range ids {
		if p, ok := s.permissions.Get(id); ok {
			tmp = append(tmp, p)
		}
	}

	return slices.Values2[[]Permission, Permission, error](tmp)
}
