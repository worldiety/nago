// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisecret

import (
	"fmt"
	"strings"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

// groupNames resolves the names of the groups. A group, which the subject cannot read, is shown by its id.
func groupNames(subject auth.Subject, findGroupByID group.FindByID, gids []group.ID) []string {
	var names []string
	for _, gid := range gids {
		name := string(gid)
		if findGroupByID != nil {
			if optGrp, err := findGroupByID(subject, gid); err == nil && optGrp.IsSome() && optGrp.Unwrap().Name != "" {
				name = optGrp.Unwrap().Name
			}
		}

		names = append(names, name)
	}

	return names
}

// accessSummary tells in a short line, who has access to the secret beside the subject, e.g. for the vault list.
func accessSummary(subject auth.Subject, findGroupByID group.FindByID, scr secret.Secret) string {
	others := 0
	for _, uid := range scr.Owners {
		if uid != subject.ID() {
			others++
		}
	}

	var parts []string
	switch {
	case others == 1:
		parts = append(parts, "1 weiteren Person")
	case others > 1:
		parts = append(parts, fmt.Sprintf("%d weiteren Personen", others))
	}

	switch names := groupNames(subject, findGroupByID, scr.Groups); len(names) {
	case 0:
	case 1:
		parts = append(parts, "der Gruppe "+names[0])
	default:
		parts = append(parts, "den Gruppen "+joinAnd(names))
	}

	if len(parts) == 0 {
		if scr.IsOwner(subject) {
			return "Nur Sie"
		}

		return "Kein Zugriff für andere"
	}

	return "Geteilt mit " + strings.Join(parts, " und ")
}

// accessSentence lists everyone who has access, e.g. below the sharing editor or in the delete dialog.
func accessSentence(wnd core.Window, users []user.User, groups []string) string {
	var people []string
	for _, usr := range users {
		if usr.ID == wnd.Subject().ID() {
			people = append([]string{"Sie"}, people...)
			continue
		}

		name := strings.TrimSpace(usr.Contact.Firstname + " " + usr.Contact.Lastname)
		if name == "" {
			name = string(usr.Email)
		}

		people = append(people, name)
	}

	switch {
	case len(people) == 0 && len(groups) == 0:
		return "Niemand hat Zugriff."
	case len(groups) == 0:
		return "Zugriff haben: " + joinAnd(people) + "."
	case len(people) == 0:
		return "Zugriff haben alle Mitglieder von " + joinAnd(groups) + "."
	default:
		return "Zugriff haben: " + strings.Join(people, ", ") + " und alle Mitglieder von " + joinAnd(groups) + "."
	}
}

func joinAnd(items []string) string {
	if len(items) <= 1 {
		return strings.Join(items, "")
	}

	return strings.Join(items[:len(items)-1], ", ") + " und " + items[len(items)-1]
}
