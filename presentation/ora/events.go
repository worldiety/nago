package ora

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
		reflect.TypeOf(SetPropertyValueRequested{}),
		reflect.TypeOf(FunctionCallRequested{}),
		reflect.TypeOf(NavigationForwardToRequested{}),
		reflect.TypeOf(NavigationResetRequested{}),
		reflect.TypeOf(NavigationReloadRequested{}),
		reflect.TypeOf(NavigationBackRequested{}),
		reflect.TypeOf(SessionAssigned{}),
		reflect.TypeOf(Ping{}),
		reflect.TypeOf(SendMultipleRequested{}),
		reflect.TypeOf(WindowInfoChanged{}),
		reflect.TypeOf(FileImportRequested{}),
		reflect.TypeOf(ThemeRequested{}),
	}
}

// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type _event interface {
	Acknowledged | EventsAggregated | NewComponentRequested | ComponentInvalidated | ComponentInvalidationRequested | ErrorOccurred | ComponentDestructionRequested | ScopeDestructionRequested | ConfigurationRequested | ConfigurationDefined | SetPropertyValueRequested | FunctionCallRequested | NavigationForwardToRequested | NavigationReloadRequested | NavigationResetRequested | NavigationBackRequested | SessionAssigned | Ping | SendMultipleRequested | WindowInfoChanged
}

func EventTypeDiscriminator(v Event) string {
	t := reflect.TypeOf(v)
	f, ok := t.FieldByName("Type")
	if !ok {
		panic(fmt.Errorf("no Type field: %T", v))
	}

	return f.Tag.Get("value")
}
