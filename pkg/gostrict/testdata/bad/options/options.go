package options

import "go.wdy.de/nago/pkg/gostrict/std/opt"

type Node struct {
	Value int32
	Next  opt.Option[*Node]
}

func Last(n *Node) int32 {
	next, _ := n.Next.Get() // want GS033
	return next.Value
}

func Unchecked(n *Node) int32 {
	next, ok := n.Next.Get()
	if ok {
		return next.Value
	}
	return next.Value // want GS033
}

func Literal() opt.Option[int32] {
	return opt.Option[int32]{} // want GS026
}

func Fine(n *Node) int32 {
	// the zero value of the field is empty, so it may be left out
	m := &Node{Value: 1}
	if next, ok := n.Next.Get(); ok {
		return next.Value + m.Value
	}
	return n.Next.Or(m).Value
}
