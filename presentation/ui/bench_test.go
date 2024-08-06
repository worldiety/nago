package ui_test

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	ui "go.wdy.de/nago/presentation/ui"
	"testing"
)

type MockCtx struct {
}

func (m MockCtx) MountCallback(f func()) ora.Ptr {
	return 0
}

func render(ctx core.RenderContext) ora.Component {
	return ui.VStack(
		ui.Image().
			Frame(ora.Frame{}.Size("", ora.L320)),

		ui.VStack(
			ui.Text("Hummel").
				Color("black").
				Size("var(--my-base)").
				AccessibilityLabel("blub").
				Frame(ora.Frame{}).
				Padding(ora.Padding{}),

			ui.HStack(func(hstack *ui.THStack) {
				hstack.Frame(ora.Frame{}.FullWidth())
				hstack.Append(
					ui.Text("WZO Terrasse"),
					ui.Spacer(),
					ui.Text("Oldenburg"),
				)
			}),
			ui.HDivider(),
			ui.Text("Andere Viecher"),
			ui.Text("gibt es auch: "),
		).Alignment(ora.Leading).
			BackgroundColor("var(--my-color)").
			Frame(ora.Frame{Width: "400dp"}),
	).
		Frame(ora.Frame{Height: ora.ViewportHeight, Width: ora.Full}).Render(ctx)

}

/*note, that we have not an exploding gc allocation rate due to boxed values in the builder pattern, because
Go will indeed aggressively inline and devirtualize those chain calls:

go build -gcflags '-m' example/cmd/tutorial-02-combining-views/main.go

mple/cmd/tutorial-02-combining-views/main.go:58:11: inlining call to ui.TImage.Frame
example/cmd/tutorial-02-combining-views/main.go:65:13: inlining call to ui.Text
example/cmd/tutorial-02-combining-views/main.go:66:12: inlining call to ui.TText.HSLColor
example/cmd/tutorial-02-combining-views/main.go:67:11: inlining call to ui.TText.Size
example/cmd/tutorial-02-combining-views/main.go:68:25: inlining call to ui.TText.AccessibilityLabel
example/cmd/tutorial-02-combining-views/main.go:69:12: devirtualizing .autotmp_5.Frame to ui.TText
example/cmd/tutorial-02-combining-views/main.go:69:12: inlining call to ui.TText.Frame
example/cmd/tutorial-02-combining-views/main.go:70:14: devirtualizing .autotmp_6.Padding to ui.TText
example/cmd/tutorial-02-combining-views/main.go:70:14: inlining call to ui.TText.Padding
example/cmd/tutorial-02-combining-views/main.go:72:18: inlining call to ui.HStack
*/

func Benchmark(b *testing.B) {
	ctx := MockCtx{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		render(ctx)
	}
}
