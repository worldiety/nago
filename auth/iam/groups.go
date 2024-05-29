package iam

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/iter"
	"strings"
)

type Group struct {
	ID          auth.GID `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
}

func (r Group) Identity() auth.GID {
	return r.ID
}

type GroupRepository = data.Repository[Group, auth.GID]

func (s *Service) AllGroups(subject auth.Subject) iter.Seq2[Group, error] {
	if err := subject.Audit(ReadGroup); err != nil {
		return iter.Empty2[Group, error]()
	}

	return s.groups.Each
}

func (s *Service) CreateGroup(subject auth.Subject, group Group) error {
	if err := subject.Audit(CreateGroup); err != nil {
		return err
	}

	if strings.TrimSpace(string(group.ID)) == "" {
		group.ID = data.RandIdent[auth.GID]()
	}

	optGroup, err := s.groups.FindByID(group.ID)
	if err != nil {
		return err
	}

	if optGroup.Valid {
		return fmt.Errorf("cannot create group, because ID already exists: %v", group.ID)
	}

	return s.groups.Save(group)
}

func (s *Service) DeleteGroup(subject auth.Subject, id auth.GID) error {
	if err := subject.Audit(DeleteGroup); err != nil {
		return err
	}

	return s.groups.DeleteByID(id)
}

func (s *Service) UpdateGroup(subject auth.Subject, group Group) error {
	if err := subject.Audit(UpdateGroup); err != nil {
		return err
	}

	return s.groups.Save(group)
}
