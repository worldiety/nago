package xtime

import "time"

const (
	// GermanDate is dd.MM.yyyy in classical notation.
	GermanDate = "02.01.2006"
)

// Date represents a day/month/year tuple without any associated timezone.
type Date struct {
	Day   int        // Year like 2024.
	Month time.Month // Month in year, offset at 1.
	Year  int        // Day of month, offset at 1.
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
