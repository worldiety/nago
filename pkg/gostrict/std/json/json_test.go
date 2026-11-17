// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package json_test

import (
	"math"
	"testing"

	"go.wdy.de/nago/pkg/gostrict/std/big"
	"go.wdy.de/nago/pkg/gostrict/std/bin"
	"go.wdy.de/nago/pkg/gostrict/std/decimal"
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/json"
	"go.wdy.de/nago/pkg/gostrict/std/json/core"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/opt"
)

type inner struct {
	N int32 `json:"n"`
}

type doc struct {
	Name    string            `json:"name"`
	ID      int64             `json:"id"`
	Max     uint64            `json:"max"`
	F       float64           `json:"f"`
	Ok      bool              `json:"ok"`
	Tags    list.List[string] `json:"tags"`
	Items   list.List[inner]
	Counts  dict.Map[string, uint32] `json:"counts"`
	Note    opt.Option[string]       `json:"note"`
	Missing opt.Option[int32]        `json:"missing"`
	Data    bin.Bytes                `json:"data"`
	Big     big.Int                  `json:"big"`
	Skip    string                   `json:"-"`
	Nested  inner                    `json:"nested,omitempty"`
	hidden  list.List[string]
}

func decode(t *testing.T, src string) doc {
	t.Helper()
	d, err := json.Decode[doc](bin.FromText(src))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestDecode(t *testing.T) {
	d := decode(t, ` {"name":"aä😀\"","id":-9223372036854775808,"max":18446744073709551615,"f":-0,"ok":true,
		"tags":["x","y"],"Items":[{"n":1},{"n":-2,"z":[1,{"q":null}]}],"counts":{"a":1,"b":2,"a":3},"note":"n","missing":null,
		"data":"AQID","big":-123456789012345678901234567890,"Skip":"no","-":"no","nested":{"n":7},"unknown":{"deep":[true,false,null]}} `)

	if d.Name != "aä😀\"" || d.ID != math.MinInt64 || d.Max != math.MaxUint64 || !math.Signbit(d.F) || !d.Ok {
		t.Fatalf("unexpected scalars %+v", d)
	}
	if d.Tags.Len() != 2 || d.Items.Len() != 2 || opt.OrZero(d.Items.Get(1)).N != -2 || d.Nested.N != 7 || d.Skip != "" {
		t.Fatalf("unexpected lists %+v", d)
	}
	if d.Counts.Len() != 2 || opt.OrZero(d.Counts.Get("a")) != 3 {
		t.Fatalf("unexpected map %v", d.Counts.HostEntries())
	}
	if opt.OrZero(d.Note.Get()) != "n" || d.Missing.IsSome() || d.Data.Hex() != "010203" || d.Big.HostString() != "-123456789012345678901234567890" {
		t.Fatalf("unexpected values %+v", d)
	}
	if d.hidden.Len() != 0 {
		t.Fatal("expected an empty list")
	}
}

func TestDecodeFillsMissing(t *testing.T) {
	d := decode(t, `{}`)
	if d.Tags.Len() != 0 || d.Counts.Len() != 0 || d.Data.Len() != 0 || d.Big.HostString() != "0" || d.Note.IsSome() {
		t.Fatalf("expected empty values, got %+v", d)
	}
	if n := decode(t, `null`); n.Tags.Len() != 0 {
		t.Fatal("expected an empty list for null")
	}
}

func TestDecodeErrors(t *testing.T) {
	tests := []struct {
		src, want string
	}{
		{``, "json: unexpected end of input at offset 0"},
		{`{"id":1.5}`, "json: expected an integer at offset 6"},
		{`{"id":"1"}`, "json: expected an integer at offset 6"},
		{`{"max":-1}`, "json: number out of range at offset 7"},
		{`{"id":9223372036854775808}`, "json: number out of range at offset 6"},
		{`{"f":1e999}`, "json: number out of range at offset 5"},
		{`{"name":1}`, "json: expected a string at offset 8"},
		{`{"name":"a` + "\x01" + `"}`, "json: invalid character in string at offset 10"},
		{`{"name":"\x"}`, "json: invalid escape at offset 9"},
		{`{"tags":["a",1]}`, "json: expected a string at offset 13"},
		{`{"tags":["a",]}`, "json: unexpected character ']' at offset 13"},
		{`{"a":1,}`, "json: unexpected character '}' at offset 7"},
		{`{"ok":tru}`, "json: unexpected character '}' at offset 9"},
		{`{"data":"!!!!"}`, "json: invalid base64 at offset 8"},
		{`{} x`, "json: unexpected data after the value at offset 3"},
		{`{"id":01}`, "json: unexpected character '1' at offset 7"},
		{`{"id":-}`, "json: unexpected character '}' at offset 7"},
		{"\xff", "json: unexpected byte 0xff at offset 0"},
	}
	for _, tt := range tests {
		_, err := json.Decode[doc](bin.FromText(tt.src))
		if err == nil || err.Error() != tt.want {
			t.Errorf("%q: got %v, want %s", tt.src, err, tt.want)
		}
	}
}

func TestDecodeInvalidUTF8(t *testing.T) {
	d, err := json.Decode[string](bin.Of(0x22, 0x61, 0xff, 0xfe, 0x5c, 0x6e, 0xff, 0x22))
	if err != nil || d != "a�\n�" {
		t.Fatalf("got %q %v", d, err)
	}
	lone, err := json.Decode[string](bin.FromText(`"\ud800x\udc00"`))
	if err != nil || lone != "�x�" {
		t.Fatalf("got %q %v", lone, err)
	}
}

func TestDecodeDepth(t *testing.T) {
	deep := ""
	for range 600 {
		deep += "["
	}
	_, err := json.Decode[list.List[string]](bin.FromText(`[` + `{"x":` + deep + `}]`))
	if err == nil || err.Error() != "json: expected a string at offset 1" {
		t.Fatalf("got %v", err)
	}
	type node struct{ X list.List[string] }
	_, err = json.Decode[node](bin.FromText(`{"y":` + deep + `}`))
	if err == nil || err.Error() != "json: nesting too deep at offset 516" {
		t.Fatalf("got %v", err)
	}
}

type out struct {
	Name    string             `json:"name"`
	F       float64            //
	Fs      list.List[float64] `json:"fs"`
	Note    opt.Option[string] `json:"note,omitempty"`
	Missing opt.Option[int32]  `json:"missing"`
	Empty   string             `json:"empty,omitempty"`
	Zero    int32              `json:"zero,omitempty"`
	Tags    list.List[string]  `json:"tags,omitempty"`
	M       dict.Map[string, int64]
	Data    bin.Bytes
	Big     big.Int
	Skip    string `json:"-"`
	hidden  int32
}

func TestEncode(t *testing.T) {
	m := dict.New[string, int64]()
	m.Set("b", 1)
	m.Set("a", math.MinInt64)
	m.Set("b", 3)
	v := out{
		Name: "a\"\\\n\x01\x7fä😀\xff", F: math.Copysign(0, -1),
		Fs:   list.Of(1e21, 1.5e-7, 0.000001, 123456789012345680000, 5e-324, 0.1, -1, 1e-6*0.999),
		Tags: list.New[string](), M: m, Data: bin.Of(1, 2), Big: big.FromInt64(-12345), Skip: "no", hidden: 1,
	}
	b, err := json.Encode(v)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"name":"a\"\\\n\u0001` + "\x7fä😀�" + `","F":-0,"fs":[1e+21,1.5e-7,0.000001,123456789012345680000,5e-324,0.1,-1,9.99e-7],` +
		`"missing":null,"zero":0,"M":{"b":3,"a":-9223372036854775808},"Data":"AQI=","Big":-12345}`
	if got := b.Text(); got != want {
		t.Fatalf("got  %s\nwant %s", got, want)
	}

	back, err := json.Decode[out](b)
	if err != nil || back.Name != "a\"\\\n\x01\x7fä😀�" || !math.Signbit(back.F) || back.Fs.Len() != 8 || opt.OrZero(back.M.Get("a")) != math.MinInt64 {
		t.Fatalf("unexpected round trip %+v %v", back, err)
	}
}

func TestEncodeRejectsNaN(t *testing.T) {
	for f, want := range map[float64]string{math.NaN(): "NaN", math.Inf(1): "+Inf", math.Inf(-1): "-Inf"} {
		b, err := json.Encode(list.Of(1, f))
		if err == nil || err.Error() != "json: unsupported value "+want || b.Len() != 0 {
			t.Errorf("%v: got %q %v", f, b.Text(), err)
		}
	}
}

// The decoder bounds the scale like package decimal does.
func TestDecimalScale(t *testing.T) {
	if core.MaxDecimalScale != decimal.MaxScale {
		t.Fatalf("core.MaxDecimalScale is %d, decimal.MaxScale %d", core.MaxDecimalScale, decimal.MaxScale)
	}
}

// A decimal is read exactly from a number or a string and written as a string with its scale.
func TestDecimal(t *testing.T) {
	type price struct {
		Net   decimal.Dec            `json:"net"`
		Rates list.List[decimal.Dec] `json:"rates"`
	}
	p, err := json.Decode[price](bin.FromText(`{"net":"12.30","rates":[0.1,1.5e3,"-0.005"]}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := decimal.String(p.Net); got != "12.30" {
		t.Fatalf("net = %s", got)
	}
	b, err := json.Encode(p)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := b.Text(), `{"net":"12.30","rates":["0.1","1500","-0.005"]}`; got != want {
		t.Fatalf("encoded %s, want %s", got, want)
	}

	var zero price
	if _, err := json.Decode[price](bin.FromText(`{"net":"1,5"}`)); err == nil || err.Error() != "json: expected a decimal at offset 7" {
		t.Fatalf("err = %v", err)
	}
	if zero, err = json.Decode[price](bin.FromText(`{}`)); err != nil || decimal.String(zero.Net) != "0" || zero.Rates.Len() != 0 {
		t.Fatalf("zero = %+v, %v", zero, err)
	}
}
