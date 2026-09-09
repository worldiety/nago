// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package aispeclink offers the requirement catalogue of an application to a language model.
//
// # Why an assistant needs this
//
// Every domain tool answers a question about data. None of them answers the question a person actually has
// when the system refuses something they expected to work: whether the refusal is correct, and what it would
// take to change it. That answer is in the requirements.
//
// Without these tools a model does not decline to answer it. It invents a reason, fluently and plausibly,
// because that is what it is for - and an invented rule is worse than silence, since it sounds like the
// system speaking about itself.
//
// # How to use it
//
//	tools := append(myDomainTools, aispeclink.Tools(specMod.UseCases)...)
//
//	SystemPromptFunc: func() string {
//		return domainPrompt + "\n\n" + aispeclink.Index(wnd.Subject(), specMod.UseCases)
//	}
//
// The index goes into the prompt and the full text stays behind a tool. That split is what keeps the prompt
// from growing with the requirement tree: the model needs to know what exists in order to decide what is
// worth looking up, and it does not need every rationale on every turn.
package aispeclink

import (
	"fmt"
	"strings"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/speclink"
	"go.wdy.de/nago/auth"
)

// Tools returns the ready-made requirement tools.
//
// They go through the use cases like every other tool, so the disclosure ceiling and the permissions apply
// unchanged: an assistant reads exactly what the person operating it is allowed to read.
func Tools(uc speclink.UseCases) []completion.Tool {
	return []completion.Tool{
		listRequirementsTool(uc),
		readRequirementTool(uc),
		readCapabilitiesTool(uc),
		readSourceDocumentTool(uc),
	}
}

// readSourceDocumentTool hands out the document a requirement was derived from.
//
// It is often the better answer to "warum": a requirement is a sentence somebody distilled from what was
// asked for, and the source is what the person asking actually wrote.
func readSourceDocumentTool(uc speclink.UseCases) completion.Tool {
	type in struct {
		Path string `json:"path" desc:"Pfad des Quelldokuments, wie er im Feld sources einer Anforderung steht; ein Anker nach # wird ignoriert"`
	}
	type out struct {
		Path     string `json:"path"`
		Document string `json:"document" desc:"der Volltext des Dokuments"`
	}

	return completion.NewSubjectTool("read_source_document",
		"Liest das Dokument, aus dem eine Anforderung abgeleitet wurde — die Worte dessen, der die Anwendung bestellt hat. "+
			"Oft die bessere Antwort auf die Warum-Frage als die daraus destillierte Regel. Den Pfad liefert das Feld sources einer Anforderung.",
		func(subject auth.Subject, i in) (out, error) {
			doc, err := uc.FindSourceDocument(subject, i.Path)
			if err != nil {
				return out{}, err
			}

			return out{Path: i.Path, Document: doc}, nil
		})
}

// listRequirementsTool searches the catalogue.
func listRequirementsTool(uc speclink.UseCases) completion.Tool {
	return completion.NewSeqTool("list_requirements",
		"Durchsucht die Anforderungen dieser Anwendung: was gefordert ist, in welchem Zustand, und bei Entscheidungen auch warum. "+
			"Nutze das, um zu finden, welche Regel eine Situation betrifft; den Volltext holt read_requirement.",
		uc.FindAllRequirements).
		WithResultDoc()
}

// readRequirementTool hands out one requirement in full.
func readRequirementTool(uc speclink.UseCases) completion.Tool {
	type in struct {
		ID string `json:"id" desc:"Kennung der Anforderung, z. B. R-QUOTE-SUBMIT"`
	}

	return completion.NewSubjectTool("read_requirement",
		"Liest eine Anforderung im Volltext, bei Entscheidungen mit Begründung und mit dem, was die Entscheidung kostet. "+
			"Nutze das, wenn jemand fragt, warum sich das System so verhält, statt eine Begründung zu erfinden.",
		func(subject auth.Subject, i in) (speclink.Requirement, error) {
			opt, err := uc.FindRequirementByID(subject, speclink.ID(i.ID))
			if err != nil {
				return speclink.Requirement{}, err
			}

			if opt.IsNone() {
				// Deliberately does not distinguish "no such requirement" from "you may not see it": the
				// use case already collapses the two, and repeating the distinction here would give it back.
				return speclink.Requirement{}, fmt.Errorf(
					"keine sichtbare Anforderung %q; suche mit list_requirements nach dem passenden Begriff", i.ID)
			}

			return opt.Unwrap(), nil
		}).
		WithResultDoc()
}

// readCapabilitiesTool lists what the application does and on what grounds.
func readCapabilitiesTool(uc speclink.UseCases) completion.Tool {
	return completion.NewSeqTool("read_capabilities",
		"Listet die Fähigkeiten dieser Anwendung: was sie tut, für Menschen erklärt, und welche Anforderung die jeweilige Fähigkeit erfüllt. "+
			"Nutze das für Orientierungsfragen und um von dort auf read_requirement weiterzugehen.",
		uc.FindAllCapabilities).
		WithResultDoc()
}

// IndexLimit caps how many requirements the [Index] renders.
//
// A prompt is sent on every single turn, so an index of a large tree is paid for continuously. Beyond this
// many the index says how many were left out and points at list_requirements, which is the better tool for a
// tree that size anyway.
const IndexLimit = 200

// Index renders the requirements as one line each, for embedding into a system prompt.
//
// It goes through the use case, so it shows only what this subject may see - an index that named a
// requirement the reader cannot then open would be a worse kind of leak than showing it outright.
//
// An error is rendered as a note rather than returned. This is called while building a prompt, where there
// is nothing useful to do with an error and a turn that fails because the catalogue was unreadable is a poor
// trade against one that simply lacks the index.
func Index(subject auth.Subject, uc speclink.UseCases) string {
	var sb strings.Builder

	sb.WriteString("--- Anforderungen dieser Anwendung ---\n")
	sb.WriteString("Kennung [Art/Zustand] Titel — normativer Satz. Volltext und Begründung über read_requirement.\n")
	sb.WriteString("Zustand planned bedeutet: bewusst noch nicht umgesetzt. Das ist etwas anderes als „gibt es nicht“.\n\n")

	count := 0
	truncated := false

	for r, err := range uc.FindAllRequirements(subject, speclink.RequirementFilter{}) {
		if err != nil {
			sb.WriteString("(Die Anforderungen sind gerade nicht lesbar. Sag das, statt eine Begründung zu erfinden.)\n")
			return sb.String()
		}

		if count == IndexLimit {
			truncated = true
			break
		}

		fmt.Fprintf(&sb, "%s [%s/%s] %s — %s\n", r.ID, r.Kind, r.Status, r.Title, r.Text)
		count++
	}

	if count == 0 {
		sb.WriteString("(keine sichtbaren Anforderungen)\n")
		return sb.String()
	}

	if truncated {
		fmt.Fprintf(&sb, "\n… nur die ersten %d von mehr. Nutze list_requirements, um gezielt zu suchen.\n", IndexLimit)
	}

	return sb.String()
}
