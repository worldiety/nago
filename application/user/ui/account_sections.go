// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/alert"
)

// AccountSectionID identifies a section of the account view, AccountEntryID an entry of a section.
type AccountSectionID string
type AccountEntryID string

const (
	SectionAccount  AccountSectionID = "nago-account"  // „Konto“: EntryProfile, EntryTheme
	SectionSettings AccountSectionID = "nago-settings" // „Einstellungen“: EntryAdminCenter, if admin.HasEntries

	EntryProfile     AccountEntryID = "nago-account-profile"
	EntryTheme       AccountEntryID = "nago-account-theme"
	EntryAdminCenter AccountEntryID = "nago-settings-admin-center"
)

// AccountSectionModel is a section of the account view as data. The ID becomes the id of the rendered section.
type AccountSectionModel struct {
	ID      AccountSectionID
	Title   string
	Entries []AccountEntryModel
}

// AccountEntryModel is an entry of a section, e.g. an AccountAction or a ui.ToggleField.
type AccountEntryModel struct {
	ID   AccountEntryID
	View core.View
}

// AccountSections changes the sections of the account view of wnd. It receives nago's sections in their order, or
// what the AccountSections registered before returned, and returns the sections to show. A section without entries
// is not shown. Avatar, name, sign-out and legal texts are not part of the sections and stay. It runs on every
// render of the view, thus keep it cheap.
type AccountSections func(wnd core.Window, sections []AccountSectionModel) []AccountSectionModel

// InsertAccountSection returns sections with s after the section after and after the sections which follow it up to
// the next of nago's sections, so that several sections inserted after the same anchor keep the order of insertion.
// If there is no such section, s comes last.
func InsertAccountSection(sections []AccountSectionModel, after AccountSectionID, s AccountSectionModel) []AccountSectionModel {
	at := len(sections)
	for i, section := range sections {
		if section.ID != after {
			continue
		}

		at = i + 1
		for at < len(sections) && !nagoSection(sections[at].ID) {
			at++
		}

		break
	}

	// a new slice, the caller may still hold sections
	res := make([]AccountSectionModel, 0, len(sections)+1)
	res = append(res, sections[:at]...)
	res = append(res, s)
	res = append(res, sections[at:]...)
	return res
}

// nagoSection reports whether id is one of nago's sections.
func nagoSection(id AccountSectionID) bool {
	return id == SectionAccount || id == SectionSettings
}

// RemoveAccountEntry returns sections without the entry id; a section left without entries is not shown.
func RemoveAccountEntry(sections []AccountSectionModel, id AccountEntryID) []AccountSectionModel {
	// new slices, the caller may still hold sections and their entries
	res := make([]AccountSectionModel, 0, len(sections))
	for _, section := range sections {
		entries := make([]AccountEntryModel, 0, len(section.Entries))
		for _, entry := range section.Entries {
			if entry.ID != id {
				entries = append(entries, entry)
			}
		}

		section.Entries = entries
		res = append(res, section)
	}

	return res
}

// AccountDialog is the dialog „Nutzerkonto“ with the AccountView of wnd, as nago's scaffold shows it, or nil while
// presented is false. It is always closeable (alert.Closeable comes first, opts cannot remove it); opts place it,
// e.g. alert.Alignment. An application with its own frame shows the same dialog.
func AccountDialog(wnd core.Window, presented *core.State[bool], opts ...alert.Option) core.View {
	if !presented.Get() {
		return nil
	}

	return alert.Dialog("Nutzerkonto", AccountView(wnd), presented, append([]alert.Option{alert.Closeable()}, opts...)...)
}
