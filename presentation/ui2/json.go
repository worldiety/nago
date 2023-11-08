package ui2

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"unicode"
)

var jsonRegistry *registry

func init() {
	jsonRegistry = &registry{fqn2json: map[goFQN]jsonName{}}
	r := jsonRegistry
	_ = r
}

type jsonName string
type goFQN string

type registry struct {
	fqn2json map[goFQN]jsonName
}

func register[T any](r *registry, v T) {
	fqn := fqnOf(v)
	r.fqn2json[fqn] = jsonName(reflect.TypeOf(v).Name())
}

func (r *registry) getJsonName(v any) (jsonName, bool) {
	n, ok := r.fqn2json[fqnOf(v)]
	return n, ok
}

func fqnOf(v any) goFQN {
	t := reflect.TypeOf(v)
	return goFQN(t.PkgPath() + "." + t.Name())
}

func getGenericTypeName(v any) (name string, typeParams []string) {
	name = reflect.TypeOf(v).Name()
	if tpstart := strings.Index(name, "["); tpstart > -1 {
		paramsStr := name[tpstart+1:]
		name = name[:tpstart]
		params := strings.Split(paramsStr[:len(paramsStr)-1], ",") // not sure how we would want to express nested type param specs
		return name, params
	}

	return name, nil
}

// marshalJSON inserts a "type" field whose value is the registered jsonName (usually just the Go type-name as is).
// All public fields are marshalled and the field name starts lower case.
func marshalJSON(v any) ([]byte, error) {
	n, tp := getGenericTypeName(v)

	buf := bytes.Buffer{}
	buf.WriteString(`{"type":"`)
	buf.WriteString(n)
	buf.WriteByte('"')
	if len(tp) > 0 {
		buf.WriteString(`,"typeParams":["`)
		buf.WriteString(strings.Join(tp, `","`))
		buf.WriteString(`"]`)
	}
	t := reflect.TypeOf(v)
	rv := reflect.ValueOf(v)
	for i := 0; i < t.NumField(); i++ {

		tf := t.Field(i)
		if !tf.IsExported() {
			continue
		}
		fname := tf.Name
		fname = string(unicode.ToLower(rune(fname[0]))) + fname[1:]
		if j := tf.Tag.Get("json"); j != "" {
			fname = j // TODO this is not correct for order and empty features
		}

		if fname == "-" {
			continue // ignored by spec
		}

		buf.WriteString(",")
		buf.WriteString(`"`)
		buf.WriteString(fname)
		buf.WriteString(`":`)
		fieldBuf, err := json.Marshal(rv.Field(i).Interface())
		if err != nil {
			return nil, err
		}
		buf.Write(fieldBuf)
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}
