// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package usercircle

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
	"slices"
	"strings"
)

func NewMyCircleMembers(repoCircle Repository, findAllUsers user.FindAll) MyCircleMembers {
	return func(subject auth.Subject, id ID) iter.Seq2[user.User, error] {
		optCircle, err := repoCircle.FindByID(id)
		if err != nil {
			return xiter.WithError[user.User](err)
		}

		if optCircle.IsNone() {
			return func(yield func(user.User, error) bool) {}
		}

		circle := optCircle.Unwrap()
		if !slices.Contains(circle.Administrators, subject.ID()) {
			return xiter.WithError[user.User](user.PermissionDeniedErr)
		}

		circleLkp := circleLookups{
			users:     make(map[user.ID]struct{}),
			groups:    make(map[group.ID]struct{}),
			roles:     make(map[role.ID]struct{}),
			blacklist: make(map[user.ID]struct{}),
		}

		for _, domain := range circle.MemberRuleDomains {
			trimmed := strings.ToLower(strings.TrimSpace(string(domain)))
			if trimmed == "" {
				continue
			}
			circleLkp.domains = append(circleLkp.domains, trimmed)
		}

		for _, ruleGroup := range circle.MemberRuleGroups {
			circleLkp.groups[ruleGroup] = struct{}{}
		}

		for _, ruleRole := range circle.MemberRuleRoles {
			circleLkp.roles[ruleRole] = struct{}{}
		}

		for _, ruleUser := range circle.MemberRuleUsers {
			circleLkp.users[ruleUser] = struct{}{}
		}

		for _, ruleUser := range circle.MemberRuleUsersBlacklist {
			circleLkp.blacklist[ruleUser] = struct{}{}
		}

		return func(yield func(user.User, error) bool) {
			for usr, err := range findAllUsers(user.SU()) {
				if err != nil {
					if !yield(usr, err) {
						return
					}

					continue
				}

				if !circleLkp.isMember(usr) {
					continue
				}

				if !yield(usr, nil) {
					return
				}
			}
		}
	}

}

type circleLookups struct {
	domains   []string
	users     map[user.ID]struct{}
	blacklist map[user.ID]struct{}
	groups    map[group.ID]struct{}
	roles     map[role.ID]struct{}
}

func (c *circleLookups) isMember(usr user.User) bool {
	if _, ok := c.blacklist[usr.ID]; ok {
		return false
	}

	if len(c.users) == 0 && len(c.groups) == 0 && len(c.domains) == 0 {
		return true
	}

	if len(c.domains) > 0 {
		for _, domain := range c.domains {
			if strings.HasSuffix(string(usr.Email), domain) {
				return true
			}
		}
	}

	if _, ok := c.users[usr.ID]; ok {
		return true
	}

	for _, gid := range usr.Groups {
		if _, ok := c.groups[gid]; ok {
			return true
		}
	}

	for _, rid := range usr.Roles {
		if _, ok := c.roles[rid]; ok {
			return true
		}
	}

	return false
}
