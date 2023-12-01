package ui

// Color is either a hex color code like #50d71e or an available color name.
// A color name is available, if it has been configured for the entire application but others also depend
// on the actual frontend. For example, a tailwind based frontend has a lot
// of colors build in, like text-rose-700, however an Android frontend does not know that.
type Color string

const (
	PrimaryIntent   Color = "primary"
	SecondaryIntent Color = "secondary"
	SubtileIntent   Color = "subtile"
	Destructive     Color = "destructive"
)
