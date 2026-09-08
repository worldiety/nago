// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ailibrary

import (
	"fmt"
	"strings"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/example/cmd/tutorial-113-ai-assistant/app/library"
)

// knowledgeTools let the assistant answer "warum ist das so".
//
// Every domain tool answers a question about the data. None of them answers the question a person actually
// has when the system refuses something they expected to work, and that question is not about data at all -
// it is whether the refusal is correct and what it would take to change it.
//
// Without these tools the model does not decline to answer. It invents a reason, fluently and plausibly,
// because that is what it is for. An invented rule is worse than silence: it sounds like the system speaking
// about itself.
//
// # These take no subject
//
// A decision is not somebody's data, so [completion.NewTool] is right here and passing the subject would
// suggest a check that does not happen. The rule from tools.go still holds for everything that touches
// records: those go through a use case, and a use case audits.
//
// Whether the reasoning itself needs guarding is a real question and the answer is a decision, not an
// oversight. Here it is public: the rules of a library are not confidential, and somebody who may see the
// stock may know why they may not borrow from it. In a system where the reasoning names customers,
// contracts or thresholds somebody could game, put it behind a use case like everything else.
func knowledgeTools() []completion.Tool {
	return []completion.Tool{
		readDecisionTool(),
		readCapabilitiesTool(),
	}
}

// readDecisionTool hands out the full text of one recorded decision.
//
// The index of what exists is already in the system prompt (see [SystemPrompt]), so the model asks for a
// specific one rather than fetching everything. That is the split that keeps the prompt from growing with
// the project.
func readDecisionTool() completion.Tool {
	type in struct {
		ID string `json:"id" desc:"Kennung der Entscheidung, z. B. DEC-AVAILABILITY-DERIVED. Die Liste steht im System-Prompt."`
	}

	return completion.NewTool("read_decision",
		"Liest eine Entscheidung im Volltext: was entschieden wurde, warum, und was es kostet. "+
			"Nutze das, wenn jemand fragt, warum sich das System so verhält, statt eine Begründung zu erfinden.",
		func(i in) (library.Decision, error) {
			dec, ok := library.FindDecision(i.ID)
			if !ok {
				// Naming what does exist turns a dead end into the next step. A bare "not found" invites the
				// model to guess another identifier, and it will.
				return library.Decision{}, fmt.Errorf(
					"keine Entscheidung %q; bekannt sind: %s", i.ID, knownDecisionIDs())
			}

			return dec, nil
		}).
		WithResultDoc()
}

// readCapabilitiesTool lists what the library can do and what each of those rests on.
//
// It answers "was kann das System" and "worauf beruht das" in one call, which is what somebody asking for
// orientation actually needs - the two halves separately are not useful.
func readCapabilitiesTool() completion.Tool {
	type in struct {
		Query string `json:"query" optional:"true" desc:"Suchbegriff, z. B. ausleihen oder Bestand. Leer liefert alles."`
	}

	return completion.NewTool("read_capabilities",
		"Listet die Fähigkeiten der Bibliothek: was sie tut, für Menschen erklärt, und auf welchen Entscheidungen sie beruht. "+
			"Nutze das für Orientierungsfragen und um von dort auf read_decision weiterzugehen.",
		func(i in) ([]library.Capability, error) {
			needle := strings.ToLower(strings.TrimSpace(i.Query))
			if needle == "" {
				return library.Capabilities(), nil
			}

			var out []library.Capability
			for _, c := range library.Capabilities() {
				if matchesAny(needle, c.UseCase, c.Tool, c.Help) {
					out = append(out, c)
				}
			}

			return out, nil
		}).
		WithResultDoc()
}

// matchesAny reports whether the needle occurs in any of the texts.
func matchesAny(needle string, texts ...string) bool {
	for _, t := range texts {
		if strings.Contains(strings.ToLower(t), needle) {
			return true
		}
	}

	return false
}

// knownDecisionIDs is what an unknown identifier is answered with.
func knownDecisionIDs() string {
	ids := make([]string, 0, len(library.Decisions()))
	for _, d := range library.Decisions() {
		ids = append(ids, d.ID)
	}

	return strings.Join(ids, ", ")
}
