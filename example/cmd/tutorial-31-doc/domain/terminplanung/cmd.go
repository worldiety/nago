// Package terminplanung oder Terminplanung.
// Terminplanung enthält alles was fachlich angesagt ist.
package terminplanung

import (
	"go.wdy.de/nago/annotation"
	"time"
)

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

var permTerminAnlegen = annotation.Usecase[TerminAnlegen](
	"Termin anlegen macht den Terminplan voll. Wir haben jetzt die Möglichkeit redundante Kommentare zu verfassen.",
	"Siehe auch: ", annotation.TypeLink[StartZeitPunkt]("Durchstarter"),
	`
Multiline:
* why
* not 
* just markdown anywhere? Problem resolving types? => just use the alias?
`,
).
	Synonyms("Termin anlegen", "Leute verplanen").
	Permission("de.worldiety.termin.anlegen", "Termin anlegen", "Leuten den Terminkalender voll machen.")

// TerminAnlegen ist der Anwendungsfall mit Ein- und Ausgabe.
type TerminAnlegen func(TerminAnlegenRequested) (TerminAnlegenRequested, error)

// SomeGenericType is blabla.
type SomeGenericType[T any] struct{}

// ZeitstempelInMinuten oder Zeitstempel in Minuten.
// Zeitstempel in Minuten ist sehr technisch aber der Kunde kennt das auch so.
//
//doc:stereotype value
//doc:alias Zeitstempel in Minuten
type ZeitstempelInMinuten int64

type StartZeitPunkt ZeitstempelInMinuten

type AlternativeZeitstempel = ZeitstempelInMinuten

// VerhaltensVertrag is just an iface.
type VerhaltensVertrag interface {
	Zerstöre()
}
