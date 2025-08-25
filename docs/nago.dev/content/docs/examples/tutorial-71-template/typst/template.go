package typst

import (
	"embed"
)

//go:embed main.typ.tpl logo.png aubina.typ
var Template embed.FS

type Address struct {
	Street string
	City   string
}

type Author struct {
	Name     string
	Birthday Date
	Address  Address
}
type Company struct {
	Name    string
	Address Address
}
type Date struct {
	Year  int
	Month int
	Day   int
}
type Training struct {
	Start Date
	End   Date
}

type Task struct {
	Description       string
	DurationInMinutes int
}

type EntryKind int

const (
	Day EntryKind = iota
	Signature
)

type Entry struct {
	Kind  EntryKind
	Date  Date
	Tasks []Task
	Place string
}
type Report struct {
	Title    string
	Subtitle string
	Author   Author
	Company  Company
	Trainer  string
	Training Training
	Entries  []Entry
}
