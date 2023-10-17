package ui

type InputText struct {
	Name       string // the unique name of this input within a form
	Value      string // the pre-filled text, as if it has been entered by the user
	Label      string // the caption of the field
	Supporting string // text which describes the required information in a sentence
	Disabled   string // if true, the field cannot be modified by the user
}
