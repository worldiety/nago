// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sitemap

import (
	"encoding/xml"
	"time"
)

// W3CTime is a time.Time that marshals to the W3C Datetime format used in sitemaps.
// It uses second precision (no nanoseconds) as required by https://www.sitemaps.org/protocol.html.
type W3CTime struct {
	time.Time
}

// NewW3CTime creates a new W3CTime from a time.Time value.
func NewW3CTime(t time.Time) *W3CTime {
	return &W3CTime{t}
}

// MarshalXML implements xml.Marshaler.
func (w W3CTime) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(w.Time.Truncate(time.Second).Format(time.RFC3339), start)
}

// UnmarshalXML implements xml.Unmarshaler.
func (w *W3CTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// fallback: date-only format
		t, err = time.Parse("2006-01-02", s)
		if err != nil {
			return err
		}
	}
	w.Time = t
	return nil
}

// ChangeFreq defines the change frequency of a URL.
type ChangeFreq string

const (
	ChangeFreqAlways  ChangeFreq = "always"
	ChangeFreqHourly  ChangeFreq = "hourly"
	ChangeFreqDaily   ChangeFreq = "daily"
	ChangeFreqWeekly  ChangeFreq = "weekly"
	ChangeFreqMonthly ChangeFreq = "monthly"
	ChangeFreqYearly  ChangeFreq = "yearly"
	ChangeFreqNever   ChangeFreq = "never"
)

// URL represents a single entry in the sitemap.
type URL struct {
	// Loc is the full URL of the page.
	Loc string `xml:"loc"`
	// LastMod is the date of the last modification in W3C Datetime format (second precision).
	LastMod *W3CTime `xml:"lastmod,omitempty"`
	// ChangeFreq indicates how frequently the page is likely to change.
	ChangeFreq ChangeFreq `xml:"changefreq,omitempty"`
	// Priority is the priority of the URL relative to other URLs (0.0 to 1.0).
	Priority float64 `xml:"priority,omitempty"`
}

// Sitemap represents a complete XML sitemap according to the Sitemaps protocol.
// See https://www.sitemaps.org/protocol.html
type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

// NewSitemap creates a new sitemap with the default namespace.
func NewSitemap() *Sitemap {
	return &Sitemap{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
	}
}

// AddURL adds a URL entry to the sitemap.
func (s *Sitemap) AddURL(url URL) {
	s.URLs = append(s.URLs, url)
}

// Index represents a sitemap index that references multiple sitemaps.
type Index struct {
	XMLName  xml.Name `xml:"sitemapindex"`
	Xmlns    string   `xml:"xmlns,attr"`
	Sitemaps []Entry  `xml:"sitemap"`
}

// Entry represents a single entry in the sitemap index.
type Entry struct {
	// Loc is the full URL of the sitemap file.
	Loc string `xml:"loc"`
	// LastMod is the date of the last modification of the sitemap file.
	LastMod *W3CTime `xml:"lastmod,omitempty"`
}

// NewIndex creates a new sitemap index with the default namespace.
func NewIndex() *Index {
	return &Index{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
	}
}

// AddSitemap adds a sitemap entry to the index.
func (si *Index) AddSitemap(entry Entry) {
	si.Sitemaps = append(si.Sitemaps, entry)
}
