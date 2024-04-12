package protocol

import (
	"fmt"
	"reflect"
)

var Events []reflect.Type

func init() {
	Events = []reflect.Type{
		reflect.TypeOf(Acknowledged{}),
		reflect.TypeOf(EventsAggregated{}),
		reflect.TypeOf(NewComponentRequested{}),
		reflect.TypeOf(ComponentInvalidated{}),
		reflect.TypeOf(ComponentInvalidationRequested{}),
		reflect.TypeOf(ErrorOccurred{}),
		reflect.TypeOf(ComponentDestructionRequested{}),
		reflect.TypeOf(ScopeDestructionRequested{}),
		reflect.TypeOf(ConfigurationRequested{}),
		reflect.TypeOf(ConfigurationDefined{}),
	}
}

func EventTypeDiscriminator(v Event) string {
	t := reflect.TypeOf(v)
	f, ok := t.FieldByName("Type")
	if !ok {
		panic(fmt.Errorf("no Type field: %T", v))
	}

	return f.Tag.Get("value")
}
