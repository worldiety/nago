// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisecret

import (
	"encoding/hex"
	"reflect"
	"slices"
	"strings"

	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/list"
)

func CreateSecretPage(wnd core.Window, pages Pages, createSecret secret.CreateSecret) core.View {
	var items []core.View
	for _, spec := range getGlobalCredentialTypeSpecs() {
		items = append(items, credentialType(wnd, pages, createSecret, spec))
	}

	return ui.VStack(
		ui.H1("Neues Secret anlegen"),
		list.List(items...).Caption(ui.Text("Verfügbare Geheimnis-Varianten")).Frame(ui.Frame{}.FullWidth()),
		ui.HStack(ui.SecondaryButton(func() {
			wnd.Navigation().Back()
		}).Title(rstring.ActionBack.Get(wnd))).
			FullWidth().
			Alignment(ui.Trailing).
			Padding(ui.Padding{Top: ui.L16}),
	).Alignment(ui.Leading).FullWidth()
}

type credentialTypeSpec struct {
	name        string
	description string
	logo        string
	refType     reflect.Type
}

func (s credentialTypeSpec) LogoView() core.View {
	logoUrl := s.logo

	border := ui.Border{}
	var ico core.View
	switch {
	case strings.HasPrefix(logoUrl, "http"):
		ico = avatar.URI(core.URI(logoUrl)).Border(border)
	case logoUrl == "":
		ico = avatar.Text(s.refType.Name())
	default:
		buf, err := hex.DecodeString(logoUrl)
		if err == nil {
			ico = avatar.Embed(buf).Border(border)
		} else {
			// hm, anything else? what about our own resource uris? or shortcuts to embedded icons?
		}

	}

	return ico
}

func newCredentialTypeSpec(rtype reflect.Type) credentialTypeSpec {
	var name string
	var description string
	var logoUrl string
	field, ok := rtype.FieldByName("_")
	if ok {
		description = field.Tag.Get("credentialDescription")
		name = field.Tag.Get("credentialName")
		logoUrl = field.Tag.Get("credentialLogo")
	}

	if name == "" {
		name = rtype.Name()
	}

	return credentialTypeSpec{
		name:        name,
		description: description,
		logo:        logoUrl,
		refType:     rtype,
	}
}

func getGlobalCredentialTypeSpecs() []credentialTypeSpec {
	decl, ok := enum.DeclarationFor[secret.Credentials]()
	if !ok {
		panic("unreachable: secret.Credentials declaration not defined")
	}

	var res []credentialTypeSpec
	for rtype := range decl.Variants() {

		res = append(res, newCredentialTypeSpec(rtype))
	}

	slices.SortFunc(res, func(a, b credentialTypeSpec) int {
		return strings.Compare(a.name, b.name)
	})

	return res
}

func credentialType(wnd core.Window, pages Pages, createSecret secret.CreateSecret, spec credentialTypeSpec) core.View {
	return list.Entry().
		Leading(spec.LogoView()).
		Headline(wnd.Bundle().Resolve(spec.name)).
		SupportingText(wnd.Bundle().Resolve(spec.description)).
		Trailing(ui.PrimaryButton(func() {
			value := reflect.New(spec.refType).Elem().Interface().(secret.Credentials)
			id, err := createSecret(wnd.Subject(), value)
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}

			wnd.Navigation().ForwardTo(pages.EditSecret, core.Values{"id": string(id)})
		}).PreIcon(heroOutline.Plus).Title(rstring.ActionAdd.Get(wnd)))
}
