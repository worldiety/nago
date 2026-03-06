// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dom

import "io"

// ── <form> ────────────────────────────────────────────────────────────────────

type Form struct{ element }

func NewForm() *Form { return &Form{newElement("form")} }

func (e *Form) node()        {}
func (e *Form) flowContent() {}

func (e *Form) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Form) Render(w io.Writer) error  { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <fieldset> ────────────────────────────────────────────────────────────────

type Fieldset struct{ element }

func NewFieldset() *Fieldset { return &Fieldset{newElement("fieldset")} }

func (e *Fieldset) node()        {}
func (e *Fieldset) flowContent() {}

// AppendChild accepts Legend (FieldsetContent) and FlowContent via AppendFlow.
func (e *Fieldset) AppendChild(n FieldsetContent) { e.appendChildNode(n) }
func (e *Fieldset) AppendFlow(n FlowContent)      { e.appendChildNode(n) }
func (e *Fieldset) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <legend> ──────────────────────────────────────────────────────────────────

type Legend struct{ element }

func NewLegend() *Legend { return &Legend{newElement("legend")} }

func (e *Legend) node()            {}
func (e *Legend) flowContent()     {}
func (e *Legend) fieldsetContent() {}

func (e *Legend) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Legend) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <label> ───────────────────────────────────────────────────────────────────

type Label struct{ element }

func NewLabel() *Label { return &Label{newElement("label")} }

func (e *Label) node()               {}
func (e *Label) flowContent()        {}
func (e *Label) phrasingContent()    {}
func (e *Label) interactiveContent() {}

func (e *Label) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Label) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <input> (void) ────────────────────────────────────────────────────────────

type Input struct{ voidElement }

func NewInput() *Input { return &Input{newVoidElement("input")} }

func (e *Input) node()               {}
func (e *Input) flowContent()        {}
func (e *Input) phrasingContent()    {}
func (e *Input) interactiveContent() {}

// ── <button> ──────────────────────────────────────────────────────────────────

type Button struct{ element }

func NewButton() *Button { return &Button{newElement("button")} }

func (e *Button) node()               {}
func (e *Button) flowContent()        {}
func (e *Button) phrasingContent()    {}
func (e *Button) interactiveContent() {}

func (e *Button) AppendChild(n PhrasingContent) { e.appendChildNode(n) }

// AppendFlow appends a FlowContent node (e.g. *Div) as a child.
func (e *Button) AppendFlow(n FlowContent) { e.appendChildNode(n) }
func (e *Button) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <select> ──────────────────────────────────────────────────────────────────

type Select struct{ element }

func NewSelect() *Select { return &Select{newElement("select")} }

func (e *Select) node()               {}
func (e *Select) flowContent()        {}
func (e *Select) phrasingContent()    {}
func (e *Select) interactiveContent() {}

// AppendChild accepts only <option> and <optgroup>.
func (e *Select) AppendChild(n SelectContent) { e.appendChildNode(n) }
func (e *Select) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <option> ──────────────────────────────────────────────────────────────────

type Option struct{ element }

func NewOption(text string) *Option {
	o := &Option{newElement("option")}
	o.kids = []Node{NewTextNode(text)}
	return o
}

func (e *Option) node()          {}
func (e *Option) selectContent() {}

func (e *Option) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <optgroup> ────────────────────────────────────────────────────────────────

type Optgroup struct{ element }

func NewOptgroup() *Optgroup { return &Optgroup{newElement("optgroup")} }

func (e *Optgroup) node()          {}
func (e *Optgroup) selectContent() {}

// AppendChild accepts only <option>.
func (e *Optgroup) AppendChild(n *Option) { e.appendChildNode(n) }
func (e *Optgroup) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <textarea> ────────────────────────────────────────────────────────────────

type Textarea struct{ element }

func NewTextarea() *Textarea { return &Textarea{newElement("textarea")} }

func (e *Textarea) node()               {}
func (e *Textarea) flowContent()        {}
func (e *Textarea) phrasingContent()    {}
func (e *Textarea) interactiveContent() {}

func (e *Textarea) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <datalist> ────────────────────────────────────────────────────────────────

type Datalist struct{ element }

func NewDatalist() *Datalist { return &Datalist{newElement("datalist")} }

func (e *Datalist) node()            {}
func (e *Datalist) flowContent()     {}
func (e *Datalist) phrasingContent() {}

// AppendChild accepts <option> elements.
func (e *Datalist) AppendChild(n *Option) { e.appendChildNode(n) }
func (e *Datalist) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <output> ──────────────────────────────────────────────────────────────────

type Output struct{ element }

func NewOutput() *Output { return &Output{newElement("output")} }

func (e *Output) node()               {}
func (e *Output) flowContent()        {}
func (e *Output) phrasingContent()    {}
func (e *Output) interactiveContent() {}

func (e *Output) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Output) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <progress> ────────────────────────────────────────────────────────────────

type Progress struct{ element }

func NewProgress() *Progress { return &Progress{newElement("progress")} }

func (e *Progress) node()               {}
func (e *Progress) flowContent()        {}
func (e *Progress) phrasingContent()    {}
func (e *Progress) interactiveContent() {}

func (e *Progress) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Progress) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <meter> ───────────────────────────────────────────────────────────────────

type Meter struct{ element }

func NewMeter() *Meter { return &Meter{newElement("meter")} }

func (e *Meter) node()            {}
func (e *Meter) flowContent()     {}
func (e *Meter) phrasingContent() {}

func (e *Meter) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Meter) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}
