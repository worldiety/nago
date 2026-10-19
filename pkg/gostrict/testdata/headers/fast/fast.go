// Package fast is a header package with a function, which calls a function of the program with records, lists,
// maps and integers of any size, for the test of the fast calls of engine/vm.
//
//gostrict:header
package fast

import "go.wdy.de/nago/pkg/gostrict/testdata/headers/fastrec"

// Apply calls f with r.
//
//gostrict:host
func Apply(r fastrec.Rec, f func(r fastrec.Rec, n int32) (fastrec.Rec, error)) (fastrec.Rec, error) {
	return f(r, 1)
}
