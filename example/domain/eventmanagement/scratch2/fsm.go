package main

type State interface {
	WorkflowInstance() string
	isState()
}
type InBearbeitung struct{ State }
type Abgeschlossen struct{ State }
type Abgelehnt struct{ State }

func (InBearbeitung) AlleGeräteWurdenNotiert() Abgeschlossen {
	panic("")
}

func (Abgeschlossen) AlexHatAbgelehnt() Abgelehnt {
	panic("")
}

func Transition(from, to, event any) {
	panic("")
}

func main() {
	var state State

	switch t := state.(type) {
	case InBearbeitung:
		state = t.AlleGeräteWurdenNotiert()
	case Abgeschlossen:
		state = t.AlexHatAbgelehnt()
	}
}
