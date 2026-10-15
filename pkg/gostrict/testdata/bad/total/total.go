package total

import (
	"errors"

	"go.wdy.de/nago/pkg/gostrict/std/list"
)

type Node struct {
	Value int32
	Next  *Node
}

type Box struct {
	Items list.List[int32]
}

func Panics() {
	panic("boom") // want GS029
}

func Nil() *Node {
	var n *Node // want GS032
	var b Box   // want GS032
	_ = Box{}   // want GS032
	_, _ = b, n
	_ = list.Make[*Node](3) // want GS032
	return nil              // want GS032
}

func Compare(n *Node) bool {
	return n == nil // want GS032
}

func Named() (n *Node) { // want GS032
	return &Node{} // want GS032
}

func Float(f float64) int32 {
	return int32(f) // want GS031
}

func Shift(x uint32, n uint32) uint32 {
	return x << n // want GS007
}

func Err() error {
	return errors.New("x")
}
