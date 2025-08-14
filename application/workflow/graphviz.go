// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

import (
	"bytes"
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"os/exec"
	"regexp"
	"strings"
	"unicode"
)

// gvAction represents a step/state in the flow
type gvAction struct {
	ID    string
	Label string
	Type  string // e.g. "start", "normal", "end", "error"
}

// gvEvent represents a triggering event with fields (UML-style)
type gvEvent struct {
	ID     string
	Name   string
	Fields []gvField // field name -> type
}

type gvField struct {
	Name string
	Type string
}

// Transition links Action -> Event -> Action
type gvTransition struct {
	From  string // Action ID
	Event string // Event ID
	To    string // Action ID
}

type gvComment struct {
	ID    string
	Label string
	To    []string
}

// FlowModel holds the entire structure
type gvFlowModel struct {
	Actions     []gvAction
	Events      []gvEvent
	Transitions []gvTransition
	Comments    []gvComment
	Label       string
}

func newGvFlowModel() *gvFlowModel {
	return &gvFlowModel{}
}

// GraphvizRenderer renders the FlowModel to Graphviz DOT code
func (fm gvFlowModel) RenderDOT(opts RenderOptions) string {
	var b strings.Builder

	// Graph header
	b.WriteString("digraph EventDrivenFlow {\n")
	b.WriteString("  rankdir=LR;\n")
	b.WriteString("  fontname=\"Helvetica\";\n")
	b.WriteString("  ratio=auto;\n") // dina4?
	b.WriteString("  node [fontname=\"Helvetica\"];\n")
	b.WriteString("  edge [fontname=\"Helvetica\"];\n\n")
	if fm.Label != "" {
		b.WriteString("  label=\"" + fm.Label + "\";\n")
	}

	// Actions
	for _, a := range fm.Actions {
		shape := "box"
		color := "#ddeeff"
		stereotype := "<br/><i>&lt;&lt;Action&gt;&gt;</i>"

		switch a.Type {
		case "start":
			shape = "circle"
			color = "lightgray"
			stereotype = ""
		case "end":
			shape = "doublecircle"
			color = "lightgray"
			stereotype = ""
		case "external":
			shape = "component"
			color = "#ea899a"
			stereotype = "<br/><i>&lt;&lt;Event Bus&gt;&gt;</i>"
		case "user":
			shape = "component"
			color = "#a1ebed"
			stereotype = "<br/><i>&lt;&lt;User&gt;&gt;</i>"
		case "error":
			color = "#ffdddd"
		}

		if !opts.ShowStereotypes {
			stereotype = ""
		}

		if !opts.ShowExternalEventBus && a.Type == "external" {
			continue
		}

		var mylabel string
		if stereotype != "" {
			mylabel = fmt.Sprintf("<%s %s>", a.Label, stereotype)
		} else {
			mylabel = fmt.Sprintf(`"%s"`, a.Label)
		}
		b.WriteString(fmt.Sprintf("  %s [label=%s, shape=%s, style=filled, fillcolor=\"%s\"];\n", a.ID, mylabel, shape, color))
	}

	b.WriteString("\n")

	// Events (as UML-style class tables)
	for _, e := range fm.Events {
		b.WriteString(fmt.Sprintf("  %s [shape=plaintext label=<\n", e.ID))
		b.WriteString("    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" BGCOLOR=\"white\">\n")
		// mix in some non-breaking spaces, because the svg export overlaps into the text glyphs, dot renders it wrong
		if opts.ShowStereotypes {
			b.WriteString(fmt.Sprintf("      <TR><TD COLSPAN=\"2\"><B>%s</B> &nbsp;<br/><i>&lt;&lt;Event&gt;&gt;</i></TD></TR>\n", e.Name))
		} else {
			b.WriteString(fmt.Sprintf("      <TR><TD COLSPAN=\"2\"><B>%s</B> &nbsp;</TD></TR>\n", e.Name))
		}

		if opts.ShowEventFields {
			for _, f := range e.Fields {
				//b.WriteString(fmt.Sprintf("      <TR><TD>%s</TD><TD>%s</TD></TR>\n", f.Name, f.Type)) // show with type
				b.WriteString(fmt.Sprintf("      <TR><TD COLSPAN=\"2\">%s</TD></TR>\n", f.Name)) // show just field name
			}
		}
		b.WriteString("    </TABLE>\n")
		b.WriteString("  >];\n")
	}

	b.WriteString("\n")

	// Transitions
	for _, t := range fm.Transitions {
		if !opts.ShowExternalEventBus && (strings.HasPrefix(t.From, "_external_") || strings.HasPrefix(t.To, "_external_")) {
			continue
		}

		if t.From != "" && t.To != "" && t.Event != "" {
			b.WriteString(fmt.Sprintf("  %s -> %s -> %s;\n", t.From, t.Event, t.To))
		} else {
			if t.From != "" && t.Event != "" {
				b.WriteString(fmt.Sprintf("  %s -> %s;\n", t.From, t.Event))
			}

			if t.To != "" && t.Event != "" {
				b.WriteString(fmt.Sprintf("  %s -> %s;\n", t.Event, t.To))
			}

		}
	}

	// comments

	if opts.ShowDescriptions {
		for _, c := range fm.Comments {
			b.WriteString(fmt.Sprintf(" %s [label=\"%s\", shape=note, style=filled, fillcolor=yellow];\n", c.ID, c.Label))
			for _, s := range c.To {
				b.WriteString(fmt.Sprintf("   %s -> %s [style=dashed, arrowhead=none];\n", s, c.ID))
			}
		}
	}

	b.WriteString("}\n")
	return b.String()
}

var re = regexp.MustCompile(`[^a-z0-9_]`)

// MakeValidID converts any string into a valid Graphviz-compatible identifier.
func makeValidID(input string) string {
	// Convert to lowercase for normalization
	id := strings.ToLower(input)

	// Replace German umlauts and sharp S (optional, customize as needed)
	id = strings.ReplaceAll(id, "ä", "ae")
	id = strings.ReplaceAll(id, "ö", "oe")
	id = strings.ReplaceAll(id, "ü", "ue")
	id = strings.ReplaceAll(id, "ß", "ss")

	// Replace any invalid characters with underscores
	id = re.ReplaceAllString(id, "_")

	// If the ID starts with a digit, prepend an underscore
	if len(id) > 0 && unicode.IsDigit(rune(id[0])) {
		id = "_" + id
	}

	// Collapse multiple consecutive underscores
	id = regexp.MustCompile(`_+`).ReplaceAllString(id, "_")

	// Trim underscores from beginning and end
	id = strings.Trim(id, "_")

	return id
}

func renderSVG(buf []byte) (core.SVG, error) {
	cmd := exec.Command("dot", "-Tsvg")

	// Provide DOT input via stdin
	cmd.Stdin = bytes.NewReader([]byte(buf))

	// Capture stdout (the SVG output)
	var out bytes.Buffer
	cmd.Stdout = &out

	// Capture stderr (for error messages from Graphviz)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("graphviz: failed to run dot: %s", err)
	}
	return out.Bytes(), nil
}
