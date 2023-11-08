package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"unicode"
)

var jsonRegistry *registry

func init() {
	jsonRegistry = &registry{fqn2json: map[goFQN]jsonName{}}
	r := jsonRegistry
	register(r, HorizontalDivider{})
	register(r, ListItem1L{})
	register(r, ListItem2L{})
	register(r, Scaffold{})
	register(r, MainDetail{})
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

// marshalJSON inserts a "type" field whose value is the registered jsonName (usually just the Go type-name as is).
// All public fields are marshalled and the field name starts lower case.
func marshalJSON(v any) ([]byte, error) {
	n, ok := jsonRegistry.getJsonName(v)
	if !ok {
		return nil, fmt.Errorf("type of %T has not been registered for a json type", v)
	}

	buf := bytes.Buffer{}
	buf.WriteString(`{"type":"`)
	buf.WriteString(string(n))
	buf.WriteByte('"')
	t := reflect.TypeOf(v)
	rv := reflect.ValueOf(v)
	for i := 0; i < t.NumField(); i++ {
		buf.WriteString(",")
		tf := t.Field(i)
		if !tf.IsExported() {
			continue
		}
		fname := tf.Name
		fname = string(unicode.ToLower(rune(fname[0]))) + fname[1:]
		if j := tf.Tag.Get("json"); j != "" {
			fname = j // TODO this is not correct for order and empty features
		}

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
