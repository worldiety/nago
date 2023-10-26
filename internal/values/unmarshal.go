package values

import (
	"fmt"
	"net/url"
	"reflect"
)

// Unmarshal takes the given values and parses them into the given struct pointer.
func Unmarshal(dst any, values url.Values) error {
	if reflect.ValueOf(dst).Kind() != reflect.Pointer {
		panic("dst must be a pointer")
	}

	typ := reflect.ValueOf(dst).Elem()

	for key, values := range values {
		field := typ.FieldByName(key)
		if !field.IsValid() {
			return fmt.Errorf("type %T does not have expected form field '%s'", dst, key)
		}

		if err := ParseValue(field, values); err != nil {
			return fmt.Errorf("value %v cannot be parsed into field %T.%s: %w", values, dst, key, err)
		}
	}

	return nil
}
