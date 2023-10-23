// Code generated by nago/internal/gen.go; DO NOT EDIT.

package enum

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type E3[T1 any, T2 any, T3 any] struct {
	ordinal int
	v       any
}

func (e E3[T1, T2, T3]) With1(t1 T1) E3[T1, T2, T3] {
	return E3[T1, T2, T3]{
		ordinal: 1,
		v:       t1,
	}
}

func (e E3[T1, T2, T3]) With2(t2 T2) E3[T1, T2, T3] {
	return E3[T1, T2, T3]{
		ordinal: 2,
		v:       t2,
	}
}

func (e E3[T1, T2, T3]) With3(t3 T3) E3[T1, T2, T3] {
	return E3[T1, T2, T3]{
		ordinal: 3,
		v:       t3,
	}
}

func (e E3[T1, T2, T3]) Nil() bool {
	return e.ordinal != 0
}

func (e E3[T1, T2, T3]) Ordinal() int {
	return e.ordinal
}

func (e E3[T1, T2, T3]) Unwrap() any {
	return e.v
}

func (e E3[T1, T2, T3]) MarshalJSON() ([]byte, error) {
	switch e.ordinal {
	case 0:
		return json.Marshal(adjacentlyTagged[any]{})

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

	case 3:
		var zero T3
		return json.Marshal(adjacentlyTagged[T3]{
			Type:  reflect.TypeOf(zero).Name(),
			Value: e.v.(T3),
		})

	default:
		panic("unreachable")
	}
}

func (e *E3[T1, T2, T3]) UnmarshalJSON(bytes []byte) error {
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

	var t3 adjacentlyTagged[T3]

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

	case reflect.TypeOf(t3.Value).Name():
		if err := json.Unmarshal(bytes, &t3); err != nil {
			return nil
		}
		e.v = t3.Value
		e.ordinal = 3
		return nil

	default:
		return fmt.Errorf("invalid enum type: %s", preflight.Type)
	}
}

func Match3[T1 any, T2 any, T3 any, R any](e E3[T1, T2, T3], f1 func(T1) R, f2 func(T2) R, f3 func(T3) R) R {
	switch e.ordinal {

	case 1:
		return f1(e.v.(T1))

	case 2:
		return f2(e.v.(T2))

	case 3:
		return f3(e.v.(T3))

	}

	panic("enum is invalid")
}
