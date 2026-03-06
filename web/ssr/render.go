// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ssr

import (
	"fmt"
	"strconv"

	"go.wdy.de/nago/pkg/dom"
	"go.wdy.de/nago/presentation/proto"
)

// RenderComponent converts a proto.Component into a dom.FlowContent node.
// Returns nil for nil input or invisible components (Invisible flag set).
// Unknown / interactive-only components return an empty <div> placeholder.
func RenderComponent(c proto.Component) dom.FlowContent {
	if c == nil {
		return nil
	}

	switch v := c.(type) {
	// ── layout ────────────────────────────────────────────────────────────────
	case *proto.VStack:
		return renderVStack(v)
	case *proto.HStack:
		return renderHStack(v)
	case *proto.Box:
		return renderBox(v)
	case *proto.Spacer:
		return renderSpacer(v)

	// ── scaffold ──────────────────────────────────────────────────────────────
	case *proto.Scaffold:
		return renderScaffold(v)

	// ── scrolling / containers ────────────────────────────────────────────────
	case *proto.ScrollView:
		return renderScrollView(v)
	case *proto.Grid:
		return renderGrid(v)

	// ── content ───────────────────────────────────────────────────────────────
	case *proto.TextView:
		return renderTextView(v)
	case *proto.TextLayout:
		return renderTextLayout(v)
	case *proto.Img:
		return renderImg(v)
	case *proto.Divider:
		return renderDivider(v)
	case *proto.RichText:
		return renderRichText(v)
	case *proto.WindowTitle:
		// WindowTitle is a side-effect-only component – no DOM node emitted.
		return nil

	// ── table ─────────────────────────────────────────────────────────────────
	case *proto.Table:
		return renderTable(v)

	// ── interactive / placeholder stubs ───────────────────────────────────────
	case *proto.TextField:
		return renderTextField(v)
	case *proto.PasswordField:
		return renderPasswordField(v)
	case *proto.Checkbox:
		return renderCheckbox(v)
	case *proto.Toggle:
		return renderToggle(v)
	case *proto.Radiobutton:
		return renderRadioButton(v)
	case *proto.DatePicker:
		return renderDatePicker(v)
	case *proto.Modal:
		return renderModal(v)
	case *proto.WebView:
		return renderWebView(v)
	case *proto.Select:
		return renderSelect(v)
	case *proto.DnDArea:
		return renderDnDArea(v)

	default:
		// Any unknown component type returns an empty placeholder <div>.
		d := dom.NewDiv()
		d.SetAttr("data-ssr-unknown", fmt.Sprintf("%T", c))
		return d
	}
}

// renderChildren converts a proto.Components slice into dom children of parent.
func renderChildren(parent *dom.Div, children proto.Components) {
	for _, child := range children {
		node := RenderComponent(child)
		if node != nil {
			parent.AppendChild(node)
		}
	}
}

// ── VStack ────────────────────────────────────────────────────────────────────

func renderVStack(v *proto.VStack) dom.FlowContent {
	if bool(v.Invisible) {
		return nil
	}

	d := dom.NewDiv()

	styles := []string{
		"display:flex",
		"flex-direction:column",
	}

	if v.Gap != "" {
		styles = append(styles, "gap:"+LengthCSS(v.Gap))
	}

	// alignment: main axis = justify-content (vertical), cross axis = align-items (horizontal)
	styles = append(styles,
		"justify-content:"+AlignmentMainAxisCSS(v.Alignment),
		"align-items:"+AlignmentCrossAxisCSS(v.Alignment),
	)

	if v.BackgroundColor != "" {
		styles = append(styles, "background-color:"+string(v.BackgroundColor))
	}
	if v.TextColor != "" {
		styles = append(styles, "color:"+string(v.TextColor))
	}

	styles = append(styles, FrameCSS(v.Frame)...)
	styles = append(styles, PaddingCSS(v.Padding)...)
	styles = append(styles, BorderCSS(v.Border)...)

	if v.Id != "" {
		d.SetAttr("id", string(v.Id))
	}

	d.SetAttr("style", JoinCSS(styles))

	renderChildren(d, v.Children)
	return d
}

// ── HStack ────────────────────────────────────────────────────────────────────

func renderHStack(v *proto.HStack) dom.FlowContent {
	if bool(v.Invisible) {
		return nil
	}

	styles := []string{
		"display:inline-flex",
		"flex-direction:row",
	}

	if bool(v.Wrap) {
		styles = append(styles, "flex-wrap:wrap")
	}
	if v.Gap != "" {
		styles = append(styles, "gap:"+LengthCSS(v.Gap))
	}

	// For HStack: main axis = justify-content (horizontal), cross axis = align-items (vertical).
	styles = append(styles,
		"justify-content:"+AlignmentMainAxisCSS(v.Alignment),
		"align-items:"+AlignmentCrossAxisCSS(v.Alignment),
	)

	if v.BackgroundColor != "" {
		styles = append(styles, "background-color:"+string(v.BackgroundColor))
	}
	if v.TextColor != "" {
		styles = append(styles, "color:"+string(v.TextColor))
	}

	styles = append(styles, FrameCSS(v.Frame)...)
	styles = append(styles, PaddingCSS(v.Padding)...)
	styles = append(styles, BorderCSS(v.Border)...)

	// ── Case 1: HStack with URL → <a href target> ─────────────────────────────
	// Mirrors: v-else-if="props.ui instanceof HStack && props.ui.url"
	if v.Url != "" {
		a := dom.NewA()
		a.SetAttr("href", string(v.Url))
		if v.Target != "" {
			a.SetAttr("target", string(v.Target))
		}
		if v.Id != "" {
			a.SetAttr("id", string(v.Id))
		}
		if v.AccessibilityLabel != "" {
			a.SetAttr("title", string(v.AccessibilityLabel))
		}
		a.SetAttr("style", JoinCSS(styles))
		renderChildrenA(a, v.Children)
		return a
	}

	// ── Case 2: stylePreset set → <button> ───────────────────────────────────
	// Mirrors: v-else-if stylePreset !== StyleNone && stylePreset !== undefined
	if v.StylePreset != proto.StyleNone {
		btn := dom.NewButton()
		if v.Id != "" {
			btn.SetAttr("id", string(v.Id))
		}
		if bool(v.Disabled) {
			btn.SetAttr("disabled", "")
		}
		if v.AccessibilityLabel != "" {
			btn.SetAttr("title", string(v.AccessibilityLabel))
		}
		btn.SetAttr("style", JoinCSS(styles))
		renderChildrenButton(btn, v.Children)
		return btn
	}

	// ── Case 3: plain <div> ───────────────────────────────────────────────────
	d := dom.NewDiv()
	if v.Id != "" {
		d.SetAttr("id", string(v.Id))
	}
	if v.AccessibilityLabel != "" {
		d.SetAttr("title", string(v.AccessibilityLabel))
	}
	d.SetAttr("style", JoinCSS(styles))
	renderChildren(d, v.Children)
	return d
}

// renderChildrenA appends rendered children into an <a> element.
func renderChildrenA(a *dom.A, children proto.Components) {
	for _, child := range children {
		node := RenderComponent(child)
		if node != nil {
			a.AppendFlow(node)
		}
	}
}

// renderChildrenButton appends rendered children into a <button> element.
func renderChildrenButton(btn *dom.Button, children proto.Components) {
	for _, child := range children {
		node := RenderComponent(child)
		if node != nil {
			btn.AppendFlow(node)
		}
	}
}

// ── Spacer ────────────────────────────────────────────────────────────────────

func renderBox(v *proto.Box) dom.FlowContent {
	d := dom.NewDiv()
	styles := []string{"position:relative"}
	if v.BackgroundColor != "" {
		styles = append(styles, "background-color:"+string(v.BackgroundColor))
	}
	styles = append(styles, FrameCSS(v.Frame)...)
	styles = append(styles, PaddingCSS(v.Padding)...)
	styles = append(styles, BorderCSS(v.Border)...)
	d.SetAttr("style", JoinCSS(styles))

	for _, ac := range v.Children {
		node := RenderComponent(ac.Component)
		if node != nil {
			d.AppendChild(node)
		}
	}
	return d
}

// ── Spacer ────────────────────────────────────────────────────────────────────

func renderSpacer(v *proto.Spacer) dom.FlowContent {
	d := dom.NewDiv()
	styles := []string{"flex-grow:1"}
	styles = append(styles, FrameCSS(v.Frame)...)
	if v.BackgroundColor != "" {
		styles = append(styles, "background-color:"+string(v.BackgroundColor))
	}
	d.SetAttr("style", JoinCSS(styles))
	return d
}

// ── Scaffold ──────────────────────────────────────────────────────────────────

func renderScaffold(v *proto.Scaffold) dom.FlowContent {
	// Outer wrapper: full-height flex row
	outer := dom.NewDiv()
	outerStyles := []string{"display:flex", "flex-direction:row"}
	if v.Height != "" {
		outerStyles = append(outerStyles, "height:"+LengthCSS(v.Height))
	} else {
		outerStyles = append(outerStyles, "height:100dvh")
	}
	outer.SetAttr("style", JoinCSS(outerStyles))

	// Sidebar / nav
	nav := dom.NewNav()
	nav.SetAttr("style", "display:flex;flex-direction:column;overflow-y:auto")

	// Logo
	if v.Logo != nil {
		logoNode := RenderComponent(v.Logo)
		if logoNode != nil {
			nav.AppendChild(logoNode)
		}
	}

	// Menu entries
	for _, entry := range v.Menu {
		renderMenuEntry(nav, entry)
	}

	// Bottom view
	if v.BottomView != nil {
		bottomNode := RenderComponent(v.BottomView)
		if bottomNode != nil {
			// spacer before bottom view
			spacer := dom.NewDiv()
			spacer.SetAttr("style", "flex-grow:1")
			nav.AppendChild(spacer)
			nav.AppendChild(bottomNode)
		}
	}

	outer.AppendChild(nav)

	// Right side: body + footer
	right := dom.NewDiv()
	right.SetAttr("style", "flex:1;display:flex;flex-direction:column;overflow:auto")

	if v.Body != nil {
		bodyNode := RenderComponent(v.Body)
		if bodyNode != nil {
			bodyWrapper := dom.NewDiv()
			bodyWrapper.SetAttr("style", "flex:1")
			bodyWrapper.AppendChild(bodyNode)
			right.AppendChild(bodyWrapper)
		}
	}

	if v.Footer != nil {
		footerNode := RenderComponent(v.Footer)
		if footerNode != nil {
			footer := dom.NewFooter()
			footer.AppendChild(footerNode)
			right.AppendChild(footer)
		}
	}

	outer.AppendChild(right)
	return outer
}

func renderMenuEntry(nav *dom.Nav, entry proto.ScaffoldMenuEntry) {
	title := string(entry.Title)
	href := ""
	if entry.RootView != "" {
		href = "/" + string(entry.RootView)
	}

	a := dom.NewA()
	if href != "" {
		a.SetAttr("href", href)
	}
	a.SetAttr("style", "display:flex;align-items:center;gap:8px;padding:8px;text-decoration:none")

	// Icon
	if entry.Icon != nil {
		iconNode := RenderComponent(entry.Icon)
		if iconNode != nil {
			iconWrapper := dom.NewSpan()
			// span accepts phrasing; use generic inner
			a.AppendChild(iconWrapper)
		}
	}

	// Title text
	if title != "" {
		span := dom.NewSpan()
		span.SetTextContent(title)
		a.AppendChild(span)
	}

	// Badge
	if entry.Badge != "" {
		badge := dom.NewSpan()
		badge.SetTextContent(string(entry.Badge))
		badge.SetAttr("style", "background:red;color:white;border-radius:9999px;padding:1px 6px;font-size:0.75em")
		a.AppendChild(badge)
	}

	nav.AppendChild(a)

	// Sub-menu entries (recursive)
	if bool(entry.Expanded) {
		for _, sub := range entry.Menu {
			subNav := dom.NewNav()
			subNav.SetAttr("style", "padding-left:16px")
			renderMenuEntry(subNav, sub)
			nav.AppendChild(subNav)
		}
	}
}

// ── ScrollView ────────────────────────────────────────────────────────────────

func renderScrollView(v *proto.ScrollView) dom.FlowContent {
	if bool(v.Invisible) {
		return nil
	}

	d := dom.NewDiv()
	styles := []string{"overflow:auto"}
	styles = append(styles, FrameCSS(v.Frame)...)
	styles = append(styles, PaddingCSS(v.Padding)...)
	styles = append(styles, BorderCSS(v.Border)...)
	if v.BackgroundColor != "" {
		styles = append(styles, "background-color:"+string(v.BackgroundColor))
	}
	d.SetAttr("style", JoinCSS(styles))

	if v.Content != nil {
		node := RenderComponent(v.Content)
		if node != nil {
			d.AppendChild(node)
		}
	}
	return d
}

// ── Grid ──────────────────────────────────────────────────────────────────────

func renderGrid(v *proto.Grid) dom.FlowContent {
	if bool(v.Invisible) {
		return nil
	}

	d := dom.NewDiv()
	cols := int(v.Columns)
	if cols <= 0 {
		cols = 1
	}

	styles := []string{
		"display:grid",
		fmt.Sprintf("grid-template-columns:repeat(%d,1fr)", cols),
	}
	if v.ColGap != "" {
		styles = append(styles, "column-gap:"+LengthCSS(v.ColGap))
	}
	if v.RowGap != "" {
		styles = append(styles, "row-gap:"+LengthCSS(v.RowGap))
	}
	if v.BackgroundColor != "" {
		styles = append(styles, "background-color:"+string(v.BackgroundColor))
	}
	styles = append(styles, FrameCSS(v.Frame)...)
	styles = append(styles, PaddingCSS(v.Padding)...)
	styles = append(styles, BorderCSS(v.Border)...)
	d.SetAttr("style", JoinCSS(styles))

	for _, cell := range v.Cells {
		cellDiv := dom.NewDiv()
		node := RenderComponent(cell.Body)
		if node != nil {
			cellDiv.AppendChild(node)
		}
		d.AppendChild(cellDiv)
	}
	return d
}

// ── TextView ──────────────────────────────────────────────────────────────────

func renderTextView(v *proto.TextView) dom.FlowContent {
	if bool(v.Invisible) {
		return nil
	}

	span := dom.NewSpan()
	span.SetTextContent(string(v.Value))

	var styles []string
	if v.Color != "" {
		styles = append(styles, "color:"+string(v.Color))
	}
	if v.BackgroundColor != "" {
		styles = append(styles, "background-color:"+string(v.BackgroundColor))
	}
	if bool(v.Underline) {
		styles = append(styles, "text-decoration:underline")
	}
	if bool(v.LineBreak) {
		styles = append(styles, "white-space:pre-wrap")
	}
	if v.Hyphens != "" {
		styles = append(styles, "hyphens:"+string(v.Hyphens))
	}
	if v.WordBreak != "" {
		styles = append(styles, "word-break:"+string(v.WordBreak))
	}

	// font
	styles = append(styles, fontCSS(v.Font)...)

	// text alignment
	if v.TextAlignment != 0 {
		styles = append(styles, "text-align:"+textAlignCSS(v.TextAlignment))
	}

	styles = append(styles, FrameCSS(v.Frame)...)
	styles = append(styles, PaddingCSS(v.Padding)...)
	styles = append(styles, BorderCSS(v.Border)...)

	if len(styles) > 0 {
		span.SetAttr("style", JoinCSS(styles))
	}
	if v.LabelFor != "" {
		span.SetAttr("for", string(v.LabelFor))
	}

	// Wrap in a div so we can return FlowContent
	d := dom.NewDiv()
	d.AppendChild(span)
	return d
}

// fontCSS converts a proto.Font into CSS declarations.
func fontCSS(f proto.Font) []string {
	var s []string
	if f.Size != "" {
		s = append(s, "font-size:"+LengthCSS(f.Size))
	}
	if f.Name != "" {
		s = append(s, "font-family:"+string(f.Name))
	}
	if f.Weight == proto.BoldFontWeight {
		s = append(s, "font-weight:bold")
	} else if f.Weight != 0 {
		s = append(s, fmt.Sprintf("font-weight:%d", uint64(f.Weight)))
	}
	if f.Style == proto.Italic {
		s = append(s, "font-style:italic")
	}
	return s
}

// textAlignCSS maps proto.TextAlignment to CSS text-align.
func textAlignCSS(a proto.TextAlignment) string {
	switch a {
	case proto.TextAlignStart:
		return "start"
	case proto.TextAlignEnd:
		return "end"
	case proto.TextAlignCenter:
		return "center"
	case proto.TextAlignJustify:
		return "justify"
	default:
		return "start"
	}
}

// ── TextLayout ────────────────────────────────────────────────────────────────

func renderTextLayout(v *proto.TextLayout) dom.FlowContent {
	if bool(v.Invisible) {
		return nil
	}

	d := dom.NewDiv()
	var styles []string
	if v.BackgroundColor != "" {
		styles = append(styles, "background-color:"+string(v.BackgroundColor))
	}
	styles = append(styles, FrameCSS(v.Frame)...)
	styles = append(styles, PaddingCSS(v.Padding)...)
	styles = append(styles, BorderCSS(v.Border)...)
	if v.TextAlignment != 0 {
		styles = append(styles, "text-align:"+textAlignCSS(v.TextAlignment))
	}
	if len(styles) > 0 {
		d.SetAttr("style", JoinCSS(styles))
	}
	renderChildren(d, v.Children)
	return d
}

// ── Img ───────────────────────────────────────────────────────────────────────

func renderImg(v *proto.Img) dom.FlowContent {
	if bool(v.Invisible) {
		return nil
	}

	img := dom.NewImg()
	if v.Uri != "" {
		img.SetAttr("src", string(v.Uri))
	}
	if v.AccessibilityLabel != "" {
		img.SetAttr("alt", string(v.AccessibilityLabel))
	} else {
		img.SetAttr("alt", "")
	}

	var styles []string
	styles = append(styles, FrameCSS(v.Frame)...)
	styles = append(styles, PaddingCSS(v.Padding)...)
	styles = append(styles, BorderCSS(v.Border)...)

	switch v.ObjectFit {
	case proto.Fill:
		styles = append(styles, "object-fit:fill")
	case proto.Contain:
		styles = append(styles, "object-fit:contain")
	case proto.Cover:
		styles = append(styles, "object-fit:cover")
	}

	if len(styles) > 0 {
		img.SetAttr("style", JoinCSS(styles))
	}

	// img is void – wrap in a div to return FlowContent
	d := dom.NewDiv()
	d.AppendChild(img)
	return d
}

// ── Divider ───────────────────────────────────────────────────────────────────

func renderDivider(v *proto.Divider) dom.FlowContent {
	hr := dom.NewHr()
	var styles []string
	styles = append(styles, FrameCSS(v.Frame)...)
	styles = append(styles, BorderCSS(v.Border)...)
	// Divider padding is treated as margin (same as UiDivider.vue with marginCSS)
	if v.Padding.Top != "" {
		styles = append(styles, "margin-top:"+LengthCSS(v.Padding.Top))
	}
	if v.Padding.Bottom != "" {
		styles = append(styles, "margin-bottom:"+LengthCSS(v.Padding.Bottom))
	}
	if v.Padding.Left != "" {
		styles = append(styles, "margin-left:"+LengthCSS(v.Padding.Left))
	}
	if v.Padding.Right != "" {
		styles = append(styles, "margin-right:"+LengthCSS(v.Padding.Right))
	}
	if len(styles) > 0 {
		hr.SetAttr("style", JoinCSS(styles))
	}

	d := dom.NewDiv()
	d.AppendChild(hr)
	return d
}

// ── RichText ──────────────────────────────────────────────────────────────────

func renderRichText(v *proto.RichText) dom.FlowContent {
	d := dom.NewDiv()
	var styles []string
	styles = append(styles, FrameCSS(v.Frame)...)
	if len(styles) > 0 {
		d.SetAttr("style", JoinCSS(styles))
	}
	// Use SetInnerHTML so that the HTML content is parsed by x/net/html (see pkg/dom/element.go).
	d.SetInnerHTML(string(v.Value))
	return d
}

// ── Table ─────────────────────────────────────────────────────────────────────

func renderTable(v *proto.Table) dom.FlowContent {
	table := dom.NewTable()

	var tableStyles []string
	tableStyles = append(tableStyles, FrameCSS(v.Frame)...)
	if v.BackgroundColor != "" {
		tableStyles = append(tableStyles, "background-color:"+string(v.BackgroundColor))
	}
	tableStyles = append(tableStyles, BorderCSS(v.Border)...)
	tableStyles = append(tableStyles, "width:100%", "text-align:left", "overflow:clip", "word-break:break-all")
	table.SetAttr("style", JoinCSS(tableStyles))

	// thead
	if len(v.Header.Columns) > 0 {
		thead := dom.NewThead()
		tr := dom.NewTr()

		// header divider
		headRowStyles := []string{}
		if v.HeaderDividerColor != "" {
			headRowStyles = append(headRowStyles,
				"border-collapse:collapse",
				"border-bottom-width:2px",
				"border-style:solid",
				"border-color:"+string(v.HeaderDividerColor),
			)
		} else if v.RowDividerColor != "" {
			headRowStyles = append(headRowStyles,
				"border-collapse:collapse",
				"border-bottom-width:2px",
				"border-style:solid",
				"border-color:"+string(v.RowDividerColor),
			)
		}
		if len(headRowStyles) > 0 {
			tr.SetAttr("style", JoinCSS(headRowStyles))
		}

		for colIdx, col := range v.Header.Columns {
			th := dom.NewTh()
			th.SetAttr("scope", "col")

			var thStyles []string
			thStyles = append(thStyles, "font-weight:normal")
			if col.Width != "" {
				thStyles = append(thStyles, "width:"+LengthCSS(col.Width))
			}
			if col.CellBackgroundColor != "" {
				thStyles = append(thStyles, "background-color:"+string(col.CellBackgroundColor))
			}
			thStyles = append(thStyles, TableAlignmentCSS(col.Alignment, true)...)

			// cell padding: use column-specific or table default
			if !col.CellPadding.IsZero() {
				thStyles = append(thStyles, PaddingCSS(col.CellPadding)...)
			} else if !v.DefaultCellPadding.IsZero() {
				thStyles = append(thStyles, PaddingCSS(v.DefaultCellPadding)...)
			}
			thStyles = append(thStyles, BorderCSS(col.CellBorder)...)

			th.SetAttr("style", JoinCSS(thStyles))
			_ = colIdx

			if col.Content != nil {
				node := RenderComponent(col.Content)
				if node != nil {
					th.AppendChild(node)
				}
			}
			tr.AppendChild(th)
		}
		thead.AppendChild(tr)
		table.AppendChild(thead)
	}

	// tbody
	tbody := dom.NewTbody()
	for rowIdx, row := range v.Rows {
		tr := dom.NewTr()

		var rowStyles []string
		if row.BackgroundColor != "" {
			rowStyles = append(rowStyles, "background-color:"+string(row.BackgroundColor))
		}
		if row.Height != "" {
			rowStyles = append(rowStyles, "height:"+LengthCSS(row.Height))
		}
		if rowIdx > 0 && v.RowDividerColor != "" {
			rowStyles = append(rowStyles,
				"border-collapse:collapse",
				"border-top-width:1px",
				"border-style:solid",
				"border-color:"+string(v.RowDividerColor),
			)
		}
		if len(rowStyles) > 0 {
			tr.SetAttr("style", JoinCSS(rowStyles))
		}

		for _, cell := range row.Cells {
			td := dom.NewTd()

			if uint64(cell.RowSpan) > 0 {
				td.SetAttr("rowspan", strconv.FormatUint(uint64(cell.RowSpan), 10))
			}
			if uint64(cell.ColSpan) > 0 {
				td.SetAttr("colspan", strconv.FormatUint(uint64(cell.ColSpan), 10))
			}

			var tdStyles []string
			if cell.BackgroundColor != "" {
				tdStyles = append(tdStyles, "background-color:"+string(cell.BackgroundColor))
			}
			tdStyles = append(tdStyles, TableAlignmentCSS(cell.Alignment, false)...)

			if !cell.Padding.IsZero() {
				tdStyles = append(tdStyles, PaddingCSS(cell.Padding)...)
			} else if !v.DefaultCellPadding.IsZero() {
				tdStyles = append(tdStyles, PaddingCSS(v.DefaultCellPadding)...)
			}
			tdStyles = append(tdStyles, BorderCSS(cell.Border)...)
			td.SetAttr("style", JoinCSS(tdStyles))

			if cell.Content != nil {
				node := RenderComponent(cell.Content)
				if node != nil {
					td.AppendChild(node)
				}
			}
			tr.AppendChild(td)
		}
		tbody.AppendChild(tr)
	}
	table.AppendChild(tbody)

	// Wrap in a div so we satisfy FlowContent
	d := dom.NewDiv()
	d.AppendChild(table)
	return d
}

// ── Shared input wrapper helper ───────────────────────────────────────────────

// renderInputWrapper wraps an input element together with a label, supporting
// text and error message – analogous to InputWrapper.vue.
// The returned *dom.Div contains (in order):
//  1. optional <label for=id> with labelText
//  2. the inputNode itself
//  3. optional <p> with errorText (styled red) or supportingText (styled muted)
func renderInputWrapper(id, labelText, errorText, supportingText string, inputNode dom.FlowContent) *dom.Div {
	wrapper := dom.NewDiv()
	wrapper.SetAttr("style", "display:flex;flex-direction:column;gap:2px")

	if labelText != "" {
		lbl := dom.NewLabel()
		if id != "" {
			lbl.SetAttr("for", id)
		}
		lbl.SetAttr("style", "font-size:0.875rem")
		lbl.AppendChild(dom.NewTextNode(labelText))
		wrapper.AppendChild(lbl)
	}

	if inputNode != nil {
		wrapper.AppendChild(inputNode)
	}

	if errorText != "" {
		p := dom.NewP()
		p.SetAttr("style", "font-size:0.75rem;color:var(--SE0,red)")
		p.AppendChild(dom.NewTextNode(errorText))
		wrapper.AppendChild(p)
	} else if supportingText != "" {
		p := dom.NewP()
		p.SetAttr("style", "font-size:0.75rem;opacity:0.6")
		p.AppendChild(dom.NewTextNode(supportingText))
		wrapper.AppendChild(p)
	}

	return wrapper
}

// ── TextField ─────────────────────────────────────────────────────────────────

func renderTextField(v *proto.TextField) dom.FlowContent {
	if bool(v.Invisible) {
		return nil
	}

	id := string(v.Id)

	var inputNode dom.FlowContent
	if uint64(v.Lines) > 0 {
		ta := dom.NewTextarea()
		if id != "" {
			ta.SetAttr("id", id)
		}
		if bool(v.Disabled) {
			ta.SetAttr("disabled", "")
		}
		ta.SetAttr("rows", strconv.FormatUint(uint64(v.Lines), 10))
		ta.SetAttr("style", "width:100%;box-sizing:border-box")
		if v.Value != "" {
			ta.SetTextContent(string(v.Value))
		}
		inputNode = ta
	} else {
		inp := dom.NewInput()
		inp.SetAttr("type", "text")
		if id != "" {
			inp.SetAttr("id", id)
		}
		if v.Value != "" {
			inp.SetAttr("value", string(v.Value))
		}
		if bool(v.Disabled) {
			inp.SetAttr("disabled", "")
		}
		inp.SetAttr("style", "width:100%;box-sizing:border-box")
		inputNode = inp
	}

	outer := dom.NewDiv()
	frameStyles := FrameCSS(v.Frame)
	if len(frameStyles) > 0 {
		outer.SetAttr("style", JoinCSS(frameStyles))
	}
	outer.AppendChild(renderInputWrapper(id, string(v.Label), string(v.ErrorText), string(v.SupportingText), inputNode))
	return outer
}

// ── PasswordField ─────────────────────────────────────────────────────────────

func renderPasswordField(v *proto.PasswordField) dom.FlowContent {
	if bool(v.Invisible) {
		return nil
	}

	id := string(v.Id)

	inp := dom.NewInput()
	if bool(v.Revealed) {
		inp.SetAttr("type", "text")
	} else {
		inp.SetAttr("type", "password")
	}
	if id != "" {
		inp.SetAttr("id", id)
	}
	if v.Value != "" {
		inp.SetAttr("value", string(v.Value))
	}
	if bool(v.Disabled) {
		inp.SetAttr("disabled", "")
	}
	if bool(v.DisableAutocomplete) {
		inp.SetAttr("autocomplete", "off")
	}
	inp.SetAttr("style", "width:100%;box-sizing:border-box")

	outer := dom.NewDiv()
	frameStyles := FrameCSS(v.Frame)
	if len(frameStyles) > 0 {
		outer.SetAttr("style", JoinCSS(frameStyles))
	}
	outer.AppendChild(renderInputWrapper(id, string(v.Label), string(v.ErrorText), string(v.SupportingText), inp))
	return outer
}

// ── Checkbox ──────────────────────────────────────────────────────────────────

func renderCheckbox(v *proto.Checkbox) dom.FlowContent {
	if bool(v.Invisible) {
		return nil
	}

	inp := dom.NewInput()
	inp.SetAttr("type", "checkbox")
	if v.Id != "" {
		inp.SetAttr("id", string(v.Id))
	}
	if bool(v.Value) {
		inp.SetAttr("checked", "")
	}
	if bool(v.Disabled) {
		inp.SetAttr("disabled", "")
	}

	d := dom.NewDiv()
	d.AppendChild(inp)
	return d
}

// ── Toggle ────────────────────────────────────────────────────────────────────

func renderToggle(v *proto.Toggle) dom.FlowContent {
	if bool(v.Invisible) {
		return nil
	}

	inp := dom.NewInput()
	inp.SetAttr("type", "checkbox")
	inp.SetAttr("role", "switch")
	if bool(v.Value) {
		inp.SetAttr("checked", "")
	}
	if bool(v.Disabled) {
		inp.SetAttr("disabled", "")
	}

	d := dom.NewDiv()
	d.AppendChild(inp)
	return d
}

// ── Radiobutton ───────────────────────────────────────────────────────────────

func renderRadioButton(v *proto.Radiobutton) dom.FlowContent {
	if bool(v.Invisible) {
		return nil
	}

	inp := dom.NewInput()
	inp.SetAttr("type", "radio")
	if v.Id != "" {
		inp.SetAttr("id", string(v.Id))
	}
	if bool(v.Value) {
		inp.SetAttr("checked", "")
	}
	if bool(v.Disabled) {
		inp.SetAttr("disabled", "")
	}

	d := dom.NewDiv()
	d.AppendChild(inp)
	return d
}

// ── Select ────────────────────────────────────────────────────────────────────

func renderSelect(v *proto.Select) dom.FlowContent {
	id := string(v.Id)

	sel := dom.NewSelect()
	if id != "" {
		sel.SetAttr("id", id)
	}
	if bool(v.Disabled) {
		sel.SetAttr("disabled", "")
	}
	sel.SetAttr("style", "width:100%;box-sizing:border-box")

	for _, opt := range v.Options {
		o := dom.NewOption(string(opt.Label))
		o.SetAttr("value", string(opt.Value))
		if bool(opt.Disabled) {
			o.SetAttr("disabled", "")
		}
		if string(opt.Value) == string(v.Value) {
			o.SetAttr("selected", "")
		}
		sel.AppendChild(o)
	}

	outer := dom.NewDiv()
	frameStyles := FrameCSS(v.Frame)
	if len(frameStyles) > 0 {
		outer.SetAttr("style", JoinCSS(frameStyles))
	}
	outer.AppendChild(renderInputWrapper(id, string(v.Label), string(v.ErrorText), string(v.SupportingText), sel))
	return outer
}

// ── DatePicker ────────────────────────────────────────────────────────────────

func renderDatePicker(v *proto.DatePicker) dom.FlowContent {
	if bool(v.Invisible) {
		return nil
	}

	inp := dom.NewInput()
	inp.SetAttr("type", "date")
	if bool(v.Disabled) {
		inp.SetAttr("disabled", "")
	}
	inp.SetAttr("style", "width:100%;box-sizing:border-box")

	// Format DateData as YYYY-MM-DD for the HTML date input value attribute.
	if y := uint64(v.Value.Year); y > 0 {
		dateStr := fmt.Sprintf("%04d-%02d-%02d",
			uint64(v.Value.Year),
			uint64(v.Value.Month),
			uint64(v.Value.Day),
		)
		inp.SetAttr("value", dateStr)
	}

	outer := dom.NewDiv()
	frameStyles := FrameCSS(v.Frame)
	if len(frameStyles) > 0 {
		outer.SetAttr("style", JoinCSS(frameStyles))
	}
	outer.AppendChild(renderInputWrapper("", string(v.Label), string(v.ErrorText), string(v.SupportingText), inp))
	return outer
}

// ── Modal ─────────────────────────────────────────────────────────────────────

// renderModal renders the modal content as a <div> with position:fixed styling
// so that SSR at least includes the content in the DOM. Vue takes over the
// interactive overlay behaviour on the client.
func renderModal(v *proto.Modal) dom.FlowContent {
	d := dom.NewDiv()

	styles := []string{
		"position:fixed",
		"inset:0",
		"display:flex",
		"align-items:center",
		"justify-content:center",
		"z-index:1000",
	}

	// Honour explicit position offsets when provided.
	if v.Top != "" {
		styles = append(styles, "top:"+LengthCSS(v.Top))
	}
	if v.Bottom != "" {
		styles = append(styles, "bottom:"+LengthCSS(v.Bottom))
	}
	if v.Left != "" {
		styles = append(styles, "left:"+LengthCSS(v.Left))
	}
	if v.Right != "" {
		styles = append(styles, "right:"+LengthCSS(v.Right))
	}

	d.SetAttr("style", JoinCSS(styles))
	d.SetAttr("role", "dialog")
	d.SetAttr("aria-modal", "true")

	if v.Content != nil {
		node := RenderComponent(v.Content)
		if node != nil {
			d.AppendChild(node)
		}
	}
	return d
}

// ── WebView ───────────────────────────────────────────────────────────────────

// renderWebView renders an <iframe> for SSR so crawlers and the initial paint
// can include the embedded resource. Interactive behaviour (sandboxing etc.)
// is handled by the Vue component on the client.
func renderWebView(v *proto.WebView) dom.FlowContent {
	iframe := dom.NewIframe()

	if v.URI != "" {
		iframe.SetAttr("src", string(v.URI))
	}
	if v.Title != "" {
		iframe.SetAttr("title", string(v.Title))
	}
	if v.Allow != "" {
		iframe.SetAttr("allow", string(v.Allow))
	}
	if v.ReferrerPolicy != "" {
		iframe.SetAttr("referrerpolicy", string(v.ReferrerPolicy))
	}

	var styles []string
	styles = append(styles, "border:none", "width:100%", "height:100%")
	styles = append(styles, FrameCSS(v.Frame)...)
	iframe.SetAttr("style", JoinCSS(styles))

	d := dom.NewDiv()
	d.AppendChild(iframe)
	return d
}

func renderDnDArea(v *proto.DnDArea) dom.FlowContent {
	d := dom.NewDiv()
	d.SetAttr("data-ssr", "dnd-area")
	renderChildren(d, v.Children)
	return d
}
