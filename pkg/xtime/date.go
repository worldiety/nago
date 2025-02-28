package xtime

import (
	"fmt"
	"strings"
	"time"
)

const (
	// GermanDate is dd.MM.yyyy in classical notation.
	GermanDate     = "02.01.2006"
	GermanDateTime = "02.01.2006 um 15:04"
)

// Date represents a day/month/year tuple without any associated timezone.
type Date struct {
	Day   int        // Day of month, offset at 1.
	Month time.Month // Month in year, offset at 1.
	Year  int        // Year like 2024.
}

func (d Date) String() string {
	return fmt.Sprintf("%d.%d.%d", d.Year, d.Month, d.Day)
}

// Time converts this date into the first time value of the determined day within the given time zone.
func (d Date) Time(loc *time.Location) time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, loc)
}

func (d Date) Zero() bool {
	return d == Date{}
}

func (d Date) Format(pattern string) string {
	return d.Time(time.UTC).Format(pattern)
}

func (d Date) After(other Date) bool {
	return d.Time(time.UTC).After(other.Time(time.UTC))
}

// TimeFrame represents a Start/End time interval in timezone less unix epoch.
type TimeFrame struct {
	StartTime UnixMilliseconds // inclusive
	EndTime   UnixMilliseconds // inclusive
	Timezone  Timezone
}

func (i TimeFrame) Duration() time.Duration {
	return time.Duration(i.EndTime-i.StartTime) * time.Millisecond
}

func (i TimeFrame) IsZero() bool {
	return i.StartTime == 0 && i.EndTime == 0
}

func (i TimeFrame) Empty() bool {
	return i.StartTime == i.EndTime
}

func (i TimeFrame) String() string {
	return i.Format(GermanDate)
}

func (i TimeFrame) Format(formatDate string) string {
	tz := i.Timezone.Location()
	start := i.StartTime.Time(tz)
	syear, smonth, sday := start.Date()
	shour := start.Hour()
	smin := start.Minute()

	end := i.EndTime.Time(tz)
	eyear, emonth, eday := end.Date()
	ehour := end.Hour()
	emin := end.Minute()

	var sb strings.Builder
	if syear == eyear && smonth == emonth && sday == eday {
		// we need the date just once
		sb.WriteString(start.Format(formatDate))
		sb.WriteString(fmt.Sprintf(" %02d:%02d - %02d:%02d", shour, smin, ehour, emin))
	} else {
		sb.WriteString(fmt.Sprintf("%s %02d:%02d - %s %02d:%02d", start.Format(formatDate), shour, smin, end.Format(formatDate), ehour, emin))
	}

	return sb.String()
}

// A Timezone represents the time zone identifier like Europe/Berlin
type Timezone string

// Location returns the loadable location. If not loadable, returns UTC.
func (t Timezone) Location() *time.Location {
	if t == "" {
		return time.UTC
	}

	loc, err := time.LoadLocation(string(t))
	if err != nil {
		return time.UTC
	}

	return loc
}
