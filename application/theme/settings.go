package theme

import (
	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/presentation/ui"
)

var _ = enum.Variant[settings.GlobalSettings, Settings](
	enum.Rename[Settings]("nago.theme.settings"),
)

type Colors struct {
	Dark  ui.Colors
	Light ui.Colors
}

type Settings struct {
	_ any `title:"Theme" description:"Theme und Einstellungen der Corporate Identity."`

	PageLogoLight image.ID `json:"pageLogoLight" label:"Seitenlogo Light-Mode"`
	PageLogoDark  image.ID `json:"pageLogoDark" label:"Seitenlogo Dark-Mode"`
	AppIconLight  image.ID `json:"appIconLight" label:"App Icon Light-Mode"`
	AppIconDark   image.ID `json:"appIconDark" label:"App Icon Dark-Mode"`

	Colors Colors `json:"colors"` // TODO form.Auto cannot render that today, also this is not wanted from the designers perspective and it affects the global applications
}

func (s Settings) GlobalSettings() bool {
	return true
}
