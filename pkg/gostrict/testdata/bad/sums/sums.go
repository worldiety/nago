package sums

type Shape interface{ isShape() }

type Circle struct{ R float64 }

type Rect struct{ W, H float64 }

type Square struct{ A float64 }

func (Circle) isShape()  {}
func (Rect) isShape()    {}
func (*Square) isShape() {}

type Named interface{ Name() string }

type Level int32

const (
	Low Level = iota
	Mid
	High
)

func Missing(s Shape) float64 {
	switch s := s.(type) { // want GS038
	case Circle:
		return s.R
	}
	return 0
}

func WithDefault(s Shape) float64 {
	switch s.(type) {
	case Circle:
		return 1
	default:
		return 0
	}
}

func NotVariant(s Shape) float64 {
	switch s.(type) {
	case Circle, Rect:
		return 1
	case *Square: // want GS013
		return 2
	}
	return 0
}

func Open(n Named) string {
	switch n.(type) { // want GS013
	}
	return n.Name()
}

func Assert(s Shape) float64 {
	c := s.(Circle) // want GS013
	return c.R
}

func Levels(l Level) string {
	switch l { // want GS038
	case Low, Mid:
		return "ok"
	}
	return "high"
}

func AllLevels(l Level) string {
	switch l {
	case Low:
		return "low"
	case Mid, High:
		return "up"
	}
	return ""
}
