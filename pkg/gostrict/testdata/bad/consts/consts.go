package consts

import "go.wdy.de/nago/pkg/gostrict/testdata/bad/consts/hdr"

const title = "Check order"

func Run(s string, n int32, f float64, k hdr.Kind) {
	hdr.Activity("Check order")
	hdr.Activity(title)
	hdr.Activity(title + " again")
	(hdr.Activity)("parenthesized")
	hdr.Activity(s) // want GS043
	hdr.Pair("a", 1, f)
	hdr.Pair(s, n, 1) // want GS043
	hdr.Mark(hdr.KindA)
	hdr.Mark(k) // want GS043
	label := func(x string) {
		hdr.Activity(x) // want GS043
	}
	label("x")

	act := hdr.Activity // want GS044
	act("x")
	apply(hdr.Activity) // want GS044
}

func apply(f func(string)) {
	f("x")
}

//gostrict:const s // want GS028
func Local(s string) {}
