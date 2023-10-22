package enum

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type E2[T1, T2 any] struct {
	ordinal int
	v       any // either A or B or nil
}

func (e E2[T1, T2]) With1(t1 T1) E2[T1, T2] {
	return E2[T1, T2]{
		ordinal: 1,
		v:       t1,
	}
}

func (e E2[T1, T2]) With2(t2 T2) E2[T1, T2] {
	return E2[T1, T2]{
		ordinal: 2,
		v:       t2,
	}
}

func (e E2[T1, T2]) MarshalJSON() ([]byte, error) {
	switch e.ordinal {
	case 0:
		return json.Marshal(adjacentlyTagged[T1]{})
	case 1:
		var zero T1
		return json.Marshal(adjacentlyTagged[T1]{
			Type:  reflect.TypeOf(zero).Name(),
			Value: e.v.(T1),
		})
	case 2:
		var zero T2
		return json.Marshal(adjacentlyTagged[T2]{
			Type:  reflect.TypeOf(zero).Name(),
			Value: e.v.(T2),
		})
	default:
		panic("unreachable")
	}
}

func (e *E2[T1, T2]) UnmarshalJSON(bytes []byte) error {
	var preflight adjacentlyTaggedPreflight
	if err := json.Unmarshal(bytes, &preflight); err != nil {
		return err
	}

	if preflight.Type == "" {
		e.v = nil
		e.ordinal = 0
		return nil
	}

	var t1 adjacentlyTagged[T1]
	var t2 adjacentlyTagged[T2]

	switch preflight.Type {
	case reflect.TypeOf(t1.Value).Name():
		if err := json.Unmarshal(bytes, &t1); err != nil {
			return nil
		}
		e.v = t1.Value
		e.ordinal = 1
		return nil
	case reflect.TypeOf(t2.Value).Name():
		if err := json.Unmarshal(bytes, &t2); err != nil {
			return nil
		}
		e.v = t2.Value
		e.ordinal = 2
		return nil
	default:
		return fmt.Errorf("invalid enum type: %s", preflight.Type)
	}
}

func (e E2[A, B]) Nil() bool {
	return e.ordinal != 0
}

func Match2[T1, T2, R any](e E2[T1, T2], f1 func(T1) R, f2 func(T2) R) R {
	switch e.ordinal {
	case 1:
		return f1(e.v.(T1))
	case 2:
		return f2(e.v.(T2))
	}
	panic("enum is invalid")
}
