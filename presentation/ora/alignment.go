package ora

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
// An empty Alignment must be interpreted as Center (="c").
//
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Alignment string

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
	Top            Alignment = "u"
	Center         Alignment = "c"
	Bottom         Alignment = "b"
	Leading        Alignment = "l"
	Trailing       Alignment = "t"
	TopLeading     Alignment = "ul"
	TopTrailing    Alignment = "ut"
	BottomLeading  Alignment = "bl"
	BottomTrailing Alignment = "bt"
)
