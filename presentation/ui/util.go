package ui

import "go.wdy.de/nago/container/slice"

// Views is actually only used because Go (1.21) cannot infer that in the polymorphic case properly from the LHS.
func Views(v ...View) slice.Slice[View] {
	return slice.Of(v...)
}

func joinViews(v View, others slice.Slice[View]) []View {
	if v == nil || others.Len() == 0 {
		return nil
	}

	res := make([]View, 0, others.Len()+1)
	if v != nil {
		res = append(res, v)
	}

	res = append(res, slice.UnsafeUnwrap(others)...)
	return res
}
