package iam

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
	"strings"
)

type Role struct {
	ID          auth.RID `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Permissions []PID    `json:"permissions,omitempty"`
}

func (r Role) Identity() auth.RID {
	return r.ID
}

type RoleRepository = data.Repository[Role, auth.RID]

func (s *Service) AllRoles(subject auth.Subject) iter.Seq2[Role, error] {
	if err := subject.Audit(ReadRole); err != nil {
		return xiter.Empty2[Role, error]()
	}

	return s.roles.All()
}

func (s *Service) CreateRole(subject auth.Subject, role Role) error {
	if err := subject.Audit(CreateRole); err != nil {
		return err
	}

	if strings.TrimSpace(string(role.ID)) == "" {
		role.ID = data.RandIdent[auth.RID]()
	}

	optRole, err := s.roles.FindByID(role.ID)
	if err != nil {
		return err
	}

	if optRole.Valid {
		return fmt.Errorf("cannot create role, because ID already exists: %v", role.ID)
	}

	return s.roles.Save(role)
}

func (s *Service) DeleteRole(subject auth.Subject, id auth.RID) error {
	if err := subject.Audit(DeleteRole); err != nil {
		return err
	}

	return s.roles.DeleteByID(id)
}

func (s *Service) UpdateRole(subject auth.Subject, role Role) error {
	if err := subject.Audit(UpdateRole); err != nil {
		return err
	}

	return s.roles.Save(role)
}
