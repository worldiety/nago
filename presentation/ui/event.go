package ui

// Event represents a json serializable boxed UI event.
type Event struct {
	typ  string
	data any
}
