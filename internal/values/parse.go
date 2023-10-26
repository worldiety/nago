package values

import (
	"fmt"
	"reflect"
	"strconv"
)

func ParseValue(field reflect.Value, value []string) error {
	switch field.Type().Kind() {
	case reflect.String:
		if len(value) != 1 {
			return fmt.Errorf("string field has no corresponding value")
		}

		field.SetString(value[0])
	case reflect.Float32, reflect.Float64:
		if len(value) != 1 {
			return fmt.Errorf("float64 field has no corresponding value")
		}

		v, err := strconv.ParseFloat(value[0], 64)
		if err != nil {
			return err
		}
		field.SetFloat(v)
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		if len(value) != 1 {
			return fmt.Errorf("int64 field has no corresponding value")
		}

		v, err := strconv.ParseInt(value[0], 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.Bool:
		if len(value) != 1 {
			return fmt.Errorf("bool field has no corresponding value")
		}

		v, err := strconv.ParseBool(value[0])
		if err != nil {
			return err
		}
		field.SetBool(v)
	case reflect.Slice:
		switch field.Type().Elem().Kind() {
		case reflect.String:
			field.Set(reflect.ValueOf(value))
		default:
			return fmt.Errorf("unsupported slice field type: %v", field)
		}
	default:
		return fmt.Errorf("unsupported field type: %v", field)
	}

	return nil
}
