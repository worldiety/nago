package main

import (
	"bytes"
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	. "go.wdy.de/nago/presentation/ui"
	"testing"
	"time"
)

func BenchmarkCircleImage(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		render()
	}
}

func TestCircleImage(t *testing.T) {
	start := time.Now()
	const max = 100_000
	for i := 0; i < max; i++ {
		render()
	}

	elapsed := time.Since(start)
	fmt.Println("elapsed:", elapsed)
	fmt.Println("per render cycle:", elapsed/max)
}

func render() {
	var view core.View
	view = VStack(
		Image().
			URI("asdf").
			Frame(Frame{}.Size("", L320)),

		CircleImage("asdf").
			AccessibilityLabel("Hummel an Lavendel").
			Padding(Padding{Top: L160.Negate()}),

		VStack(
			Text("Hummel").
				Font(Title),

			HStack(
				Text("WZO Terrasse"),
				Spacer(),
				Text("Oldenburg"),
			).Font(Font{Size: L12}).
				Frame(Frame{}.FullWidth()),

			HLine(),
			Text("Es gibt auch").Font(Title),
			Text("Andere Viecher"),
		).Alignment(Leading).
			Frame(Frame{Width: L320}),
	).
		Frame(Frame{Height: ViewportHeight, Width: Full})

	node := view.Render(dummyCtx{})
	var buf bytes.Buffer
	w := proto.NewBinaryWriter(&buf)
	if err := proto.Marshal(w, node); err != nil {
		panic(err)
	}
}

type dummyCtx struct {
}

func (d dummyCtx) Window() core.Window {
	return nil
}

func (d dummyCtx) MountCallback(f func()) proto.Ptr {
	return 1
}

func (d dummyCtx) Handle(bytes []byte) (hnd proto.Ptr, created bool) {
	return 1, true
}
