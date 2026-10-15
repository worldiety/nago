package bigs

import (
	"go.wdy.de/nago/pkg/gostrict/std/big"
	"go.wdy.de/nago/pkg/gostrict/std/decimal"
)

type Amount struct {
	Value decimal.Dec
	Unit  string
}

func Equal(a, b big.Int, x, y decimal.Dec, p, q Amount) bool {
	return a == b || // want GS041
		x != y || // want GS041
		p == q || // want GS041
		big.Cmp(a, b) == 0 || decimal.Cmp(x, y) == 0
}

func Zero() decimal.Dec {
	var d decimal.Dec // want GS032
	return d
}

func Unchecked(a, b big.Int) big.Int {
	q, _ := big.Quo(a, b) // want GS033
	return q
}
