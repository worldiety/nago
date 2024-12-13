package admin

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/rquery"
)

func NewGroups(groups FindAllGroups) QueryGroups {
	return func(subject auth.Subject, filterText string) []Group {
		return filter(subject, groups(), filterText)
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
				// obviously not configured correctly or has never been setup, like disabled session or mail management etc.
				continue
			}

			if entry.Role != "" && !subject.HasRole(entry.Role) {
				continue
			}

			if entry.Permission != "" && !subject.HasPermission(entry.Permission) {
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
