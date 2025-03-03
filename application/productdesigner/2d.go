package productdesigner

type Layer struct {
	Name    string
	Objects []Positionable
}

type Positionable interface{}
