package ui

import "go.wdy.de/nago/container/slice"

// Views is actually only used because Go (1.21) cannot infer that in the polymorphic case properly from the LHS.
func Views(v ...View) slice.Slice[View] {
	return slice.Of(v...)
}
