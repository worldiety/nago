package ui

import "go.wdy.de/nago/container/slice"

type Datepicker struct {
	id                 CID
	disabled           Bool
	label              String
	hint               String
	error              String
	expanded           Bool
	selectedDay        Int
	selectedMonthIndex Int
	selectedYear       Int
	onClicked          *Func
	onSelectionChanged *Func
	properties         slice.Slice[Property]
}

func NewDatepicker(with func(datepicker *Datepicker)) *Datepicker {
	c := &Datepicker{
		id:                 nextPtr(),
		disabled:           NewShared[bool]("disabled"),
		label:              NewShared[string]("label"),
		hint:               NewShared[string]("hint"),
		error:              NewShared[string]("error"),
		expanded:           NewShared[bool]("expanded"),
		selectedDay:        NewShared[int64]("selectedDay"),
		selectedMonthIndex: NewShared[int64]("selectedMonthIndex"),
		selectedYear:       NewShared[int64]("selectedYear"),
		onClicked:          NewFunc("onClicked"),
		onSelectionChanged: NewFunc("onSelectionChanged"),
	}

	c.properties = slice.Of[Property](c.disabled, c.label, c.hint, c.error, c.expanded, c.selectedDay, c.selectedMonthIndex, c.selectedYear, c.onClicked, c.onSelectionChanged)
	if with != nil {
		with(c)
	}
	return c
}

func (c *Datepicker) ID() CID {
	return c.id
}

func (c *Datepicker) Type() string {
	return "Datepicker"
}

func (c *Datepicker) Disabled() Bool {
	return c.disabled
}

func (c *Datepicker) Label() String { return c.label }

func (c *Datepicker) Hint() String { return c.hint }

func (c *Datepicker) Error() String { return c.error }

func (c *Datepicker) Expanded() Bool {
	return c.expanded
}

func (c *Datepicker) SelectedDay() Int {
	return c.selectedDay
}

func (c *Datepicker) SelectedMonthIndex() Int {
	return c.selectedMonthIndex
}

func (c *Datepicker) SelectedYear() Int {
	return c.selectedYear
}

func (c *Datepicker) OnClicked() *Func {
	return c.onClicked
}

func (c *Datepicker) OnSelectionChanged() *Func {
	return c.onSelectionChanged
}

func (c *Datepicker) Properties() slice.Slice[Property] {
	return c.properties
}
