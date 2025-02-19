package scheduler

import (
	"go.wdy.de/nago/pkg/data"
	"time"
)

type Settings struct {
	ID         ID            `json:"id" visible:"false"`
	StartDelay time.Duration `json:"startDelay,omitempty" label:"Verzögerung nach Systemstart"`
	PauseTime  time.Duration `json:"pauseTime,omitempty" label:"Verzögerung zwischen den Ausführungen"`
	Disabled   bool          `json:"disabled,omitempty" label:"Deaktiviert"`
}

func (s Settings) Identity() ID {
	return s.ID
}

func (s Settings) WithIdentity(id ID) Settings {
	s.ID = id
	return s
}

type SettingsRepository data.Repository[Settings, ID]
