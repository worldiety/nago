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

	PageLogoLight image.ID `label:"Seitenlogo Light-Mode"`
	PageLogoDark  image.ID `label:"Seitenlogo Dark-Mode"`
}

func (s Settings) GlobalSettings() bool {
	return true
}
