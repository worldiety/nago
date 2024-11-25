package crud

import (
	"fmt"
	"go.wdy.de/nago/image"
	http_image "go.wdy.de/nago/image/http"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
)

type PickOneImageStyle int

const (
	PickOneImageStyleTeaser PickOneStyle = iota
	PickOneImageStyleAvatar
)

type PickOneImageOptions[E, T any] struct {
	Label        string
	Style        PickOneImageStyle // Default is PickOneImageStyleTeaser
	Paraphe      func(E) string
	CreateSrcSet image.CreateSrcSet
}

// PickOneImage binds a single field of an arbitrary string type (which will be semantically
// a [image.ID]) to an upload option.
func PickOneImage[E any, T ~string](opts PickOneImageOptions[E, T], property Property[E, std.Option[T]]) Field[E] {
	if opts.Paraphe == nil {
		opts.Paraphe = func(e E) string {
			type paraphe interface {
				Paraphe() string
			}

			if p, ok := any(e).(paraphe); ok {
				return p.Paraphe()
			}

			return fmt.Sprintf("%v", e)
		}
	}

	return Field[E]{
		Label: opts.Label,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[image.ID](self.Window, self.ID+"-form.local").Init(func() image.ID {
				var tmp E
				tmp = entity.Get()
				return image.ID(property.Get(&tmp).UnwrapOr(""))
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue image.ID) {
				var tmp E
				tmp = entity.Get()
				oldValue := property.Get(&tmp)
				if newValue == "" {
					property.Set(&tmp, std.None[T]())
				} else {
					property.Set(&tmp, std.Some[T](T(newValue)))
				}

				entity.Set(tmp)
				if image.ID(oldValue.UnwrapOr("")) != newValue {
					entity.Notify()
				}

				handleValidation(self, entity, errState)
			})

			entity.Observe(func(newValue E) {
				tmp := entity.Get()
				v := property.Get(&tmp).UnwrapOr("")
				state.Set(image.ID(v))
				state.Notify()
			})

			if self.requiresValidation() {
				state.Notify()
			}

			return avatarPicker(self.Window, opts.CreateSrcSet, self.ID, state.Get(), state, opts.Paraphe(entity.Get()))
		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			tmp := entity.Get()
			v := property.Get(&tmp)
			id := image.ID(v.UnwrapOr(""))
			var img core.View
			if id != "" {
				// TODO replace me with source set due to different density problem
				uri := core.URI(http_image.NewURL(http_image.Endpoint, id, image.FitCover, 32, 32))
				img = avatar.URI(uri).Size(ui.L32)
			} else {
				img = avatar.Text(opts.Paraphe(tmp)).Size(ui.L32)
			}

			return ui.TableCell(img)
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			var tmp E
			tmp = entity.Get()
			v := property.Get(&tmp)
			id := image.ID(v.UnwrapOr(""))
			var img core.View
			if id != "" {
				// TODO replace me with source set due to different density problem
				uri := core.URI(http_image.NewURL(http_image.Endpoint, id, image.FitCover, 144, 144))
				img = avatar.URI(uri).Size(ui.L144)
			} else {
				img = avatar.Text(opts.Paraphe(tmp)).Size(ui.L144)
			}

			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				img,
			).Alignment(ui.Trailing)
		},
		Comparator: nil,
		Stringer: func(e E) string {
			return string(property.Get(&e).UnwrapOr(""))
		},
	}
}

func avatarPicker(wnd core.Window, setCreator image.CreateSrcSet, selfId string, id image.ID, state *core.State[image.ID], paraphe string) ui.DecoredView {
	if setCreator == nil {
		fn, ok := core.SystemService[image.CreateSrcSet](wnd.Application())
		if !ok {
			panic("image.CreateSrcSet not available")
		}

		setCreator = fn
	}

	var img core.View
	if id != "" {
		// TODO replace me with source set due to different density problem
		uri := core.URI(http_image.NewURL(http_image.Endpoint, id, image.FitCover, 144, 144))
		img = avatar.URI(uri).Size(ui.L120)
	} else {
		img = avatar.Text(paraphe).Size(ui.L120)
	}

	return ui.Box(ui.BoxLayout{

		Center: img,
		BottomTrailing: ui.HStack(ui.ImageIcon(heroOutline.Plus).StrokeColor(ui.ColorBlack).Frame(ui.Frame{}.FullWidth())).
			Action(func() {
				wnd.ImportFiles(core.ImportFilesOptions{
					ID:               selfId + "-upload",
					Multiple:         false,
					AllowedMimeTypes: []string{"image/png", "image/jpeg"},
					OnCompletion: func(files []core.File) {
						if len(files) == 0 {
							// cancel, bug
							return
						}

						if setCreator == nil {
							alert.ShowBannerMessage(wnd, alert.Message{Title: "implementation error", Message: "SrcSet creator has not been set"})
							return
						}

						srcSet, err := setCreator(wnd.Subject(), image.Options{}, files[0])
						if err != nil {
							alert.ShowBannerError(wnd, err)
							return
						}

						// update our state
						state.Set(srcSet.ID)
						state.Notify()
					},
				})
			}).
			BackgroundColor(ui.ColorWhite).
			Frame(ui.Frame{}.Size(ui.L32, ui.L32)).
			Padding(ui.Padding{}.All(ui.L2)).
			Border(ui.Border{}.Width(ui.L4).Circle().Color(ui.ColorBlack)),
	}).Frame(ui.Frame{}.Size(ui.L120, ui.L120))
}
