// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uindbinspector

import (
	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

// Localized strings for the ndb inspector UI. Plain strings use MustString;
// strings with interpolated values use MustVarString with {placeholders}.
var (
	// generic / access
	StrNoAccessTitle = i18n.MustString("nago.ndbinspector.no_access.title",
		i18n.Values{language.German: "Kein Zugriff", language.English: "No access"})
	StrNoAccessBody = i18n.MustString("nago.ndbinspector.no_access.body",
		i18n.Values{language.German: "Es fehlt die Berechtigung nago.ndb.inspector.", language.English: "The nago.ndb.inspector permission is missing."})
	StrNoNdbTitle = i18n.MustString("nago.ndbinspector.no_ndb.title",
		i18n.Values{language.German: "Keine ndb Datenbank", language.English: "No ndb database"})
	StrNoNdbBody = i18n.MustString("nago.ndbinspector.no_ndb.body",
		i18n.Values{language.German: "Es wurde keine ndb Datenbank registriert.", language.English: "No ndb database has been registered."})

	// selectors
	StrDatabase = i18n.MustString("nago.ndbinspector.database",
		i18n.Values{language.German: "Datenbank", language.English: "Database"})
	StrEngine = i18n.MustString("nago.ndbinspector.engine",
		i18n.Values{language.German: "Engine", language.English: "Engine"})
	StrColumn = i18n.MustString("nago.ndbinspector.column",
		i18n.Values{language.German: "Spalte", language.English: "Column"})

	// messages page
	StrMessagesTitle = i18n.MustString("nago.ndbinspector.messages.title",
		i18n.Values{language.German: "ndb Nachrichten", language.English: "ndb Messages"})
	StrNoMsgEngineTitle = i18n.MustString("nago.ndbinspector.messages.no_engine.title",
		i18n.Values{language.German: "Keine Message-Datenbank", language.English: "No message database"})
	StrNoMsgEngineBody = i18n.MustString("nago.ndbinspector.messages.no_engine.body",
		i18n.Values{language.German: "In dieser ndb Datenbank wurde keine msgstore-Engine gefunden.", language.English: "No msgstore engine was found in this ndb database."})
	StrNoStreams = i18n.MustString("nago.ndbinspector.messages.no_streams",
		i18n.Values{language.German: "Keine Streams vorhanden.", language.English: "No streams available."})
	StrMessageTypes = i18n.MustString("nago.ndbinspector.messages.types",
		i18n.Values{language.German: "Nachrichtentypen", language.English: "Message types"})
	StrMessageTypesHint = i18n.MustString("nago.ndbinspector.messages.types_hint",
		i18n.Values{language.German: "Mehrfachauswahl – leer = alle Typen", language.English: "Multi-select – empty = all types"})
	StrSelectEngineHint = i18n.MustString("nago.ndbinspector.messages.select_engine_hint",
		i18n.Values{language.German: "Wähle eine Engine aus, um Nachrichten anzuzeigen.", language.English: "Select an engine to view messages."})
	StrNoMessagesInWindow = i18n.MustString("nago.ndbinspector.messages.none_in_window",
		i18n.Values{language.German: "Keine Nachrichten in diesem Fenster", language.English: "No messages in this window"})
	StrFromSeq = i18n.MustString("nago.ndbinspector.messages.from_seq",
		i18n.Values{language.German: "Ab Seq", language.English: "From Seq"})
	StrTimeIndex = i18n.MustString("nago.ndbinspector.messages.time_index",
		i18n.Values{language.German: "Zeitindex", language.English: "Time index"})

	// message table columns
	StrColSeq = i18n.MustString("nago.ndbinspector.col.seq",
		i18n.Values{language.German: "Seq", language.English: "Seq"})
	StrColTime = i18n.MustString("nago.ndbinspector.col.time",
		i18n.Values{language.German: "Zeit", language.English: "Time"})
	StrColType = i18n.MustString("nago.ndbinspector.col.type",
		i18n.Values{language.German: "Typ", language.English: "Type"})
	StrColTrace = i18n.MustString("nago.ndbinspector.col.trace",
		i18n.Values{language.German: "Trace", language.English: "Trace"})
	StrColSize = i18n.MustString("nago.ndbinspector.col.size",
		i18n.Values{language.German: "Größe", language.English: "Size"})
	StrColPreview = i18n.MustString("nago.ndbinspector.col.preview",
		i18n.Values{language.German: "Vorschau", language.English: "Preview"})
	StrColField = i18n.MustString("nago.ndbinspector.col.field",
		i18n.Values{language.German: "Feld", language.English: "Field"})
	StrColValue = i18n.MustString("nago.ndbinspector.col.value",
		i18n.Values{language.German: "Wert", language.English: "Value"})
	StrColMillis = i18n.MustString("nago.ndbinspector.col.millis",
		i18n.Values{language.German: "Millis", language.English: "Millis"})

	// message detail
	StrMessageX = i18n.MustVarString("nago.ndbinspector.messages.message_x",
		i18n.Values{language.German: "Nachricht {seq}", language.English: "Message {seq}"})
	StrEmpty = i18n.MustString("nago.ndbinspector.empty",
		i18n.Values{language.German: "<leer>", language.English: "<empty>"})
	StrBinaryDataX = i18n.MustVarString("nago.ndbinspector.binary_data_x",
		i18n.Values{language.German: "<binäre Daten, {bytes} Bytes>", language.English: "<binary data, {bytes} bytes>"})

	// knife tools (messages)
	StrDeleteMessage = i18n.MustString("nago.ndbinspector.messages.delete_message",
		i18n.Values{language.German: "Nachricht löschen", language.English: "Delete message"})
	StrDeleteMessageBody = i18n.MustVarString("nago.ndbinspector.messages.delete_message_body",
		i18n.Values{language.German: "Seq {seq} aus Stream {type} als gelöscht markieren (Tombstone)?", language.English: "Mark Seq {seq} in stream {type} as deleted (tombstone)?"})
	StrDelete = i18n.MustString("nago.ndbinspector.action.delete",
		i18n.Values{language.German: "Löschen", language.English: "Delete"})
	StrDeleteStream = i18n.MustString("nago.ndbinspector.messages.delete_stream",
		i18n.Values{language.German: "Stream löschen", language.English: "Delete stream"})
	StrDeleteStreamBody = i18n.MustVarString("nago.ndbinspector.messages.delete_stream_body",
		i18n.Values{language.German: "Den gesamten Stream {type} unwiderruflich löschen?", language.English: "Irreversibly delete the whole stream {type}?"})

	// message stat row: "Seq {min}–{max}{pending} · {count} Nachrichten · {segments} Segmente · {size}"
	StrMsgStatRow = i18n.MustVarString("nago.ndbinspector.messages.stat_row",
		i18n.Values{
			language.German:  "Seq {min}–{max}{pending} · {count} Nachrichten · {segments} Segmente · {size}",
			language.English: "Seq {min}–{max}{pending} · {count} messages · {segments} segments · {size}",
		})
	StrMsgPageLabel = i18n.MustVarString("nago.ndbinspector.messages.page_label",
		i18n.Values{
			language.German:  "Seq {min}–{max} · Seite {page} von {pages}",
			language.English: "Seq {min}–{max} · Page {page} of {pages}",
		})

	// timeseries page
	StrTimeseriesTitle = i18n.MustString("nago.ndbinspector.timeseries.title",
		i18n.Values{language.German: "ndb Zeitreihen", language.English: "ndb Time series"})
	StrNoTsEngineTitle = i18n.MustString("nago.ndbinspector.timeseries.no_engine.title",
		i18n.Values{language.German: "Keine Timeseries-Datenbank", language.English: "No time series database"})
	StrNoTsEngineBody = i18n.MustString("nago.ndbinspector.timeseries.no_engine.body",
		i18n.Values{language.German: "In dieser ndb Datenbank wurde keine tsdb-Engine gefunden.", language.English: "No tsdb engine was found in this ndb database."})
	StrSelectColumnHint = i18n.MustString("nago.ndbinspector.timeseries.select_column_hint",
		i18n.Values{language.German: "Wähle links eine Spalte aus.", language.English: "Select a column on the left."})
	StrNoDataTitle = i18n.MustString("nago.ndbinspector.timeseries.no_data.title",
		i18n.Values{language.German: "Keine Daten", language.English: "No data"})
	StrNoDataBody = i18n.MustVarString("nago.ndbinspector.timeseries.no_data.body",
		i18n.Values{language.German: "Die Spalte {column} enthält keine Datenpunkte.", language.English: "Column {column} contains no data points."})
	StrNoColumns = i18n.MustString("nago.ndbinspector.timeseries.no_columns",
		i18n.Values{language.German: "Keine Spalten vorhanden.", language.English: "No columns available."})
	StrEmptyRange = i18n.MustString("nago.ndbinspector.timeseries.empty_range",
		i18n.Values{language.German: "leer", language.English: "empty"})

	// column stat row: "{scheme} · {count} Punkte · {chunks} Chunks · {size}"
	StrColStatRow = i18n.MustVarString("nago.ndbinspector.timeseries.stat_row",
		i18n.Values{
			language.German:  "{scheme} · {count} Punkte · {chunks} Chunks · {size}",
			language.English: "{scheme} · {count} points · {chunks} chunks · {size}",
		})
	StrRangeSpan = i18n.MustVarString("nago.ndbinspector.timeseries.range_span",
		i18n.Values{language.German: "{from} – {to}", language.English: "{from} – {to}"})

	// chart controls
	StrRangeTotal = i18n.MustString("nago.ndbinspector.timeseries.range_total",
		i18n.Values{language.German: "Gesamt", language.English: "All"})
	StrFromMs = i18n.MustString("nago.ndbinspector.timeseries.from_ms",
		i18n.Values{language.German: "Von (ms)", language.English: "From (ms)"})
	StrToMs = i18n.MustString("nago.ndbinspector.timeseries.to_ms",
		i18n.Values{language.German: "Bis (ms)", language.English: "To (ms)"})
	StrChartTime = i18n.MustString("nago.ndbinspector.timeseries.chart_time",
		i18n.Values{language.German: "Zeit", language.English: "Time"})
	StrChartValue = i18n.MustString("nago.ndbinspector.timeseries.chart_value",
		i18n.Values{language.German: "Wert", language.English: "Value"})
	StrChartNoData = i18n.MustString("nago.ndbinspector.timeseries.chart_no_data",
		i18n.Values{language.German: "Keine Datenpunkte im gewählten Bereich", language.English: "No data points in the selected range"})
	StrChartCaption = i18n.MustVarString("nago.ndbinspector.timeseries.chart_caption",
		i18n.Values{
			language.German:  "Bereich {from} – {to} · M4 auf ~{buckets} Buckets · {points} gezeichnete Punkte",
			language.English: "Range {from} – {to} · M4 to ~{buckets} buckets · {points} drawn points",
		})

	// string window
	StrFromMsShort = i18n.MustString("nago.ndbinspector.timeseries.from_ms_short",
		i18n.Values{language.German: "Ab (ms)", language.English: "From (ms)"})
	StrNoValuesInWindow = i18n.MustString("nago.ndbinspector.timeseries.none_in_window",
		i18n.Values{language.German: "Keine Werte in diesem Fenster", language.English: "No values in this window"})
	StrTsPageLabel = i18n.MustVarString("nago.ndbinspector.timeseries.page_label",
		i18n.Values{
			language.German:  "{from} – {to} · Seite {page} von {pages}",
			language.English: "{from} – {to} · Page {page} of {pages}",
		})

	// knife tools (timeseries)
	StrCompact = i18n.MustString("nago.ndbinspector.timeseries.compact",
		i18n.Values{language.German: "Kompaktieren", language.English: "Compact"})
	StrDeleteRange = i18n.MustString("nago.ndbinspector.timeseries.delete_range",
		i18n.Values{language.German: "Bereich löschen", language.English: "Delete range"})
	StrDeleteRangeTitle = i18n.MustString("nago.ndbinspector.timeseries.delete_range.title",
		i18n.Values{language.German: "Datenbereich löschen", language.English: "Delete data range"})
	StrDeleteRangeBody = i18n.MustVarString("nago.ndbinspector.timeseries.delete_range.body",
		i18n.Values{
			language.German:  "Alle Datenpunkte der Spalte {column} im Bereich {from} – {to} als gelöscht markieren?",
			language.English: "Mark all data points of column {column} in the range {from} – {to} as deleted?",
		})
)
