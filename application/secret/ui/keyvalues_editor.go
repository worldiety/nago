// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisecret

import (
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"

	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dropdown"
	"go.wdy.de/nago/presentation/ui/tags"
)

// keyRule is the hint for a key, as long as there is none to tell what it does.
const keyRule = "Buchstabe zuerst, dann Buchstaben, Ziffern, _ . -"

// maxLogoUpload lets a file above the limit arrive, because the upload of a larger one fails without any message and
// [readLogo] can tell the user why it is refused.
const maxLogoUpload = 16 * secret.MaxLogoSVGBytes

// TKeyValuesEditor is a composite component (Key Values Editor). It edits a [secret.KeyValues]: the sections, an input
// for each field by its kind and protection, a menu to protect, move and remove each field, a dialog for its key and
// the logo. It depends only on the bound state, so an application with a secret management of its own can use it, too.
type TKeyValuesEditor struct {
	state       *core.State[secret.KeyValues]
	stored      map[string]bool
	storedBase  string
	disabled    bool
	frame       ui.Frame
	replacing   *core.State[[]string] // the identifiers of the fields, whose value is being replaced or edited
	editing     *core.State[string]   // the identifier of the section or field in the dialog
	editingOpen *core.State[bool]
}

// KeyValuesEditor creates an editor of the state. Initialize the state with [secret.WithoutWriteOnly], because a
// password field sends its value to the browser, and see also [TKeyValuesEditor.Stored].
func KeyValuesEditor(state *core.State[secret.KeyValues]) TKeyValuesEditor {
	return TKeyValuesEditor{state: state, frame: ui.Frame{}.FullWidth()}
}

// Stored tells the stored secret, so that a write-only field can show that a value is stored and that it needs to be
// entered again after the address has changed. None of its values reaches the browser.
func (c TKeyValuesEditor) Stored(stored secret.KeyValues) TKeyValuesEditor {
	c.stored = stored.WriteOnlyFields()
	c.storedBase = stored.BaseURL()
	return c
}

// Disabled shows the inputs without allowing any change.
func (c TKeyValuesEditor) Disabled(disabled bool) TKeyValuesEditor {
	c.disabled = disabled
	return c
}

// Frame sets the layout frame, which is the full width by default.
func (c TKeyValuesEditor) Frame(frame ui.Frame) TKeyValuesEditor {
	c.frame = frame
	return c
}

// Render builds the editor.
func (c TKeyValuesEditor) Render(ctx core.RenderContext) core.RenderNode {
	kv := c.state.Get().WithIDs()
	c.state.Set(kv) // only changes a model which has never been saved, because the inputs are bound by the identifiers
	c.replacing = core.DerivedState[[]string](c.state, ".replacing")
	c.editing = core.DerivedState[string](c.state, ".editing")
	c.editingOpen = core.DerivedState[bool](c.state, ".editing.open")

	views := []core.View{c.header(ctx.Window(), kv)}
	for s := range kv.Sections {
		views = append(views, c.sectionView(kv, s))
	}

	switch {
	case len(kv.Sections) == 0:
		views = append(views, c.emptyState(kv))
	case !c.disabled:
		views = append(views, ui.TertiaryButton(c.addSection).PreIcon(heroOutline.Plus).Title("Abschnitt hinzufügen"))
	}

	// the dialog comes last, because an input keeps its place only as long as the views before it do
	views = append(views, c.dialog(kv))

	return ui.VStack(views...).Alignment(ui.Leading).Gap(ui.L24).Frame(c.frame).Render(ctx)
}

func (c TKeyValuesEditor) header(wnd core.Window, kv secret.KeyValues) core.View {
	return ui.HStack(
		c.logoView(wnd, kv),
		ui.VStack(
			c.textField(".name", "Name", kv.Name, "Programme finden die Verbindung unter diesem Namen.", func(x *secret.KeyValues, v string) {
				x.Name = v
			}),
			c.textField(".description", "Beschreibung", kv.Description, "", func(x *secret.KeyValues, v string) {
				x.Description = v
			}).Lines(2),
		).Alignment(ui.Leading).Gap(ui.L8).FullWidth(),
	).Alignment(ui.Top).Gap(ui.L16).FullWidth()
}

func (c TKeyValuesEditor) sectionView(kv secret.KeyValues, s int) core.View {
	section := kv.Sections[s]

	rows := make([]core.View, 0, len(section.Fields)+1)
	for f := range section.Fields {
		rows = append(rows, c.fieldRow(kv, s, f))
	}

	if !c.disabled {
		rows = append(rows, ui.HStack(
			c.addFieldMenu(kv, section.ID, ui.TertiaryButton(nil).PreIcon(heroOutline.Plus).Title("Feld hinzufügen")),
		).Alignment(ui.Leading).FullWidth())
	}

	var up, down func()
	if s > 0 {
		up = func() { c.moveSection(section.ID, -1) }
	}

	if s < len(kv.Sections)-1 {
		down = func() { c.moveSection(section.ID, 1) }
	}

	title := strings.TrimSpace("Abschnitt " + section.Label)

	return sectionBlock(sectionHeading(section, c.actionsMenu(title, section.ID, up, down, func() {
		c.removeSection(section.ID)
	})), rows)
}

func (c TKeyValuesEditor) emptyState(kv secret.KeyValues) core.View {
	if c.disabled {
		return smallMuted("Noch keine Felder")
	}

	return ui.VStack(
		muted(ui.ImageIcon(heroOutline.Key).Frame(ui.Frame{}.Size(ui.L32, ui.L32))),
		ui.Text("Noch keine Felder").Font(ui.TitleMedium),
		muted(ui.Text("Fügen Sie Adresse und Anmeldung Ihrer API hinzu.").Font(ui.BodyMedium)),
		c.addFieldMenu(kv, "", ui.SecondaryButton(nil).PreIcon(heroOutline.Plus).Title("Feld hinzufügen")),
	).Gap(ui.L8).
		BackgroundColor(ui.ColorCardBody).
		Border(ui.Border{}.Radius(ui.L16)).
		Padding(ui.Padding{}.All(ui.L32)).
		Frame(ui.Frame{}.FullWidth())
}

func (c TKeyValuesEditor) fieldRow(kv secret.KeyValues, s, f int) core.View {
	section := kv.Sections[s]
	field := section.Fields[f]

	var up, down func()
	if f > 0 {
		up = func() { c.moveField(field.ID, -1) }
	}

	if f < len(section.Fields)-1 {
		down = func() { c.moveField(field.ID, 1) }
	}

	menu := c.actionsMenu(fieldTitle(field), field.ID, up, down, func() {
		c.removeField(field.ID)
	})

	alignment := ui.Top
	if isToggle(field) {
		alignment = ui.Center
	} else if menu != nil {
		// an input has its label above, the menu belongs next to the input itself
		menu = ui.VStack(menu).Padding(ui.Padding{Top: ui.L(18)})
	}

	return ui.HStack(
		ui.VStack(
			c.valueInput(kv, field),
			c.metaLine(kv, field),
		).Alignment(ui.Leading).Gap(ui.L4).FullWidth(),
		menu,
	).Alignment(alignment).Gap(ui.L8).FullWidth().Padding(ui.Padding{}.Vertical(ui.L4))
}

// metaLine holds the state of a field, so that it needs no paragraph of help below its input.
func (c TKeyValuesEditor) metaLine(kv secret.KeyValues, field secret.KeyValueField) core.View {
	return ui.HStack(
		fieldUse(field),
		ui.If(c.losesStored(field), tags.StatusBadge(ui.SW0, "Gespeicherter Wert geht verloren")),
		ui.Spacer(),
		ui.If(c.isCancelable(kv, field), ui.TertiaryButton(func() {
			c.cancelReplace(field)
		}).Title(cancelTitle(field))),
		c.protectionChip(field),
	).Alignment(ui.Leading).Wrap(true).Gap(ui.L8).FullWidth()
}

func isToggle(field secret.KeyValueField) bool {
	return field.Kind == secret.FieldBool && field.Protection == secret.Visible
}

func (c TKeyValuesEditor) valueInput(kv secret.KeyValues, field secret.KeyValueField) core.View {
	if isToggle(field) {
		state := bind(c.state, ".f."+field.ID+".bool", field.Value == "true", func(v bool) {
			c.updateField(field.ID, func(x *secret.KeyValueField) { x.Value = strconv.FormatBool(v) })
		})

		return ui.ToggleField(fieldTitle(field), state.Get()).InputValue(state).SupportingText(field.SupportingText).Disabled(c.disabled).Frame(ui.Frame{}.FullWidth())
	}

	if !c.isInputOpen(kv, field) {
		return c.closedValue(field)
	}

	if c.needsNewValue(kv, field) {
		return ui.VStack(c.openInput(kv, field), tags.StatusBadge(ui.SW0, "Neu eingeben")).Alignment(ui.Leading).Gap(ui.L4).FullWidth()
	}

	return c.openInput(kv, field)
}

func (c TKeyValuesEditor) openInput(kv secret.KeyValues, field secret.KeyValueField) core.View {
	state := bind(c.state, ".f."+field.ID+".value", field.Value, func(v string) {
		c.updateField(field.ID, func(x *secret.KeyValueField) { x.Value = v })
	})

	if field.Protection != secret.Visible && field.Kind != secret.FieldMultiLine {
		return ui.PasswordField(fieldTitle(field), state.Get()).
			InputValue(state).
			Style(ui.TextFieldFilled).
			SupportingText(field.SupportingText).
			Disabled(c.disabled).
			AutoComplete(false).
			FullWidth()
	}

	lines := 0
	if field.Kind == secret.FieldMultiLine {
		lines = 4 // also if it is protected, because a password field would drop the line breaks of a pasted key
	}

	input := ui.TextField(fieldTitle(field), state.Get()).
		InputValue(state).
		Style(ui.TextFieldFilled).
		ClearButton(false).
		SupportingText(c.supportingText(kv, field)).
		Lines(lines).
		Disabled(c.disabled).
		FullWidth()

	switch field.Kind {
	case secret.FieldInt:
		input = input.KeyboardType(ui.KeyboardInteger).ShowZero(true) // a stored 0 is a value and no empty field
	case secret.FieldURL:
		input = input.KeyboardType(ui.KeyboardURL)
	}

	return input
}

func (c TKeyValuesEditor) supportingText(kv secret.KeyValues, field secret.KeyValueField) string {
	if field.Key == secret.KeyBaseURL && c.addressChanged(kv) && slices.Contains(slices.Collect(maps.Values(c.stored)), true) {
		return strings.TrimSpace(field.SupportingText + " Neue Adresse: gespeicherte Werte neu eingeben.")
	}

	return field.SupportingText
}

// closedValue shows a protected value without an input: a write-only value is not available at all and a password
// field can not hold a multi-line value.
func (c TKeyValuesEditor) closedValue(field secret.KeyValueField) core.View {
	var value core.View = ui.Text(concealedText).Font(concealedFont)
	action := "Aufdecken"
	if field.Protection == secret.WriteOnly {
		value = tags.StatusBadge(ui.SG0, "Gespeichert")
		action = "Ersetzen"
	}

	return ui.VStack(
		ui.Text(fieldTitle(field)).Font(ui.BodyMedium),
		ui.HStack(
			value,
			ui.Spacer(),
			ui.If(!c.disabled, ui.TertiaryButton(func() {
				c.setReplacing(field.ID, true)
			}).Title(action)),
		).Gap(ui.L8).FullWidth(),
		ui.If(field.SupportingText != "", smallMuted(field.SupportingText)),
	).Alignment(ui.Leading).FullWidth()
}

func (c TKeyValuesEditor) isInputOpen(kv secret.KeyValues, field secret.KeyValueField) bool {
	switch field.Protection {
	case secret.WriteOnly:
		return !c.stored[field.ID] || c.addressChanged(kv) || field.Value != "" || c.isReplacing(field.ID)
	case secret.Concealed:
		return field.Kind != secret.FieldMultiLine || field.Value == "" || c.isReplacing(field.ID)
	default:
		return true
	}
}

// isCancelable tells whether [TKeyValuesEditor.closedValue] can be shown again.
func (c TKeyValuesEditor) isCancelable(kv secret.KeyValues, field secret.KeyValueField) bool {
	if c.disabled || !c.isReplacing(field.ID) {
		return false
	}

	switch field.Protection {
	case secret.WriteOnly:
		return c.stored[field.ID] && !c.addressChanged(kv)
	case secret.Concealed:
		return field.Kind == secret.FieldMultiLine
	default:
		return false
	}
}

// cancelTitle tells the difference: a new write-only value is dropped, an edited concealed value is only hidden again.
func cancelTitle(field secret.KeyValueField) string {
	if field.Protection == secret.Concealed {
		return "Verdecken"
	}

	return "Abbrechen"
}

func (c TKeyValuesEditor) cancelReplace(field secret.KeyValueField) {
	if field.Protection == secret.WriteOnly {
		c.updateField(field.ID, func(x *secret.KeyValueField) { x.Value = "" })
	}

	c.setReplacing(field.ID, false)
}

// needsNewValue tells whether saving would lose the stored value, because the address is another one.
func (c TKeyValuesEditor) needsNewValue(kv secret.KeyValues, field secret.KeyValueField) bool {
	return field.Protection == secret.WriteOnly && c.stored[field.ID] && c.addressChanged(kv) && field.Value == ""
}

// losesStored tells whether the stored value is gone after saving, because the field is not write-only anymore.
func (c TKeyValuesEditor) losesStored(field secret.KeyValueField) bool {
	return c.stored[field.ID] && field.Protection != secret.WriteOnly && field.Value == ""
}

func (c TKeyValuesEditor) addressChanged(kv secret.KeyValues) bool {
	return kv.BaseURL() != c.storedBase
}

// protectionChip shows the protection of a field and changes it in one place.
func (c TKeyValuesEditor) protectionChip(field secret.KeyValueField) core.View {
	current := protectionInfoOf(field.Protection)

	chip := ui.HStack(
		ui.ImageIcon(current.icon).Frame(ui.Frame{}.Size(ui.L14, ui.L14)),
		ui.Text(current.title).Font(ui.BodySmall),
		ui.If(!c.disabled, ui.ImageIcon(heroOutline.ChevronDown).Frame(ui.Frame{}.Size(ui.L12, ui.L12))),
	).Gap(ui.L4).
		BackgroundColor(ui.M1).
		Border(ui.Border{}.Radius(ui.L16).Color(ui.ColorLine).Width(ui.L1)).
		Padding(ui.Padding{}.Horizontal(ui.L8).Vertical(ui.L2)).
		AccessibilityLabel("Schutz von " + fieldTitle(field) + ": " + current.title)

	if c.disabled {
		return chip
	}

	var items []ui.TMenuItem
	for _, info := range protectionInfos(field.Kind) {
		items = append(items, ui.MenuItem(func() {
			c.setProtection(field.ID, info.protection)
		}, ui.HStack(
			ui.ImageIcon(info.icon).Frame(ui.Frame{}.Size(ui.L16, ui.L16)),
			ui.VStack(
				ui.Text(info.title),
				smallMuted(info.description),
			).Alignment(ui.Leading),
			ui.Spacer(),
			ui.If(info.protection == field.Protection, ui.ImageIcon(heroOutline.Check).Frame(ui.Frame{}.Size(ui.L16, ui.L16))),
		).Gap(ui.L8).FullWidth()))
	}

	return ui.Menu(chip, ui.MenuGroup(items...))
}

type protectionInfo struct {
	protection  secret.FieldProtection
	icon        core.SVG
	title       string
	description string
}

var allProtectionInfos = []protectionInfo{
	{secret.Visible, heroOutline.Eye, "Sichtbar", "Klartext für alle mit Zugriff"},
	{secret.Concealed, heroOutline.EyeSlash, "Verdeckt", "Maskiert, Programme und KI lesen ihn"},
	{secret.WriteOnly, heroOutline.LockClosed, "Nur schreibbar", "Nie wieder lesbar, nur ersetzbar"},
}

func protectionInfoOf(protection secret.FieldProtection) protectionInfo {
	for _, info := range allProtectionInfos {
		if info.protection == protection {
			return info
		}
	}

	return allProtectionInfos[0]
}

func protectionInfos(kind secret.FieldKind) []protectionInfo {
	if secret.WriteOnlyAllowed(kind) {
		return allProtectionInfos
	}

	return slices.DeleteFunc(slices.Clone(allProtectionInfos), func(info protectionInfo) bool {
		return info.protection == secret.WriteOnly
	})
}

func (c TKeyValuesEditor) setProtection(id string, protection secret.FieldProtection) {
	c.updateField(id, func(x *secret.KeyValueField) { x.Protection = protection })
	c.setReplacing(id, false)
}

// actionsMenu is the menu of a section or a field. Without a neighbour, up or down is nil.
func (c TKeyValuesEditor) actionsMenu(title, id string, up, down, remove func()) core.View {
	if c.disabled {
		return nil
	}

	items := []ui.TMenuItem{menuEntry(heroOutline.PencilSquare, "Bearbeiten", func() {
		c.editing.Set(id)
		c.editingOpen.Set(true)
	})}

	if up != nil {
		items = append(items, menuEntry(heroOutline.ArrowUp, "Nach oben", up))
	}

	if down != nil {
		items = append(items, menuEntry(heroOutline.ArrowDown, "Nach unten", down))
	}

	return ui.Menu(
		ui.TertiaryButton(nil).PreIcon(heroOutline.EllipsisVertical).AccessibilityLabel("Aktionen: "+title),
		ui.MenuGroup(items...),
		ui.MenuGroup(menuEntry(heroOutline.Trash, "Entfernen", remove)),
	)
}

func menuEntry(icon core.SVG, title string, action func()) ui.TMenuItem {
	return ui.MenuItem(action, ui.HStack(
		ui.ImageIcon(icon).Frame(ui.Frame{}.Size(ui.L16, ui.L16)),
		ui.Text(title),
	).Gap(ui.L8))
}

// fieldPreset is an entry of the menu to add fields. A password manager offers types and no keys to learn.
type fieldPreset struct {
	icon   core.SVG
	title  string
	fields []secret.KeyValueField
	askKey bool // the key needs the input of the user
}

// connectionPresets offers what a connection consumes, and only what can still be added, because a key is unique and
// only one kind of login can set the Authorization header.
func connectionPresets(kv secret.KeyValues) []fieldPreset {
	keys := map[string]bool{}
	for field := range kv.Fields() {
		keys[field.Key] = true
	}

	var presets []fieldPreset
	if !keys[secret.KeyBaseURL] {
		presets = append(presets, fieldPreset{icon: heroOutline.GlobeAlt, title: "Adresse", fields: []secret.KeyValueField{
			{Key: secret.KeyBaseURL, Label: "Adresse", Kind: secret.FieldURL},
		}})
	}

	hasBasic := keys[secret.KeyHeaderBasicAuthUser] || keys[secret.KeyHeaderBasicAuthPassword]
	if !keys[secret.KeyHeaderBearerAuth] && !hasBasic {
		presets = append(presets, fieldPreset{icon: heroOutline.Key, title: "Bearer-Token", fields: []secret.KeyValueField{
			{Key: secret.KeyHeaderBearerAuth, Label: "Token", Kind: secret.FieldString, Protection: secret.WriteOnly},
		}}, fieldPreset{icon: heroOutline.User, title: "Basic-Auth", fields: []secret.KeyValueField{
			{Key: secret.KeyHeaderBasicAuthUser, Label: "Benutzername", Kind: secret.FieldString},
			{Key: secret.KeyHeaderBasicAuthPassword, Label: "Passwort", Kind: secret.FieldString, Protection: secret.WriteOnly},
		}})
	}

	headerKey := "HEADER_X_API_KEY"
	if keys[headerKey] {
		headerKey = secret.KeyPrefixHeader
	}

	return append(presets,
		fieldPreset{icon: heroOutline.Link, title: "API-Schlüssel im Header", askKey: true, fields: []secret.KeyValueField{
			{Key: headerKey, Label: "API-Schlüssel", Kind: secret.FieldString, Protection: secret.WriteOnly},
		}},
		fieldPreset{icon: heroOutline.Link, title: "Query-Parameter", askKey: true, fields: []secret.KeyValueField{
			{Key: secret.KeyPrefixQuery, Label: "Parameter", Kind: secret.FieldString, Protection: secret.WriteOnly},
		}},
	)
}

var valuePresets = []fieldPreset{
	{icon: heroOutline.Bars3BottomLeft, title: "Text", askKey: true, fields: []secret.KeyValueField{{Kind: secret.FieldString}}},
	{icon: heroOutline.EyeSlash, title: "Verdeckter Text", askKey: true, fields: []secret.KeyValueField{{Kind: secret.FieldString, Protection: secret.Concealed}}},
	{icon: heroOutline.Hashtag, title: "Ganzzahl", askKey: true, fields: []secret.KeyValueField{{Kind: secret.FieldInt}}},
	{icon: heroOutline.CheckCircle, title: "Schalter", askKey: true, fields: []secret.KeyValueField{{Kind: secret.FieldBool, Value: "false"}}},
	{icon: heroOutline.GlobeAlt, title: "URL", askKey: true, fields: []secret.KeyValueField{{Kind: secret.FieldURL}}},
	{icon: heroOutline.DocumentText, title: "Mehrzeiliger Text", askKey: true, fields: []secret.KeyValueField{{Kind: secret.FieldMultiLine}}},
}

// addFieldMenu adds to the section, or to a new one, if the identifier is empty.
func (c TKeyValuesEditor) addFieldMenu(kv secret.KeyValues, sectionID string, anchor core.View) core.View {
	items := func(presets []fieldPreset) []ui.TMenuItem {
		res := make([]ui.TMenuItem, 0, len(presets))
		for _, preset := range presets {
			res = append(res, menuEntry(preset.icon, preset.title, func() { c.addPreset(sectionID, preset) }))
		}

		return res
	}

	return ui.Menu(anchor, ui.MenuGroup(items(connectionPresets(kv))...), ui.MenuGroup(items(valuePresets)...))
}

func (c TKeyValuesEditor) addPreset(sectionID string, preset fieldPreset) {
	first := ""
	c.update(func(kv *secret.KeyValues) {
		idx := slices.IndexFunc(kv.Sections, func(s secret.KeyValueSection) bool { return s.ID == sectionID })
		if idx < 0 {
			kv.Sections = append(kv.Sections, secret.KeyValueSection{ID: data.RandIdent[string]()})
			idx = len(kv.Sections) - 1
		}

		for _, field := range preset.fields {
			field.ID = data.RandIdent[string]()
			kv.Sections[idx].Fields = append(kv.Sections[idx].Fields, field)
			if first == "" {
				first = field.ID
			}
		}
	})

	if preset.askKey {
		c.editing.Set(first)
		c.editingOpen.Set(true)
	}
}

func (c TKeyValuesEditor) addSection() {
	id := data.RandIdent[string]()
	c.update(func(kv *secret.KeyValues) {
		kv.Sections = append(kv.Sections, secret.KeyValueSection{ID: id})
	})
	c.editing.Set(id)
	c.editingOpen.Set(true)
}

// dialog edits what is rarely changed. The inputs apply at once, the page decides about saving.
func (c TKeyValuesEditor) dialog(kv secret.KeyValues) core.View {
	if !c.editingOpen.Get() || c.disabled {
		return nil
	}

	id := c.editing.Get()
	for _, section := range kv.Sections {
		if section.ID == id {
			return alert.Dialog("Abschnitt bearbeiten", c.sectionForm(section), c.editingOpen, alert.Closeable(), alert.Ok())
		}
	}

	for field := range kv.Fields() {
		if field.ID == id {
			return alert.Dialog("Feld bearbeiten", c.fieldForm(field), c.editingOpen, alert.Closeable(), alert.Ok())
		}
	}

	return nil
}

func (c TKeyValuesEditor) sectionForm(section secret.KeyValueSection) core.View {
	return ui.VStack(
		c.sectionText(section, ".label", "Titel", section.Label, func(x *secret.KeyValueSection, v string) {
			x.Label = v
		}),
		c.sectionText(section, ".supporting", "Erläuterung", section.SupportingText, func(x *secret.KeyValueSection, v string) {
			x.SupportingText = v
		}),
	).Alignment(ui.Leading).Gap(ui.L16).FullWidth()
}

func (c TKeyValuesEditor) fieldForm(field secret.KeyValueField) core.View {
	kind := bind(c.state, ".f."+field.ID+".kind", field.Kind, func(v secret.FieldKind) {
		c.updateField(field.ID, func(x *secret.KeyValueField) { applyKind(x, v) })
	})

	return ui.VStack(
		c.fieldText(field, ".label", "Bezeichnung", field.Label, "", func(x *secret.KeyValueField, v string) { x.Label = v }),
		c.fieldText(field, ".key", "Schlüssel", field.Key, keyRule, applyKey),
		c.fieldText(field, ".supporting", "Erläuterung", field.SupportingText, "", func(x *secret.KeyValueField, v string) { x.SupportingText = v }),
		dropdown.Dropdown("Art", fieldKindOptions, kind.Get()).
			InputValue(kind).
			Frame(ui.Frame{}.FullWidth()),
		fieldUse(field),
	).Alignment(ui.Leading).Gap(ui.L16).FullWidth()
}

func (c TKeyValuesEditor) logoView(wnd core.Window, kv secret.KeyValues) core.View {
	view := emptyTile()
	if kv.LogoSVG != "" {
		view = tile(logoImage(kv.LogoSVG, ui.L40))
	}

	if c.disabled {
		return view
	}

	items := []ui.TMenuItem{menuEntry(heroOutline.ArrowUpTray, "Logo hochladen", func() {
		c.importLogo(wnd)
	})}

	if kv.LogoSVG != "" {
		items = append(items, menuEntry(heroOutline.Trash, "Logo entfernen", func() {
			c.update(func(x *secret.KeyValues) { x.LogoSVG = "" })
		}))
	}

	return ui.Menu(view.AccessibilityLabel("Logo ändern"), ui.MenuGroup(items...))
}

func (c TKeyValuesEditor) importLogo(wnd core.Window) {
	wnd.ImportFiles(core.ImportFilesOptions{
		ID:               c.state.ID() + ".logo",
		MaxBytes:         maxLogoUpload,
		AllowedMimeTypes: []string{"image/svg+xml"},
		OnCompletion: func(files []core.File) {
			if len(files) == 0 {
				return
			}

			logo, err := readLogo(files[0])
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}

			wnd.Post(func() {
				c.update(func(x *secret.KeyValues) { x.LogoSVG = logo })
			})
		},
	})
}

func readLogo(file core.File) (string, error) {
	r, err := file.Open()
	if err != nil {
		return "", err
	}

	defer r.Close()

	buf, err := io.ReadAll(io.LimitReader(r, secret.MaxLogoSVGBytes+1))
	if err != nil {
		return "", err
	}

	return secret.ParseLogoSVG(buf)
}

func (c TKeyValuesEditor) isReplacing(id string) bool {
	return slices.Contains(c.replacing.Get(), id)
}

func (c TKeyValuesEditor) setReplacing(id string, replacing bool) {
	ids := slices.DeleteFunc(slices.Clone(c.replacing.Get()), func(s string) bool { return s == id })
	if replacing {
		ids = append(ids, id)
	}

	c.replacing.Set(ids)
}

func (c TKeyValuesEditor) textField(postfix, label, value, hint string, apply func(*secret.KeyValues, string)) ui.TTextField {
	state := bind(c.state, postfix, value, func(v string) {
		c.update(func(x *secret.KeyValues) { apply(x, v) })
	})

	return ui.TextField(label, state.Get()).InputValue(state).Style(ui.TextFieldFilled).ClearButton(false).SupportingText(hint).Disabled(c.disabled).FullWidth()
}

func (c TKeyValuesEditor) sectionText(section secret.KeyValueSection, postfix, label, value string, apply func(*secret.KeyValueSection, string)) core.View {
	state := bind(c.state, ".s."+section.ID+postfix, value, func(v string) {
		c.update(func(x *secret.KeyValues) {
			if i := slices.IndexFunc(x.Sections, func(s secret.KeyValueSection) bool { return s.ID == section.ID }); i >= 0 {
				apply(&x.Sections[i], v)
			}
		})
	})

	return ui.TextField(label, state.Get()).InputValue(state).Style(ui.TextFieldFilled).ClearButton(false).FullWidth()
}

func (c TKeyValuesEditor) fieldText(field secret.KeyValueField, postfix, label, value, hint string, apply func(*secret.KeyValueField, string)) core.View {
	state := bind(c.state, ".f."+field.ID+postfix, value, func(v string) {
		c.updateField(field.ID, func(x *secret.KeyValueField) { apply(x, v) })
	})

	return ui.TextField(label, state.Get()).InputValue(state).Style(ui.TextFieldFilled).ClearButton(false).SupportingText(hint).FullWidth()
}

// update changes a copy, because the page compares the state with its initial value, which shares the slices.
func (c TKeyValuesEditor) update(fn func(*secret.KeyValues)) {
	next := c.state.Get().Clone()
	fn(&next)
	c.state.Update(next)
}

func (c TKeyValuesEditor) updateField(id string, fn func(*secret.KeyValueField)) {
	c.update(func(kv *secret.KeyValues) {
		for i := range kv.Sections {
			for j := range kv.Sections[i].Fields {
				if kv.Sections[i].Fields[j].ID == id {
					fn(&kv.Sections[i].Fields[j])
					return
				}
			}
		}
	})
}

func (c TKeyValuesEditor) moveField(id string, delta int) {
	c.update(func(kv *secret.KeyValues) {
		for i := range kv.Sections {
			if moveItem(kv.Sections[i].Fields, delta, func(f secret.KeyValueField) bool { return f.ID == id }) {
				return
			}
		}
	})
}

func (c TKeyValuesEditor) moveSection(id string, delta int) {
	c.update(func(kv *secret.KeyValues) {
		moveItem(kv.Sections, delta, func(s secret.KeyValueSection) bool { return s.ID == id })
	})
}

// moveItem swaps the matching item with its neighbour and returns whether the item has been found.
func moveItem[T any](items []T, delta int, match func(T) bool) bool {
	i := slices.IndexFunc(items, match)
	if i < 0 {
		return false
	}

	if j := i + delta; j >= 0 && j < len(items) {
		items[i], items[j] = items[j], items[i]
	}

	return true
}

// removeField and removeSection leave a nil slice behind, because a loaded secret has one, too, and the page detects
// changes by a deep comparison.
func (c TKeyValuesEditor) removeField(id string) {
	c.update(func(kv *secret.KeyValues) {
		for i := range kv.Sections {
			kv.Sections[i].Fields = nilIfEmpty(slices.DeleteFunc(kv.Sections[i].Fields, func(f secret.KeyValueField) bool {
				return f.ID == id
			}))
		}
	})
}

func (c TKeyValuesEditor) removeSection(id string) {
	c.update(func(kv *secret.KeyValues) {
		kv.Sections = nilIfEmpty(slices.DeleteFunc(kv.Sections, func(s secret.KeyValueSection) bool {
			return s.ID == id
		}))
	})
}

func nilIfEmpty[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}

	return s
}

// bind returns the state of an input. The model stays the source of truth, so a value which has changed in another
// way, e.g. together with the kind of the field, replaces the text of the input.
func bind[T comparable](parent *core.State[secret.KeyValues], postfix string, value T, apply func(T)) *core.State[T] {
	state := core.DerivedState[T](parent, postfix).Init(func() T {
		return value
	})

	if state.Get() != value {
		state.Set(value)
	}

	state.Observe(apply)
	return state
}

// applyKey gives the address the kind it needs right away, otherwise saving would only tell the user to do it.
func applyKey(field *secret.KeyValueField, key string) {
	field.Key = strings.TrimSpace(key)
	if field.Key == secret.KeyBaseURL {
		applyKind(field, secret.FieldURL)
	}
}

// applyKind keeps only a value and a protection, which fit the new kind.
func applyKind(field *secret.KeyValueField, kind secret.FieldKind) {
	field.Kind = kind
	if field.Protection == secret.WriteOnly && !secret.WriteOnlyAllowed(kind) {
		field.Protection = secret.Concealed
	}

	if !secret.ValueFitsKind(kind, field.Value) {
		field.Value = ""
	}

	if kind == secret.FieldBool && field.Value == "" {
		field.Value = "false"
	}
}

var fieldKindOptions = []dropdown.Option[secret.FieldKind]{
	{Value: secret.FieldString, Label: "Text"},
	{Value: secret.FieldMultiLine, Label: "Mehrzeiliger Text"},
	{Value: secret.FieldInt, Label: "Ganzzahl"},
	{Value: secret.FieldBool, Label: "Schalter"},
	{Value: secret.FieldURL, Label: "URL"},
}
