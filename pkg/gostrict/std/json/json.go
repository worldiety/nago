// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package json reads JSON into the types of a program and writes them as JSON, e.g. the body of a response or a request
// of package lib/net:
//
//	type Hour struct {
//		Time string  `json:"time"`
//		Temp float64 `json:"temp"`
//	}
//
//	type Forecast struct {
//		City  string               `json:"city"`
//		Hours list.List[Hour]      `json:"hours"`
//		Alert opt.Option[string]   `json:"alert"`
//		Codes dict.Map[string, int32]
//	}
//
//	f, err := json.Decode[Forecast](res.Body)
//	body, err := json.Encode(f)
//
// The type decides, what is read, so that nothing in between is allocated: bool, string, int32, uint32, int64,
// uint64, float64, big.Int, decimal.Dec, bin.Bytes as standard Base64, list.List and list.Frozen as an array,
// dict.Map with string keys as an object, opt.Option and structs, whose fields are of these types. An exported field
// takes the key of its tag json:"key" or its name exactly, json:"-" leaves it out. A key, which the type does not
// know, is skipped, a missing key or null leaves the zero value, which is an empty list, map or bin.Bytes, 0 of a
// big.Int and of a decimal.Dec and None of an option. An integer takes no fraction and no exponent and must fit its
// type. A decimal.Dec is read exactly, without a float in between, from a number or from a string, which holds a
// number, e.g. 12.30, 1.5e3 or "12.30"; its scale is the number of digits of the fraction less the exponent, at most
// decimal.MaxScale.
//
// Encode writes the same types compactly in UTF-8: the fields of a struct in their order, the entries of a map in the
// order of insertion and None as null. A field with the option json:"key,omitempty" is left out, if its value is
// null, "", [] or {}. A float64 is written like in JavaScript, its shortest text, which reads back the same. A
// decimal.Dec is written as a string like decimal.String, e.g. "12.30", so that no reader loses its digits or its
// scale to a float.
//
// The type may be a type parameter of a generic function or type, e.g. func load[T any](b bin.Bytes) (T, error),
// because each engine translates a generic function for each of its instantiations: every instantiation must name a
// type, which json can read and write, also through further generic functions, and a generic type, whose method reads
// or writes its type parameter, must be instantiated only with such types.
//
//gostrict:header
package json

import "go.wdy.de/nago/pkg/gostrict/std/bin"

// Decode reads the JSON document in data into a value of type T. The error tells what is wrong at which offset of
// data, e.g. "json: expected a string at offset 12", and the value is the zero value of T then.
func Decode[T any](data bin.Bytes) (T, error) {
	return decode[T](data)
}

// Encode writes v as JSON. The error tells about a float64, which is NaN or infinite, because JSON has no form for it,
// and the result is empty then.
func Encode[T any](v T) (bin.Bytes, error) {
	return encode(v)
}
