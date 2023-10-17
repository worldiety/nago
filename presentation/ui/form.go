package ui

import "go.wdy.de/nago/container/slice"

type Form struct {
	Views slice.Slice[InputType]
}

func (Form) isView() {}
