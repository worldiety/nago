// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package speclink

import (
	"github.com/worldiety/i18n"
	"github.com/worldiety/speclink/spec"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
	"golang.org/x/text/language"
)

var (
	StrPermReadInternalName = i18n.MustString("nago.speclink.perm.read_internal.name", i18n.Values{
		language.German:  "Interne Anforderungen lesen",
		language.English: "Read internal requirements",
	})
	StrPermReadInternalDesc = i18n.MustString("nago.speclink.perm.read_internal.desc", i18n.Values{
		language.German:  "Träger dieser Berechtigung sehen auch Anforderungen, die als intern oder vertraulich eingestuft sind. Ohne sie sind nur öffentliche sichtbar.",
		language.English: "Holders also see requirements classified as internal or confidential. Without it only public ones are visible.",
	})
)

var (
	PermFindAllRequirements = permission.DeclareFindAll[FindAllRequirements]("nago.speclink.requirement.find_all", "Anforderung")
	PermFindRequirementByID = permission.DeclareFindByID[FindRequirementByID]("nago.speclink.requirement.find_by_id", "Anforderung")
	PermFindAllCapabilities = permission.DeclareFindAll[FindAllCapabilities]("nago.speclink.capability.find_all", "Fähigkeit")
	PermFindSourceDocument  = permission.DeclareFindByID[FindSourceDocument]("nago.speclink.source.find_by_id", "Quelldokument")

	// PermReadInternal raises the disclosure ceiling. It is a second permission on the same use cases rather
	// than a use case of its own, because it does not unlock a capability - it widens the answer of ones
	// that already exist.
	PermReadInternal = permission.Declare[FindAllRequirements](
		"nago.speclink.requirement.read_internal",
		StrPermReadInternalName.String(),
		StrPermReadInternalDesc.String(),
	)
)

// ceilingFor returns the highest classification this subject may be shown.
//
// # Why this exists at all
//
// speclink states plainly that Disclosure is not a control: "Nothing enforces it... The field states an
// intent for the consumer that shows the text to somebody". This package is that consumer, and an assistant
// reading requirements aloud to whoever asks is exactly the situation the field was introduced for. So the
// intent is turned into a rule here, once, rather than in every application.
//
// # Why the default is Public even though the zero value means Public
//
// The zero value of spec.Disclosure is Public deliberately - speclink argues, convincingly, that a
// requirement says what the system must do and that is what its users are entitled to know, and that reading
// silence as secrecy produces a catalogue nobody classifies honestly. So an unclassified requirement is
// public by intent, not by omission, and showing it is right.
//
// What this adds is the other end: somebody who did take the trouble to mark a requirement internal meant
// it, and the default reader must not see it.
func ceilingFor(subject auth.Subject) spec.Disclosure {
	if subject != nil && subject.HasPermission(PermReadInternal) {
		return spec.Confidential
	}

	return spec.Public
}

// visibleInList reports whether a requirement may appear in a listing for this ceiling.
//
// Secret is excluded from every listing regardless of permission, because that is what the classification
// means: "disclosed individually and never in bulk". A search that happened to match one would otherwise
// hand it out in exactly the way the class forbids.
func visibleInList(r spec.Requirement, ceiling spec.Disclosure) bool {
	if r.Disclosure == spec.Secret {
		return false
	}

	return r.Disclosure <= ceiling
}

// visibleIndividually reports whether a requirement may be handed out when asked for by identity.
//
// Secret is reachable here, and only here, for a subject that may read internal material. Asking for a
// specific identity is the individual disclosure the class allows; it also means the caller already knew the
// identity, so nothing is revealed by the lookup that was not known before it.
func visibleIndividually(r spec.Requirement, ceiling spec.Disclosure) bool {
	if r.Disclosure == spec.Secret {
		return ceiling >= spec.Confidential
	}

	return r.Disclosure <= ceiling
}
