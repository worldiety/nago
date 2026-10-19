// Package fastrec contains the records, which the fast calls of the test of engine/vm pass.
package fastrec

import (
	"go.wdy.de/nago/pkg/gostrict/std/big"
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

type Inner struct {
	X float64
}

// Look has more fields than fit into registers and never changes, so the vm shares it.
type Look struct {
	A, B, C, D, E, F, G, H, I string
}

type Rec struct {
	Name   string
	Tags   list.List[string]
	Scores dict.Map[string, int32]
	Amount big.Int
	Inner  Inner
	Look   Look
	Items  list.List[Inner]
	N      int32
	U      uint32
	F      float64
	OK     bool
}
