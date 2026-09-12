// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xerror

import (
	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

var (
	StrAccessDenied = i18n.MustString("nago.xerror.access_denied", i18n.Values{
		language.German:  "Zugriff verweigert",
		language.English: "Access denied",
	})

	StrNotLoggedInMsg = i18n.MustString("nago.xerror.not_logged_in_msg", i18n.Values{
		language.German:  "Diese Funktion steht nur eingeloggten Nutzern zur Verfügung.",
		language.English: "This function is only available to logged in users.",
	})

	StrDeniedMsg = i18n.MustString("nago.xerror.denied_msg", i18n.Values{
		language.German:  "Es besteht keine Berechtigung, um diese Inhalte oder Funktionen zu verwenden.",
		language.English: "You do not have permission to use this content or function.",
	})

	// StrDeniedGrantHintX is only appended when the required permission could actually be
	// named. Without a name the sentence would degenerate into an unusable instruction.
	StrDeniedGrantHintX = i18n.MustVarString("nago.xerror.denied_grant_hint_x", i18n.Values{
		language.German:  "Ein übergeordneter Rechteinhaber muss das Recht {permission} zunächst explizit erteilen.",
		language.English: "A superordinate rights holder must explicitly grant the permission {permission} first.",
	})

	StrNotFound = i18n.MustString("nago.xerror.not_found", i18n.Values{
		language.German:  "Element nicht gefunden",
		language.English: "Element not found",
	})

	StrNotFoundMsg = i18n.MustString("nago.xerror.not_found_msg", i18n.Values{
		language.German:  "Der Anwendungsfall konnte nicht ausgeführt werden, da ein Element erwartet aber nicht gefunden wurde.",
		language.English: "The use case could not be executed because an expected element was not found.",
	})

	StrAlreadyExists = i18n.MustString("nago.xerror.already_exists", i18n.Values{
		language.German:  "Element bereits vorhanden",
		language.English: "Element already exists",
	})

	StrAlreadyExistsMsg = i18n.MustString("nago.xerror.already_exists_msg", i18n.Values{
		language.German:  "Der Anwendungsfall konnte nicht ausgeführt werden, da ein Element nicht bereits vorhanden sein darf, aber gefunden wurde.",
		language.English: "The use case could not be executed because an element must not already exist, but was found.",
	})

	StrPasswordTooWeak = i18n.MustString("nago.xerror.password_too_weak", i18n.Values{
		language.German:  "Kennwort zu schwach",
		language.English: "Password too weak",
	})

	StrPwComplexity = i18n.MustString("nago.xerror.pw_complexity", i18n.Values{
		language.German:  "Die Kennwortkomplexität ist zu niedrig.",
		language.English: "The password complexity is too low.",
	})

	StrPwUpperLower = i18n.MustString("nago.xerror.pw_upper_lower", i18n.Values{
		language.German:  "Das Kennwort muss mindestens einen Groẞ- und einen Kleinbuchstaben enthalten.",
		language.English: "The password must contain at least one upper and one lower case letter.",
	})

	StrPwMinLengthX = i18n.MustVarString("nago.xerror.pw_min_length_x", i18n.Values{
		language.German:  "Das Kennwort muss mindestens {min} Zeichen enthalten.",
		language.English: "The password must contain at least {min} characters.",
	})

	StrPwSpecial = i18n.MustString("nago.xerror.pw_special", i18n.Values{
		language.German:  "Das Kennwort muss mindestens ein Sonderzeichen enthalten.",
		language.English: "The password must contain at least one special character.",
	})

	StrPwTooLong = i18n.MustString("nago.xerror.pw_too_long", i18n.Values{
		language.German:  "Das Kennwort ist zu lang.",
		language.English: "The password is too long.",
	})

	StrPwNumber = i18n.MustString("nago.xerror.pw_number", i18n.Values{
		language.German:  "Das Kennwort muss mindestens eine Zahl enthalten.",
		language.English: "The password must contain at least one number.",
	})

	StrPwUnusable = i18n.MustString("nago.xerror.pw_unusable", i18n.Values{
		language.German:  "Das Kennwort kann nicht verwendet werden.",
		language.English: "The password cannot be used.",
	})

	StrValidationFailed = i18n.MustString("nago.xerror.validation_failed", i18n.Values{
		language.German:  "Eingabe unvollständig",
		language.English: "Invalid input",
	})

	StrValidationFailedMsg = i18n.MustString("nago.xerror.validation_failed_msg", i18n.Values{
		language.German:  "Bitte prüfen Sie die markierten Felder.",
		language.English: "Please check the highlighted fields.",
	})

	StrDeniedRetryHint = i18n.MustString("nago.xerror.denied_retry_hint", i18n.Values{
		language.German:  "Dies ist ein Berechtigungsfehler. Ein erneuter identischer Aufruf wird ebenfalls fehlschlagen.",
		language.English: "This is an authorization failure. Repeating the identical call will fail as well.",
	})

	StrUnexpected = i18n.MustString("nago.xerror.unexpected", i18n.Values{
		language.German:  "Fehler",
		language.English: "Error",
	})

	StrUnexpectedMsgX = i18n.MustVarString("nago.xerror.unexpected_msg_x", i18n.Values{
		language.German:  "Ein unerwarteter Fehler ist aufgetreten. Sie können sich mit dem folgenden Code an den Support wenden: {token}",
		language.English: "An unexpected error occurred. You can contact support with the following code: {token}",
	})
)
