// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

var (
	StrConfirmTitle = i18n.MustString("nago.ai.confirm.title", i18n.Values{
		language.German:  "Änderung bestätigen",
		language.English: "Confirm change",
	})
	StrConfirmIntro = i18n.MustString("nago.ai.confirm.intro", i18n.Values{
		language.German:  "Der Assistent möchte etwas ändern. Prüfen Sie, was passieren soll:",
		language.English: "The assistant wants to make a change. Check what is about to happen:",
	})
	StrConfirmApprove = i18n.MustString("nago.ai.confirm.approve", i18n.Values{
		language.German:  "Ausführen",
		language.English: "Execute",
	})
	StrConfirmDecline = i18n.MustString("nago.ai.confirm.decline", i18n.Values{
		language.German:  "Ablehnen",
		language.English: "Decline",
	})
	StrConfirmDeclined = i18n.MustString("nago.ai.confirm.declined", i18n.Values{
		language.German:  "Der Nutzer hat diese Änderung abgelehnt. Führe sie nicht aus und frage nach, wie stattdessen vorgegangen werden soll.",
		language.English: "The user declined this change. Do not perform it and ask how to proceed instead.",
	})
)

// prettyArguments formats the raw tool arguments for human review. The arguments are shown verbatim on
// purpose: a summary written by the model is exactly the thing that must not be trusted here.
func prettyArguments(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "{}"
	}

	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return string(raw)
	}

	return buf.String()
}

// readableArguments shows the arguments of a call to review them: each field of an object on its own line, and a
// multi-line string, e.g. a whole script, with its line breaks instead of escaped ones. The values stay verbatim,
// because an approval executes exactly these. Anything else falls back to [prettyArguments].
func readableArguments(raw json.RawMessage) string {
	dec := json.NewDecoder(bytes.NewReader(raw))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return prettyArguments(raw)
	}

	var sb strings.Builder
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return prettyArguments(raw)
		}

		key, _ := tok.(string)
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return prettyArguments(raw)
		}

		if sb.Len() > 0 {
			sb.WriteString("\n")
		}

		var str string
		if json.Unmarshal(value, &str) == nil && strings.Contains(str, "\n") {
			sb.WriteString(key + ":\n" + str + "\n")
			continue
		}

		sb.WriteString(key + ": " + prettyArguments(value))
	}

	if sb.Len() == 0 {
		return "{}"
	}

	return sb.String()
}
