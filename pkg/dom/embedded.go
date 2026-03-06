// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dom

import "io"

// ── <img> (void) ──────────────────────────────────────────────────────────────

type Img struct{ voidElement }

func NewImg() *Img { return &Img{newVoidElement("img")} }

func (e *Img) node()               {}
func (e *Img) flowContent()        {}
func (e *Img) phrasingContent()    {}
func (e *Img) embeddedContent()    {}
func (e *Img) interactiveContent() {} // when usemap is set
func (e *Img) pictureContent()     {}
func (e *Img) figureContent()      {}

// ── <video> ───────────────────────────────────────────────────────────────────

type Video struct{ element }

func NewVideo() *Video { return &Video{newElement("video")} }

func (e *Video) node()            {}
func (e *Video) flowContent()     {}
func (e *Video) phrasingContent() {}
func (e *Video) embeddedContent() {}
func (e *Video) figureContent()   {}

// AppendChild accepts <source>, <track> and fallback FlowContent.
func (e *Video) AppendChild(n MediaContent)   { e.appendChildNode(n) }
func (e *Video) AppendFallback(n FlowContent) { e.appendChildNode(n) }
func (e *Video) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <audio> ───────────────────────────────────────────────────────────────────

type Audio struct{ element }

func NewAudio() *Audio { return &Audio{newElement("audio")} }

func (e *Audio) node()            {}
func (e *Audio) flowContent()     {}
func (e *Audio) phrasingContent() {}
func (e *Audio) embeddedContent() {}
func (e *Audio) figureContent()   {}

func (e *Audio) AppendChild(n MediaContent)   { e.appendChildNode(n) }
func (e *Audio) AppendFallback(n FlowContent) { e.appendChildNode(n) }
func (e *Audio) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <source> (void) ───────────────────────────────────────────────────────────

type Source struct{ voidElement }

func NewSource() *Source { return &Source{newVoidElement("source")} }

func (e *Source) node()           {}
func (e *Source) mediaContent()   {}
func (e *Source) pictureContent() {}

// ── <track> (void) ────────────────────────────────────────────────────────────

type Track struct{ voidElement }

func NewTrack() *Track { return &Track{newVoidElement("track")} }

func (e *Track) node()         {}
func (e *Track) mediaContent() {}

// ── <picture> ─────────────────────────────────────────────────────────────────

type Picture struct{ element }

func NewPicture() *Picture { return &Picture{newElement("picture")} }

func (e *Picture) node()            {}
func (e *Picture) flowContent()     {}
func (e *Picture) phrasingContent() {}
func (e *Picture) embeddedContent() {}
func (e *Picture) figureContent()   {}

// AppendChild accepts <source> and <img>.
func (e *Picture) AppendChild(n PictureContent) { e.appendChildNode(n) }
func (e *Picture) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <canvas> ──────────────────────────────────────────────────────────────────

type Canvas struct{ element }

func NewCanvas() *Canvas { return &Canvas{newElement("canvas")} }

func (e *Canvas) node()            {}
func (e *Canvas) flowContent()     {}
func (e *Canvas) phrasingContent() {}
func (e *Canvas) embeddedContent() {}
func (e *Canvas) figureContent()   {}

// Canvas accepts transparent content (phrasing in phrasing context).
func (e *Canvas) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Canvas) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <iframe> ──────────────────────────────────────────────────────────────────

type Iframe struct{ element }

func NewIframe() *Iframe { return &Iframe{newElement("iframe")} }

func (e *Iframe) node()               {}
func (e *Iframe) flowContent()        {}
func (e *Iframe) phrasingContent()    {}
func (e *Iframe) embeddedContent()    {}
func (e *Iframe) interactiveContent() {}
func (e *Iframe) figureContent()      {}

func (e *Iframe) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <embed> (void) ────────────────────────────────────────────────────────────

type Embed struct{ voidElement }

func NewEmbed() *Embed { return &Embed{newVoidElement("embed")} }

func (e *Embed) node()               {}
func (e *Embed) flowContent()        {}
func (e *Embed) phrasingContent()    {}
func (e *Embed) embeddedContent()    {}
func (e *Embed) interactiveContent() {}
func (e *Embed) figureContent()      {}

// ── <object> ──────────────────────────────────────────────────────────────────

type Object struct{ element }

func NewObject() *Object { return &Object{newElement("object")} }

func (e *Object) node()               {}
func (e *Object) flowContent()        {}
func (e *Object) phrasingContent()    {}
func (e *Object) embeddedContent()    {}
func (e *Object) interactiveContent() {}
func (e *Object) figureContent()      {}

func (e *Object) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Object) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <map> ─────────────────────────────────────────────────────────────────────

type Map struct{ element }

func NewMap() *Map { return &Map{newElement("map")} }

func (e *Map) node()            {}
func (e *Map) flowContent()     {}
func (e *Map) phrasingContent() {}

func (e *Map) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Map) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <area> (void) ─────────────────────────────────────────────────────────────

type Area struct{ voidElement }

func NewArea() *Area { return &Area{newVoidElement("area")} }

func (e *Area) node()            {}
func (e *Area) flowContent()     {}
func (e *Area) phrasingContent() {}
