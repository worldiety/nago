package protocol

import (
	"encoding/json"
	"fmt"
	"reflect"
)

var factoryTable map[EventType]func() Event

func init() {
	factoryTable = map[EventType]func() Event{}
	for _, r := range Events {
		if f, ok := r.FieldByName(""); ok {
			factoryTable[EventType(f.Tag.Get("value"))] = func() Event {
				return reflect.New(r).Elem().Interface().(Event)
			}
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

	evtFac := factoryTable[m.Type]
	if evtFac == nil {
		return nil, fmt.Errorf("protocol error: unknown event type %v: %v", m.Type, string(buf))
	}

	evt := evtFac()
	err := json.Unmarshal(buf, &evt)
	return evt, err
}

func Marshal(t Event) []byte {
	buf, err := json.Marshal(t)
	if err != nil {
		panic(fmt.Errorf("unreachable model problem %w", err))
	}

	return buf
}
