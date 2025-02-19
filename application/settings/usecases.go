package settings

import "reflect"

type Settings interface {
	Settings() bool // open sum type which can be extended by anyone
	IsZero() bool
}

type Load func(id string, t reflect.Type) (Settings, error)
type Store func(id string, settings Settings) error
