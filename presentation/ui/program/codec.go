// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package program

import (
	"fmt"
	"math"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/presentation/proto"
)

// encodeValue converts a neutral value of the boundary, see [gostrict.ValueKind], into the protocol. Lists,
// records and maps are lists, so the value needs no type.
func encodeValue(v any) (proto.GostrictValue, error) {
	switch v := v.(type) {
	case bool:
		return &proto.GostrictBool{V: proto.Bool(v)}, nil
	case int32:
		return &proto.GostrictInt{V: proto.Int(v)}, nil
	case uint32:
		return &proto.GostrictInt{V: proto.Int(v)}, nil
	case float64:
		if v == 0 && math.Signbit(v) {
			return &proto.GostrictNegZero{}, nil
		}
		return &proto.GostrictFloat{V: proto.Float(v)}, nil
	case string:
		return &proto.GostrictStr{V: proto.Str(v)}, nil
	case []any:
		values, err := encodeValues(v)
		if err != nil {
			return nil, err
		}
		return &proto.GostrictList{V: values}, nil
	case gostrict.Variant:
		fields, err := encodeValues(v.Fields)
		if err != nil {
			return nil, err
		}
		return &proto.GostrictVariant{Tag: proto.Str(v.Tag), Fields: fields}, nil
	}
	return nil, fmt.Errorf("%T can not cross the boundary of a program in the frontend", v)
}

func encodeValues(values []any) (proto.GostrictValues, error) {
	res := make(proto.GostrictValues, len(values))
	for i, v := range values {
		e, err := encodeValue(v)
		if err != nil {
			return nil, err
		}
		res[i] = e
	}
	return res, nil
}

// decoder converts the values of the protocol into neutral values with the schema of a bundle.
type decoder struct {
	b *gostrict.Bundle
}

func (d decoder) value(v proto.GostrictValue, t gostrict.Type) (any, error) {
	switch t.Kind {
	case gostrict.ValueBool:
		var b bool
		if x, ok := v.(*proto.GostrictBool); ok {
			b = bool(x.V)
		}
		return b, nil
	case gostrict.ValueInt32:
		return int32(d.int(v)), nil
	case gostrict.ValueUint32:
		return uint32(d.int(v)), nil
	case gostrict.ValueFloat64:
		switch x := v.(type) {
		case *proto.GostrictFloat:
			return float64(x.V), nil
		case *proto.GostrictNegZero:
			return math.Copysign(0, -1), nil
		}
		return 0.0, nil
	case gostrict.ValueString:
		var s string
		if x, ok := v.(*proto.GostrictStr); ok {
			s = string(x.V)
		}
		return s, nil
	case gostrict.ValueList, gostrict.ValueOut:
		items := d.list(v)
		res := make([]any, len(items))
		for i, item := range items {
			e, err := d.value(item, *t.Elem)
			if err != nil {
				return nil, fmt.Errorf("element %d: %w", i, err)
			}
			res[i] = e
		}
		return res, nil
	case gostrict.ValueMap:
		items := d.list(v)
		res := make([]any, len(items))
		for i, item := range items {
			kv := d.list(item)
			if len(kv) != 2 {
				return nil, fmt.Errorf("entry %d is no pair", i)
			}
			k, err := d.value(kv[0], *t.Key)
			if err != nil {
				return nil, err
			}
			e, err := d.value(kv[1], *t.Elem)
			if err != nil {
				return nil, err
			}
			res[i] = []any{k, e}
		}
		return res, nil
	case gostrict.ValueRecord:
		return d.record(t.Name, d.list(v))
	case gostrict.ValueUnion:
		x, ok := v.(*proto.GostrictVariant)
		if !ok {
			return nil, fmt.Errorf("expected a variant of %s", t.Name)
		}
		tag := string(x.Tag)
		if !d.isVariant(t.Name, tag) {
			return nil, fmt.Errorf("%s is no variant of %s", tag, t.Name)
		}
		fields, err := d.record(tag, x.Fields)
		if err != nil {
			return nil, err
		}
		return gostrict.Variant{Tag: tag, Fields: fields}, nil
	}
	return nil, fmt.Errorf("%s can not cross the boundary of a program in the frontend", t)
}

func (d decoder) record(name string, values proto.GostrictValues) ([]any, error) {
	rec, ok := d.b.Records[name]
	if !ok {
		return nil, fmt.Errorf("unknown record %s", name)
	}
	if len(values) != len(rec.Fields) {
		return nil, fmt.Errorf("%s has %d fields, got %d values", name, len(rec.Fields), len(values))
	}
	res := make([]any, len(rec.Fields))
	for i, f := range rec.Fields {
		v, err := d.value(values[i], f.Type)
		if err != nil {
			return nil, fmt.Errorf("field %s of %s: %w", f.Name, name, err)
		}
		res[i] = v
	}
	return res, nil
}

func (d decoder) isVariant(union, tag string) bool {
	for _, v := range d.b.Unions[union].Variants {
		if v == tag {
			return true
		}
	}
	return false
}

func (d decoder) int(v proto.GostrictValue) int64 {
	if x, ok := v.(*proto.GostrictInt); ok {
		return int64(x.V)
	}
	return 0
}

func (d decoder) list(v proto.GostrictValue) proto.GostrictValues {
	if x, ok := v.(*proto.GostrictList); ok {
		return x.V
	}
	return nil
}

// message decodes a message of the program, which is one of the records, which its exports return in a msg.Out.
func (d decoder) message(m proto.GostrictMessage) (gostrict.Message, error) {
	name := string(m.Type)
	if !d.isOutgoing(name) {
		return gostrict.Message{}, fmt.Errorf("the program does not send the message %s", name)
	}
	fields, err := d.record(name, m.Fields)
	if err != nil {
		return gostrict.Message{}, err
	}
	return gostrict.Message{Type: name, Fields: fields}, nil
}

// isOutgoing tells, whether an export of the program returns the message, so that the frontend can not make the
// backend decode an arbitrary record.
func (d decoder) isOutgoing(name string) bool {
	for _, f := range d.b.Exports {
		for _, r := range f.Results {
			if r.Kind != gostrict.ValueOut {
				continue
			}
			if r.Elem.Name == name || (r.Elem.Kind == gostrict.ValueUnion && d.isVariant(r.Elem.Name, name)) {
				return true
			}
		}
	}
	return false
}

func encodeMessages(msgs []gostrict.Message) (proto.GostrictMessages, error) {
	res := make(proto.GostrictMessages, len(msgs))
	for i, m := range msgs {
		fields, err := encodeValues(m.Fields)
		if err != nil {
			return nil, fmt.Errorf("message %s: %w", m.Type, err)
		}
		res[i] = proto.GostrictMessage{Type: proto.Str(m.Type), Fields: fields}
	}
	return res, nil
}
