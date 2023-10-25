package ui

import "go.wdy.de/nago/container/slice"

// Views is actually only used because Go (1.21) cannot infer that in the polymorphic case properly from the LHS.
func Views(v ...View) slice.Slice[View] {
	return slice.Of(v...)
}

func joinViews(v View, others slice.Slice[View]) []View {
	if v == nil && others.Len() == 0 {
		return nil
	}

	res := make([]View, 0, others.Len()+1)
	if v != nil {
		res = append(res, v)
	}

	res = append(res, slice.UnsafeUnwrap(others)...)
	return res
}

type Iterable[T any] interface {
	Each(f func(idx int, v T))
}

func Map[In, Out any](iter Iterable[In], f func(int, In) Out) slice.Slice[Out] {
	var tmp []Out
	iter.Each(func(idx int, v In) {
		tmp = append(tmp, f(idx, v))
	})

	return slice.Of(tmp...)
}
