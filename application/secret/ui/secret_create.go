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
	"strconv"
	"strings"

	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/list"
)

// CreateSecretPage lets the user choose the type of a new secret and then fill it in. Nothing is stored, before the
// user saves it, so that leaving the page does not leave an empty secret behind.
func CreateSecretPage(wnd core.Window, pages Pages, uc UseCases) core.View {
	if key := wnd.Values()["type"]; key != "" {
		spec, ok := credentialTypeSpecByKey(key)
		if !ok {
			return alert.Banner("Nicht gefunden", "Diese Art von Secret ist nicht verfügbar.")
		}

		draft := secret.Secret{
			Owners:      []user.ID{wnd.Subject().ID()},
			Credentials: reflect.New(spec.refType).Elem().Interface().(secret.Credentials),
		}

		return secretEditor(wnd, pages, uc, draft, true)
	}

	var items []core.View
	for _, spec := range getGlobalCredentialTypeSpecs() {
		items = append(items, credentialType(wnd, pages, spec))
	}

	return ui.VStack(
		ui.H1("Neues Secret anlegen"),
		list.List(items...).Caption(ui.Text("Welche Art von Secret möchten Sie anlegen?")).Frame(ui.Frame{}.FullWidth()),
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
	hidden      bool
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
	var hidden bool
	field, ok := rtype.FieldByName("_")
	if ok {
		description = field.Tag.Get("credentialDescription")
		name = field.Tag.Get("credentialName")
		logoUrl = field.Tag.Get("credentialLogo")
		hidden, _ = strconv.ParseBool(field.Tag.Get("credentialHidden"))
	}

	if name == "" {
		name = rtype.Name()
	}

	return credentialTypeSpec{
		name:        name,
		description: description,
		logo:        logoUrl,
		hidden:      hidden,
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
		spec := newCredentialTypeSpec(rtype)
		if spec.hidden {
			continue
		}

		res = append(res, spec)
	}

	slices.SortFunc(res, func(a, b credentialTypeSpec) int {
		return strings.Compare(a.name, b.name)
	})

	return res
}

func credentialType(wnd core.Window, pages Pages, spec credentialTypeSpec) core.View {
	return list.Entry().
		Leading(spec.LogoView()).
		Headline(wnd.Bundle().Resolve(spec.name)).
		SupportingText(wnd.Bundle().Resolve(spec.description)).
		Trailing(ui.ImageIcon(heroOutline.ChevronRight)).
		Action(func() {
			wnd.Navigation().ForwardTo(pages.CreateSecret, core.Values{"type": spec.key()})
		})
}

// key identifies the type of credentials, e.g. in a navigation parameter.
func (s credentialTypeSpec) key() string {
	return s.refType.PkgPath() + "." + s.refType.Name()
}

func credentialTypeSpecByKey(key string) (credentialTypeSpec, bool) {
	for _, spec := range getGlobalCredentialTypeSpecs() {
		if spec.key() == key {
			return spec, true
		}
	}

	return credentialTypeSpec{}, false
}
