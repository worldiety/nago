// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package program

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"go.wdy.de/nago/pkg/gostrict"
)

// The values of a program cross the protocol as JSON, which the browser parses natively. A record is an object
// with the names of its fields, which the translated program reads as they are, and a list is an array. The
// values, which JSON can not express, are objects with a key starting with $, which no field of Go can have:
//
//	{"$m": [[k, v], ...]}       a map in the order of its entries
//	{"$": "pkg.Circle", ...}    a variant of a sealed interface with the fields of its record
//	{"$f": "NaN"}               NaN, "Inf" or "-Inf"
//	{"$e": "message"}           an error, which is not nil
//	{"$r": 7}                   a state, which the program in the browser holds
//
// So the frontend converts the values with a reviver of JSON.parse and a replacer of JSON.stringify and needs no
// schema.

// codec converts the neutral values of a bundle, see [gostrict.ValueKind], from and into JSON.
type codec struct {
	b *gostrict.Bundle
	// state converts a state in the browser, which only the engine of the browser exchanges
	state func(v any) (any, error)
	// token converts a reference of the browser into a token
	token func(ref int64, t gostrict.Type) any
}

// encode returns the JSON of the neutral values of the types.
func (c codec) encode(values []any, types []gostrict.Type) (string, error) {
	tree := make([]any, len(values))
	for i, v := range values {
		e, err := c.tree(v, types[i])
		if err != nil {
			return "", fmt.Errorf("value %d: %w", i, err)
		}
		tree[i] = e
	}
	buf, err := json.Marshal(tree)
	return string(buf), err
}

// encodeRecord returns the JSON of the fields of the record with the full name.
func (c codec) encodeRecord(name string, fields []any) (string, error) {
	e, err := c.tree(fields, gostrict.Type{Kind: gostrict.ValueRecord, Name: name})
	if err != nil {
		return "", err
	}
	buf, err := json.Marshal(e)
	return string(buf), err
}

func (c codec) tree(v any, t gostrict.Type) (any, error) {
	switch t.Kind {
	case gostrict.ValueBool, gostrict.ValueInt32, gostrict.ValueUint32, gostrict.ValueString:
		return v, nil
	case gostrict.ValueFloat64:
		f, _ := v.(float64)
		switch {
		case math.IsNaN(f):
			return map[string]any{"$f": "NaN"}, nil
		case math.IsInf(f, 1):
			return map[string]any{"$f": "Inf"}, nil
		case math.IsInf(f, -1):
			return map[string]any{"$f": "-Inf"}, nil
		}
		return f, nil
	case gostrict.ValueError:
		if v == nil {
			return nil, nil
		}
		return map[string]any{"$e": v.(error).Error()}, nil
	case gostrict.ValueState:
		if c.state == nil {
			return nil, errors.New("a state can not cross the protocol")
		}
		return c.state(v)
	case gostrict.ValueList, gostrict.ValueOut:
		items, _ := v.([]any)
		res := make([]any, len(items))
		for i, item := range items {
			e, err := c.tree(item, *t.Elem)
			if err != nil {
				return nil, err
			}
			res[i] = e
		}
		return res, nil
	case gostrict.ValueMap:
		entries, _ := v.([]any)
		res := make([]any, len(entries))
		for i, e := range entries {
			kv := e.([]any)
			k, err := c.tree(kv[0], *t.Key)
			if err != nil {
				return nil, err
			}
			val, err := c.tree(kv[1], *t.Elem)
			if err != nil {
				return nil, err
			}
			res[i] = []any{k, val}
		}
		return map[string]any{"$m": res}, nil
	case gostrict.ValueRecord:
		return c.recordTree(t.Name, v)
	case gostrict.ValueUnion:
		x, ok := v.(gostrict.Variant)
		if !ok || !c.isVariant(t.Name, x.Tag) {
			return nil, fmt.Errorf("the value is no variant of %s", t.Name)
		}
		obj, err := c.recordTree(x.Tag, x.Fields)
		if err != nil {
			return nil, err
		}
		obj["$"] = x.Tag
		return obj, nil
	}
	return nil, fmt.Errorf("%s can not cross the protocol", t)
}

func (c codec) recordTree(name string, v any) (map[string]any, error) {
	rec, ok := c.b.Records[name]
	if !ok {
		return nil, fmt.Errorf("unknown record %s", name)
	}
	fields, _ := v.([]any)
	if len(fields) != len(rec.Fields) {
		return nil, fmt.Errorf("%s has %d fields, got %d values", name, len(rec.Fields), len(fields))
	}
	obj := make(map[string]any, len(fields))
	for i, f := range rec.Fields {
		e, err := c.tree(fields[i], f.Type)
		if err != nil {
			return nil, fmt.Errorf("field %s of %s: %w", f.Name, name, err)
		}
		obj[f.JS] = e
	}
	return obj, nil
}

// decode parses the JSON array of the values of the types.
func (c codec) decode(src string, types []gostrict.Type) ([]any, error) {
	var tree []any
	if err := json.Unmarshal([]byte(src), &tree); err != nil {
		return nil, err
	}
	if len(tree) != len(types) {
		return nil, fmt.Errorf("expected %d values, got %d", len(types), len(tree))
	}
	res := make([]any, len(types))
	for i, t := range types {
		v, err := c.value(tree[i], t)
		if err != nil {
			return nil, fmt.Errorf("value %d: %w", i, err)
		}
		res[i] = v
	}
	return res, nil
}

// decodeRecord parses the JSON of the record with the full name into its fields.
func (c codec) decodeRecord(name, src string) ([]any, error) {
	var tree any
	if err := json.Unmarshal([]byte(src), &tree); err != nil {
		return nil, err
	}
	return c.record(name, tree)
}

func (c codec) value(x any, t gostrict.Type) (any, error) {
	switch t.Kind {
	case gostrict.ValueBool:
		b, _ := x.(bool)
		return b, nil
	case gostrict.ValueInt32:
		f, _ := x.(float64)
		return int32(f), nil
	case gostrict.ValueUint32:
		f, _ := x.(float64)
		return uint32(f), nil
	case gostrict.ValueFloat64:
		return float(x), nil
	case gostrict.ValueString:
		s, _ := x.(string)
		return s, nil
	case gostrict.ValueError:
		if m, ok := x.(map[string]any); ok {
			msg, _ := m["$e"].(string)
			return errors.New(msg), nil
		}
		return nil, nil
	case gostrict.ValueState:
		m, _ := x.(map[string]any)
		ref, ok := m["$r"].(float64)
		if !ok || c.token == nil {
			return nil, fmt.Errorf("expected a state of %s", t.Name)
		}
		return c.token(int64(ref), t), nil
	case gostrict.ValueList, gostrict.ValueOut:
		items, _ := x.([]any)
		res := make([]any, len(items))
		for i, item := range items {
			e, err := c.value(item, *t.Elem)
			if err != nil {
				return nil, fmt.Errorf("element %d: %w", i, err)
			}
			res[i] = e
		}
		return res, nil
	case gostrict.ValueMap:
		m, _ := x.(map[string]any)
		entries, _ := m["$m"].([]any)
		res := make([]any, len(entries))
		for i, e := range entries {
			kv, ok := e.([]any)
			if !ok || len(kv) != 2 {
				return nil, fmt.Errorf("entry %d is no pair", i)
			}
			k, err := c.value(kv[0], *t.Key)
			if err != nil {
				return nil, err
			}
			val, err := c.value(kv[1], *t.Elem)
			if err != nil {
				return nil, err
			}
			res[i] = []any{k, val}
		}
		return res, nil
	case gostrict.ValueRecord:
		return c.record(t.Name, x)
	case gostrict.ValueUnion:
		m, _ := x.(map[string]any)
		tag, _ := m["$"].(string)
		if !c.isVariant(t.Name, tag) {
			return nil, fmt.Errorf("%q is no variant of %s", tag, t.Name)
		}
		fields, err := c.record(tag, m)
		if err != nil {
			return nil, err
		}
		return gostrict.Variant{Tag: tag, Fields: fields}, nil
	}
	return nil, fmt.Errorf("%s can not cross the protocol", t)
}

// float returns a number or one of the values, which JSON can not express.
func float(x any) float64 {
	switch x := x.(type) {
	case float64:
		return x
	case map[string]any:
		switch x["$f"] {
		case "NaN":
			return math.NaN()
		case "Inf":
			return math.Inf(1)
		case "-Inf":
			return math.Inf(-1)
		case "-0":
			return math.Copysign(0, -1)
		}
	}
	return 0
}

func (c codec) record(name string, x any) ([]any, error) {
	rec, ok := c.b.Records[name]
	if !ok {
		return nil, fmt.Errorf("unknown record %s", name)
	}
	obj, _ := x.(map[string]any)
	res := make([]any, len(rec.Fields))
	for i, f := range rec.Fields {
		v, err := c.value(obj[f.JS], f.Type)
		if err != nil {
			return nil, fmt.Errorf("field %s of %s: %w", f.Name, name, err)
		}
		res[i] = v
	}
	return res, nil
}

func (c codec) isVariant(union, tag string) bool {
	for _, v := range c.b.Unions[union].Variants {
		if v == tag {
			return true
		}
	}
	return false
}

// isOutgoing tells, whether an export of the program returns the message, so that the frontend can not make the
// backend decode an arbitrary record.
func (c codec) isOutgoing(name string) bool {
	for _, f := range c.b.Exports {
		for _, r := range f.Results {
			if r.Kind != gostrict.ValueOut {
				continue
			}
			if r.Elem.Name == name || (r.Elem.Kind == gostrict.ValueUnion && c.isVariant(r.Elem.Name, name)) {
				return true
			}
		}
	}
	return false
}
