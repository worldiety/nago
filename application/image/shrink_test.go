// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package image

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"math/rand/v2"
	"runtime"
	"testing"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"golang.org/x/image/draw"
)

func TestShrinkMatchesBiLinear(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	noise := func(pix []uint8) {
		for i := range pix {
			pix[i] = uint8(rng.UintN(256))
		}
	}

	gray := image.NewGray(image.Rect(0, 0, 301, 203))
	noise(gray.Pix)
	ycc := image.NewYCbCr(image.Rect(0, 0, 640, 480), image.YCbCrSubsampleRatio420)
	noise(ycc.Y)
	noise(ycc.Cb)
	noise(ycc.Cr)
	rgba := image.NewRGBA(image.Rect(0, 0, 257, 129))
	for i := 0; i < len(rgba.Pix); i += 4 {
		a := uint8(rng.UintN(256))
		rgba.Pix[i], rgba.Pix[i+1], rgba.Pix[i+2], rgba.Pix[i+3] = uint8(rng.UintN(uint(a)+1)), uint8(rng.UintN(uint(a)+1)), uint8(rng.UintN(uint(a)+1)), a
	}

	for _, c := range []struct {
		src    image.Image
		dw, dh int
	}{
		{gray, 144, 96}, {gray, 32, 32}, {ycc, 320, 240}, {ycc, 152, 112}, {ycc, 40, 32}, {rgba, 128, 64}, {rgba, 64, 32},
	} {
		got := image.NewRGBA(image.Rect(0, 0, c.dw, c.dh))
		shrink(got, c.src)
		want := image.NewRGBA(got.Rect)
		draw.BiLinear.Scale(want, want.Rect, c.src, c.src.Bounds(), draw.Over, nil)
		for i := range got.Pix {
			if d := int(got.Pix[i]) - int(want.Pix[i]); d < -1 || d > 1 {
				t.Fatalf("%T %dx%d: byte %d is %d, BiLinear has %d", c.src, c.dw, c.dh, i, got.Pix[i], want.Pix[i])
			}
		}
	}
}

func TestShrinkCoversEverySourcePixel(t *testing.T) {
	// with an uneven ratio, every output pixel still gets a share of the source and the bottom right
	// pixel of the source ends up in the bottom right of the output
	src := image.NewRGBA(image.Rect(10, 20, 10+101, 20+67))
	src.SetRGBA(10+100, 20+66, color.RGBA{R: 255, A: 255})

	dst := image.NewRGBA(image.Rect(0, 0, 48, 32))
	shrink(dst, src)
	if c := dst.RGBAAt(47, 31); c.R == 0 {
		t.Fatalf("bottom right pixel lost: %v", c)
	}
	if c := dst.RGBAAt(0, 0); c != (color.RGBA{}) {
		t.Fatalf("top left must stay transparent: %v", c)
	}
}

func TestShrinkPremultipliesNRGBA(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	// fully transparent pixels must not tint the result, whatever color they carry
	src.SetNRGBA(1, 0, color.NRGBA{G: 255})
	src.SetNRGBA(0, 1, color.NRGBA{G: 255})
	src.SetNRGBA(1, 1, color.NRGBA{G: 255})

	dst := image.NewRGBA(image.Rect(0, 0, 1, 1))
	shrink(dst, src)
	if c := dst.RGBAAt(0, 0); c != (color.RGBA{R: 64, A: 64}) {
		t.Fatalf("got %v", c)
	}
}

func TestShrinkMatchesAtForYCbCr(t *testing.T) {
	src := image.NewYCbCr(image.Rect(0, 0, 64, 64), image.YCbCrSubsampleRatio420)
	for i := range src.Y {
		src.Y[i] = uint8(i * 7)
	}
	for i := range src.Cb {
		src.Cb[i], src.Cr[i] = uint8(i*3), uint8(255-i)
	}

	fast := image.NewRGBA(image.Rect(0, 0, 32, 32))
	shrink(fast, src)
	slow := image.NewRGBA(image.Rect(0, 0, 32, 32))
	shrink(slow, struct{ image.Image }{src}) // hides the type, uses At
	if !bytes.Equal(fast.Pix, slow.Pix) {
		t.Fatal("fast path for YCbCr differs from At")
	}
}

func TestShrinkNeedsNoBufferOfTheSourceSize(t *testing.T) {
	src := image.NewYCbCr(image.Rect(0, 0, 4800, 6400), image.YCbCrSubsampleRatio420)
	dst := image.NewRGBA(image.Rect(0, 0, 2400, 3200))

	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	shrink(dst, src)
	runtime.ReadMemStats(&after)

	// weights and a few rows of the width of the image; BiLinear needed 2400 × 6400 × 32 bytes
	if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 4<<20 {
		t.Fatalf("shrink allocated %d bytes", allocated)
	}
}

func TestCreateSrcSetBuildsThePyramid(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 500, 300))
	for i := range src.Pix {
		src.Pix[i] = 200
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, src, nil); err != nil {
		t.Fatal(err)
	}

	images := mem.NewBlobStore("images")
	create := NewCreateSrcSet(Options{}, json.NewSloppyJSONRepository[SrcSet, ID](mem.NewBlobStore("sets")), images)
	set, err := create(permission.SU(), Options{}, MemFile{Filename: "a.jpg", Bytes: buf.Bytes()})
	if err != nil {
		t.Fatal(err)
	}

	want := [][2]int{{500, 300}, {248, 144}, {120, 72}, {56, 32}}
	if len(set.Images) != len(want) {
		t.Fatalf("got %d steps: %+v", len(set.Images), set.Images)
	}
	for i, w := range want {
		if set.Images[i].Width != w[0] || set.Images[i].Height != w[1] {
			t.Fatalf("step %d: got %dx%d, want %dx%d", i, set.Images[i].Width, set.Images[i].Height, w[0], w[1])
		}
	}

	// the smallest step keeps the color of the source
	r, err := images.NewReader(t.Context(), string(set.Images[len(set.Images)-1].Data))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Unwrap().Close()
	thumb, err := jpeg.Decode(r.Unwrap())
	if err != nil {
		t.Fatal(err)
	}
	if c, _, _, _ := thumb.At(20, 10).RGBA(); c>>8 < 195 || c>>8 > 205 {
		t.Fatalf("color changed: %d", c>>8)
	}
}
