package protocol

import "reflect"

var Events []reflect.Type

func init() {
	Events = []reflect.Type{
		reflect.TypeOf(Acknowledged{}),
		reflect.TypeOf(EventsAggregated{}),
		reflect.TypeOf(NewComponentRequested{}),
		reflect.TypeOf(ComponentInvalidated{}),
	}
}
