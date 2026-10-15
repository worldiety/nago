package globals

import (
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

type Point struct{ X, Y float64 }

func (p *Point) Move(dx float64) { p.X += dx }

func (p Point) Len() float64 { return p.X + p.Y }

type Box struct{ Items list.List[int32] }

var limit int32 = 10
var origin = Point{X: 1, Y: 2}
var table = list.Freeze(list.Of[int32](1, 2, 3))
var names = list.Of("a", "b")           // want GS036
var index = dict.New[string, int32]()   // want GS036
var box = Box{Items: list.New[int32]()} // want GS036
var shared = &Point{}                   // want GS036
var counter int32

func Change() float64 {
	counter++       // want GS036
	limit = 20      // want GS036
	origin.X = 3    // want GS036
	origin.Move(1)  // want GS036
	local := origin // a copy may change
	local.Move(1)
	v, _ := table.Get(0)
	return origin.Len() + local.X + float64(v+limit)
}
