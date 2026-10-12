// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package colorpicker

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"testing"

	"github.com/worldiety/material-color-utilities/hct"
	"go.wdy.de/nago/presentation/ui"
)

func TestNormalizeHex(t *testing.T) {
	tests := []struct {
		in    string
		alpha bool
		want  ui.Color
		ok    bool
	}{
		{"#EB3F33", true, "#eb3f33", true},
		{"#eb3f33ff", true, "#eb3f33", true},
		{"#eb3f3380", true, "#eb3f3380", true},
		{"#eb3f3380", false, "#eb3f33", true},
		{" #eb3f33 ", true, "#eb3f33", true},
		{"eb3f33", true, "", false},
		{"#eb3", true, "", false},
		{"#eb3f3g", true, "", false},
		{"#+b3f33", true, "", false},
		{"I0", true, "", false},
		{"", true, "", false},
	}

	for _, tt := range tests {
		got, ok := normalizeHex(tt.in, tt.alpha)
		if got != tt.want || ok != tt.ok {
			t.Errorf("normalizeHex(%q, %v) = %q, %v, want %q, %v", tt.in, tt.alpha, got, ok, tt.want, tt.ok)
		}
	}
}

// The frontend calculates HCT with the TypeScript library of Google, from which our Go library has been ported.
// testdata/hct_ts.json has been generated with that library, so the picker shows the same values as the server.
func TestHCTMatchesTypeScript(t *testing.T) {
	buf, err := os.ReadFile("testdata/hct_ts.json")
	if err != nil {
		t.Fatal(err)
	}

	var vectors struct {
		FromHex []struct {
			Hex    string  `json:"hex"`
			Hue    float64 `json:"hue"`
			Chroma float64 `json:"chroma"`
			Tone   float64 `json:"tone"`
		} `json:"fromHex"`
		FromHct []struct {
			Hue    float64 `json:"hue"`
			Chroma float64 `json:"chroma"`
			Tone   float64 `json:"tone"`
			Hex    string  `json:"hex"`
		} `json:"fromHct"`
	}

	if err := json.Unmarshal(buf, &vectors); err != nil {
		t.Fatal(err)
	}

	// the float rounding differs slightly for very dark colors, far below what the picker shows
	const tolerance = 1e-4
	for _, v := range vectors.FromHex {
		rgb, err := strconv.ParseUint(v.Hex[1:], 16, 32)
		if err != nil {
			t.Fatal(err)
		}

		h := hct.FromInt(int(0xff000000 | rgb))
		if math.Abs(h.GetHue()-v.Hue) > tolerance || math.Abs(h.GetChroma()-v.Chroma) > tolerance || math.Abs(h.GetTone()-v.Tone) > tolerance {
			t.Errorf("%s: got HCT %v %v %v, want %v %v %v", v.Hex, h.GetHue(), h.GetChroma(), h.GetTone(), v.Hue, v.Chroma, v.Tone)
		}
	}

	for _, v := range vectors.FromHct {
		h := hct.From(v.Hue, v.Chroma, v.Tone)
		got := fmt.Sprintf("#%06x", h.ToInt()&0xffffff)
		if got != v.Hex {
			t.Errorf("HCT %v %v %v: got %s, want %s", v.Hue, v.Chroma, v.Tone, got, v.Hex)
		}
	}

	if len(vectors.FromHex) == 0 || len(vectors.FromHct) == 0 {
		t.Fatal("no test vectors")
	}
}
