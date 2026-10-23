// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package admin

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/rquery"
	"log/slog"
)

func NewGroups(groups FindAllGroups) QueryGroups {
	return func(subject auth.Subject, filterText string) []Group {
		return filter(subject, groups(subject), filterText)
	}
}

func filter(subject auth.Subject, groups []Group, text string) []Group {
	var res []Group

	predicate := rquery.SimplePredicate[string](text)

	for _, group := range groups {
		fgrp := Group{
			Title: group.Title,
		}
		for _, entry := range group.Entries {
			if entry.Target == "" {
				slog.Error("admin center group entry target path is empty, card ignored", "title", entry.Title)
				// obviously not configured correctly or has never been setup, like disabled session or mail management etc.
				continue
			}

			if !permitted(subject, entry) {
				continue
			}

			if text != "" {
				if predicate(entry.Title) || predicate(entry.Text) {
					fgrp.Entries = append(fgrp.Entries, entry)
				}
			} else {
				fgrp.Entries = append(fgrp.Entries, entry)
			}

		}

		if len(fgrp.Entries) > 0 {
			// ignore entire and empty sections
			res = append(res, fgrp)
		}
	}

	return res
}

// AnyVisible returns true, if QueryGroups would show at least one of the cards of the given groups to the subject
// for an empty filter text. It stops at the first visible card and does not merge or sort anything.
func AnyVisible(subject auth.Subject, groups ...Group) bool {
	for _, group := range groups {
		for _, entry := range group.Entries {
			if entry.Target != "" && permitted(subject, entry) {
				return true
			}
		}
	}

	return false
}

func permitted(subject auth.Subject, entry Card) bool {
	if entry.Role != "" && !subject.HasRole(entry.Role) {
		return false
	}

	if entry.Permission != "" && !subject.HasPermission(entry.Permission) {
		return false
	}

	return true
}
