package handlers

import "go.wdy.de/nago/pkg/gostrict/std/msg"

type State struct{ N int32 }

type Ping struct{ Seq int32 }

type Pong struct{ Seq int32 }

func OnPing(s *State, p Ping) msg.Out[Pong] {
	out := msg.New[Pong]()
	out.Send(Pong{Seq: p.Seq})
	return out
}

func AlsoOnPing(s *State, p Ping) { // want GS037
	s.N++
}

var outbox = msg.New[Pong]() // want GS036
