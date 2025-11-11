// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xjson_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"go.wdy.de/nago/pkg/xjson"
)

type t1 struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type t2 struct {
	Type     string `json:"type"`
	Tool     string `json:"tool"`
	FileId   string `json:"file_id"`
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
}

type box struct {
	Value any
}

var boxVariants = []xjson.VariantOption{xjson.Variant[t1]("text"), xjson.Variant[t2]("tool_file")}

func (b box) MarshalJSON() ([]byte, error) {
	return xjson.MarshalInternally("type", b.Value, boxVariants...)
}

func (b *box) UnmarshalJSON(data []byte) error {
	v, err := xjson.UnmarshalInternally("type", data, boxVariants...)
	b.Value = v
	return err
}

func TestMarshalInternally(t *testing.T) {
	const x0 = `{
      "type" : "text",
      "text" : "Hier ist ein Bild von einem Vogel für dich:\n\n"
    }`

	const x1 = `{
      "type" : "tool_file",
      "tool" : "image_generation",
      "file_id" : "b9116078-b141-40d9-86e3-443a7a6023e5",
      "file_name" : "image_generated_0",
      "file_type" : "png"
    }`

	jsons := []string{x0, x1}

	for _, str := range jsons {
		var tmp box
		if err := json.Unmarshal([]byte(str), &tmp); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%#v\n", tmp.Value)
	}

}

type box2 struct {
	Value []any
}

func (b box2) MarshalJSON() ([]byte, error) {
	return xjson.MarshalInternally("type", b.Value, boxVariants...)
}

func (b *box2) UnmarshalJSON(data []byte) error {
	v, err := xjson.UnmarshalInternally("type", data, boxVariants...)
	if err != nil {
		return err
	}
	b.Value = v.([]any)
	return err
}

func TestMarshalInternally2(t *testing.T) {
	const x0 = ` [ {
      "type" : "text",
      "text" : "Hier ist ein Bild von einem Vogel für dich:\n\n"
    }, {
      "type" : "tool_file",
      "tool" : "image_generation",
      "file_id" : "b9116078-b141-40d9-86e3-443a7a6023e5",
      "file_name" : "image_generated_0",
      "file_type" : "png"
    } ]`

	var tmp box2
	if err := json.Unmarshal([]byte(x0), &tmp); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", tmp.Value)

}
