package form

import (
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/alert"
)

func wndImportFiles(wnd core.Window, setCreator image.CreateSrcSet, selfId string, state *core.State[image.ID]) {
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
}
