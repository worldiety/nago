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
// uint64, float64, big.Int, bin.Bytes as standard Base64, list.List and list.Frozen as an array, dict.Map with
// string keys as an object, opt.Option and structs, whose fields are of these types. An exported field takes the
// key of its tag json:"key" or its name exactly, json:"-" leaves it out. A key, which the type does not know, is
// skipped, a missing key or null leaves the zero value, which is an empty list, map or bin.Bytes, 0 of a big.Int and
// None of an option. An integer takes no fraction and no exponent and must fit its type.
//
// Encode writes the same types compactly in UTF-8: the fields of a struct in their order, the entries of a map in the
// order of insertion and None as null. A field with the option json:"key,omitempty" is left out, if its value is
// null, "", [] or {}. A float64 is written like in JavaScript, its shortest text, which reads back the same.
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
