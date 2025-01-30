package ui

import "go.wdy.de/nago/presentation/proto"

// Alignment is specified as follows:
//
//	┌─TopLeading───────────Top─────────TopTrailing─┐
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│ Leading            Center            Trailing│
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	└BottomLeading───────Bottom──────BottomTrailing┘
//
// An empty Alignment must be interpreted as Center (="c"), if not otherwise specified.
//
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Alignment uint

func (a Alignment) ora() proto.Alignment {
	return proto.Alignment(a)
}

func Alignments() []Alignment {
	return []Alignment{
		Top,
		Center,
		Bottom,
		Leading,
		Trailing,
		TopLeading,
		TopTrailing,
		BottomLeading,
		BottomTrailing,
	}
}

func (a Alignment) String() string {
	switch a {
	case Top:
		return "top"
	case Bottom:
		return "bottom"
	case Center:
		return "center"
	case Leading:
		return "leading"
	case Trailing:
		return "trailing"
	case TopLeading:
		return "top-leading"
	case BottomLeading:
		return "bottom-leading"
	case TopTrailing:
		return "top-trailing"
	case BottomTrailing:
		return "bottom-trailing"
	default:
		return string(a)
	}
}

const (
	Center         Alignment = Alignment(proto.Center)
	Top            Alignment = Alignment(proto.Top)
	Bottom         Alignment = Alignment(proto.Bottom)
	Leading        Alignment = Alignment(proto.Leading)
	Trailing       Alignment = Alignment(proto.Trailing)
	TopLeading     Alignment = Alignment(proto.TopLeading)
	TopTrailing    Alignment = Alignment(proto.TopTrailing)
	BottomLeading  Alignment = Alignment(proto.BottomLeading)
	BottomTrailing Alignment = Alignment(proto.BottomTrailing)
)
