package ui

import "go.wdy.de/nago/container/slice"

type Datepicker struct {
	id                 CID
	disabled           Bool
	label              String
	hint               String
	error              String
	expanded           Bool
	rangeMode          Bool
	startDateSelected  Bool
	selectedStartDay   Int
	selectedStartMonth Int
	selectedStartYear  Int
	endDateSelected    Bool
	selectedEndDay     Int
	selectedEndMonth   Int
	selectedEndYear    Int
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
		rangeMode:          NewShared[bool]("rangeMode"),
		startDateSelected:  NewShared[bool]("startDateSelected"),
		selectedStartDay:   NewShared[int64]("selectedStartDay"),
		selectedStartMonth: NewShared[int64]("selectedStartMonth"),
		selectedStartYear:  NewShared[int64]("selectedStartYear"),
		endDateSelected:    NewShared[bool]("endDateSelected"),
		selectedEndDay:     NewShared[int64]("selectedEndDay"),
		selectedEndMonth:   NewShared[int64]("selectedEndMonth"),
		selectedEndYear:    NewShared[int64]("selectedEndYear"),
		onClicked:          NewFunc("onClicked"),
		onSelectionChanged: NewFunc("onSelectionChanged"),
	}

	c.properties = slice.Of[Property](c.disabled, c.label, c.hint, c.error, c.expanded, c.rangeMode, c.startDateSelected, c.selectedStartDay, c.selectedStartMonth, c.selectedStartYear, c.endDateSelected, c.selectedEndDay, c.selectedEndMonth, c.selectedEndYear, c.onClicked, c.onSelectionChanged)
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

func (c *Datepicker) RangeMode() Bool { return c.rangeMode }

func (c *Datepicker) StartDateSelected() Bool {
	return c.startDateSelected
}

func (c *Datepicker) SelectedStartDay() Int {
	return c.selectedStartDay
}

func (c *Datepicker) SelectedStartMonth() Int {
	return c.selectedStartMonth
}

func (c *Datepicker) SelectedStartYear() Int {
	return c.selectedStartYear
}

func (c *Datepicker) EndDateSelected() Bool {
	return c.endDateSelected
}

func (c *Datepicker) SelectedEndDay() Int {
	return c.selectedEndDay
}

func (c *Datepicker) SelectedEndMonth() Int {
	return c.selectedEndMonth
}

func (c *Datepicker) SelectedEndYear() Int {
	return c.selectedEndYear
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
