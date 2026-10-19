package fastuse

import (
	"errors"

	"go.wdy.de/nago/pkg/gostrict/std/big"
	"go.wdy.de/nago/pkg/gostrict/testdata/headers/fast"
	"go.wdy.de/nago/pkg/gostrict/testdata/headers/fastrec"
)

var look = fastrec.Look{A: "a", B: "b", C: "c", D: "d", E: "e", F: "f", G: "g", H: "h", I: "i"}

// Run passes r through a function of the program, which changes each kind of its fields, and fails, if fail is
// set.
func Run(r fastrec.Rec, fail bool) (fastrec.Rec, string) {
	res, err := fast.Apply(r, func(r fastrec.Rec, n int32) (fastrec.Rec, error) {
		if fail {
			return r, errors.New("failed " + r.Name)
		}
		r.N += n
		r.U -= 2
		r.F *= 1.5
		r.OK = !r.OK
		r.Name += "!"
		r.Tags.Append("x")
		r.Scores.Set("z", 9)
		r.Amount = big.Mul(r.Amount, big.FromInt32(1000000000))
		r.Inner.X = -r.Inner.X
		r.Look = look
		r.Items.Append(fastrec.Inner{X: 7})
		return r, nil
	})
	if err != nil {
		return r, err.Error()
	}
	return res, "ok"
}
