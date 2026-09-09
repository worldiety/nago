// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package cfgspeclink wires the requirement catalogue into a running application.
package cfgspeclink

import (
	"fmt"
	"io/fs"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/speclink"
	uispeclink "go.wdy.de/nago/application/speclink/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/layout"
	"golang.org/x/text/language"
)

var (
	StrAdminGroup = i18n.MustString("nago.speclink.admin.group", i18n.Values{
		language.German:  "Spezifikation",
		language.English: "Specification",
	})
	StrAdminCardDesc = i18n.MustString("nago.speclink.admin.card_desc", i18n.Values{
		language.German:  "Die Anforderungen durchsehen, die in dieses Programm einkompiliert sind.",
		language.English: "Browse the requirements compiled into this program.",
	})

	StrRoleReaderName = i18n.MustString("nago.speclink.role.reader.name", i18n.Values{
		language.German:  "Anforderungen lesen",
		language.English: "Requirements Reader",
	})
	StrRoleReaderDesc = i18n.MustString("nago.speclink.role.reader.desc", i18n.Values{
		language.German:  "Erlaubt es, die Anforderungen der Anwendung zu lesen — in der Verwaltung und über den KI-Assistenten. Öffentlich eingestufte Anforderungen; für interne braucht es zusätzlich die Berechtigung „Interne Anforderungen lesen“.",
		language.English: "Allows reading the application's requirements, in the administration and through the AI assistant. Publicly classified ones; internal material additionally needs the „Read internal requirements“ permission.",
	})
)

// RoleRequirementsReader is the system role that lets somebody read the requirement catalogue.
//
// It is shipped as a role rather than documented as three permissions for the same reason the assistant role
// is: a list of identifiers an operator has to reproduce by hand is a list that gets reproduced wrongly, and
// the failure is silent - an empty page, or an assistant that answers "warum" with a permission error.
const RoleRequirementsReader role.ID = "nago.speclink.reader"

// Management is what the rest of the application depends on.
type Management struct {
	UseCases speclink.UseCases
	Pages    uispeclink.Pages
}

// Options configures optional parts of the integration.
type Options struct {
	// SourceDocuments holds the documents the requirements were derived from, usually an embed.FS over
	// requirements/_sources. When set, the assistant can read the wording the requirements were distilled
	// from, which is frequently the better answer to "warum".
	//
	// Optional. Without it the reading use case reports that no documents were embedded, rather than being
	// absent - a caller does not have to test for a missing capability.
	SourceDocuments fs.FS
}

// Enable makes the requirement catalogue of this binary readable.
//
// The catalogue itself needs no wiring: spec.Declare fills it during package initialisation of whatever the
// binary links in. What is wired here is the authorization, the screen and the role.
func Enable(cfg *application.Configurator, opts ...Options) (Management, error) {
	management, ok := core.FromContext[Management](cfg.Context(), "")
	if ok {
		return management, nil
	}

	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}

	uc := speclink.NewUseCases(opt.SourceDocuments)

	if err := cfg.DeclareSystemRole(role.Role{
		ID:          RoleRequirementsReader,
		Name:        StrRoleReaderName.Get(cfg.SysUser()),
		Description: StrRoleReaderDesc.Get(cfg.SysUser()),
	}, ReaderPermissions()...); err != nil {
		return Management{}, fmt.Errorf("cannot declare the requirements reader role: %w", err)
	}

	management = Management{
		UseCases: uc,
		Pages: uispeclink.Pages{
			Requirements: "admin/speclink/requirements",
		},
	}

	cfg.RootViewWithDecoration(management.Pages.Requirements, func(wnd core.Window) core.View {
		return layout.WithBackButton(wnd, uispeclink.PageRequirements(wnd, uc))
	}, application.Purpose(
		"Die Anforderungen durchsehen, die in dieses Programm einkompiliert sind: was gefordert ist, in welchem Zustand, und bei Entscheidungen warum."))

	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		return admin.Group{
			Title: StrAdminGroup.Get(subject),
			Entries: []admin.Card{
				{
					Title:      StrTitleOf(subject),
					Text:       StrAdminCardDesc.Get(subject),
					Target:     management.Pages.Requirements,
					Permission: speclink.PermFindAllRequirements,
				},
			},
		}
	})

	cfg.AddContextValue(core.ContextValue("", management))

	return management, nil
}

// StrTitleOf is the page title in the subject's language, reused for the admin card so the two cannot drift.
func StrTitleOf(subject auth.Subject) string {
	return uispeclink.StrTitle.Get(subject)
}

// ReaderPermissions are what [RoleRequirementsReader] carries.
//
// PermReadInternal is deliberately not among them: reading internal material is a decision about a person,
// not about a feature, and bundling it here would hand it out to everybody who may read anything at all.
func ReaderPermissions() []permission.ID {
	return []permission.ID{
		speclink.PermFindAllRequirements,
		speclink.PermFindRequirementByID,
		speclink.PermFindAllCapabilities,
		speclink.PermFindSourceDocument,
	}
}
