package protocol

import (
	"encoding/json"
	"fmt"
	"reflect"
)

var factoryTable map[EventType]reflect.Type // the interface type must be any within the fat pointer, otherwise json.unmarshal does not work

func init() {
	factoryTable = map[EventType]reflect.Type{}
	for _, r := range Events {
		if f, ok := r.FieldByName("Type"); ok {
			factoryTable[EventType(f.Tag.Get("value"))] = r
		}
	}
}

type msgObj struct {
	Type EventType `json:"type"`
}

func Unmarshal(buf []byte) (Event, error) {
	var m msgObj
	if err := json.Unmarshal(buf, &m); err != nil {
		return nil, err
	}

	typ := factoryTable[m.Type]
	if typ == nil {
		return nil, fmt.Errorf("protocol error: unknown event type %v: %v", m.Type, string(buf))
	}

	value := reflect.New(typ).Elem()
	err := json.Unmarshal(buf, value.Addr().Interface())
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal event: %w: %v", err, string(buf))
	}
	evt := value.Interface()
	if e, ok := evt.(Event); ok {
		return e, nil
	} else {
		return nil, fmt.Errorf("not an Event: %T: %s", evt, string(buf))
	}
}

func Marshal(t Event) []byte {
	buf, err := json.Marshal(t)
	if err != nil {
		panic(fmt.Errorf("unreachable model problem %w", err))
	}

	return buf
}
