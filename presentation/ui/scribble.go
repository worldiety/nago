package ui

import (
	"go.wdy.de/nago/container/slice"
)

type DeleteEvent[ID any] struct {
	Entity ID
}

type CreateEvent string
type UpdateEvent string

type Overview struct {
	Delete DeleteEvent[string]
	Create CreateEvent
	Update UpdateEvent
}

type Detail struct {
}

type Opt[T any] struct{}

type ListPersona struct {
	ID    any // identifies the group/collection/repository of ???, required
	Items slice.Slice[ListItem]

	Delete        Opt[DeleteView] // ???
	Edit          Opt[EditView]
	DeleteBatch   Opt[Confirmation]
	BatchBeliebig any //???
}

type CreateView struct {
}

type EditView struct {
	EntityID any //??? required
}

type DeleteView struct {
}

// see https://m3.material.io/components/dialogs/guidelines#0119a5cb-6943-4f0e-9e06-a587ad848aa4
type Confirmation struct {
	Icon           Image  // optional
	Headline       string // optional
	SupportingText string // required
	Dismissive     Action // required
	Confirming     Action // required
}

type Action struct {
	Icon      Image  // optional
	LabelText string // optional ???
	Event     any    // how to transport the actual action event???
	F         func()
}

// RMI???
type EventHandler[E any] struct {
	Event   E
	Handler func(E)
}

// Lesend: normale Entity: schreib result: mit Fehlern pro Feld

type Fontname string

type FontIcon struct {
	Codepoint rune
	Font      Fontname
}
