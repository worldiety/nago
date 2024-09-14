// Package terminplanung oder Terminplanung.
// Terminplanung enthält alles was fachlich angesagt ist.
package terminplanung

import "time"

// TerminAnlegenRequested oder Termin anlegen gewünscht.
// Termin anlegen gewünscht ist die Wurzel allen Übels.
//   - und 1
//   - und 2
type TerminAnlegenRequested struct {
	Vorname  string
	Nachname string
	Um       time.Time
	Dauer    time.Time
	// Blabla Field
	Bla SomeGenericType[string] `json:"bla"`
}

type A struct {
	SomeField string
}

type B A
type C B

// TerminErstellt ist wie [TerminAnlegenRequested] nur als Kurzschreibweise.
type TerminErstellt TerminAnlegenRequested

// TerminAnlegen ist der Anwendungsfall mit Ein- und Ausgabe.
type TerminAnlegen func(TerminAnlegenRequested) (TerminAnlegenRequested, error)

// SomeGenericType is blabla.
type SomeGenericType[T any] struct{}

// ZeitstempelInMinuten oder Zeitstempel in Minuten.
// Zeitstempel in Minuten ist sehr technisch aber der Kunde kennt das auch so.
type ZeitstempelInMinuten int64

type StartZeitPunkt ZeitstempelInMinuten

type AlternativeZeitstempel = ZeitstempelInMinuten

// VerhaltensVertrag is just an iface.
type VerhaltensVertrag interface {
	Zerstöre()
}
