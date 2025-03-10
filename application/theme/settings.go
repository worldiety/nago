package theme

import (
	"github.com/worldiety/enum"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/settings"
)

var _ = enum.Variant[settings.GlobalSettings, Settings](
	enum.Rename[Settings]("nago.theme.settings"),
)

type Settings struct {
	_ any `title:"Theme" description:"Theme und Einstellungen der Corporate Identity."`

	PageLogoLight image.ID `json:"pageLogoLight" label:"Seitenlogo Light-Mode"`
	PageLogoDark  image.ID `json:"pageLogoDark" label:"Seitenlogo Dark-Mode"`
	AppIconLight  image.ID `json:"appIconLight" label:"App Icon Light-Mode"`
	AppIconDark   image.ID `json:"appIconDark" label:"App Icon Dark-Mode"`
}

func (s Settings) GlobalSettings() bool {
	return true
}
