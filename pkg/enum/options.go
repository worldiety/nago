package enum

import "reflect"

type Option interface {
	apply(enumCfg *enumCfg)
}

type optFunc func(enumCfg *enumCfg)

func (f optFunc) apply(enumCfg *enumCfg) {
	f(enumCfg)
}

type enumCfg struct {
	noZero         bool
	sealed         bool
	fromNameToType map[string]reflect.Type
	fromTypeToName map[reflect.Type]string
	jsonOpts       JSON
}

func (c enumCfg) NoZero() bool { return c.noZero }

func (c enumCfg) Sealed() bool { return c.sealed }

func Rename[T any](name string) Option {
	return optFunc(func(enumCfg *enumCfg) {
		t := reflect.TypeFor[T]()
		enumCfg.fromNameToType[name] = t
		enumCfg.fromTypeToName[t] = name
	})
}

func NoZero() Option {
	return optFunc(func(enumCfg *enumCfg) {
		enumCfg.noZero = true
	})
}

func Sealed() Option {
	return optFunc(func(enumCfg *enumCfg) {
		enumCfg.sealed = true
	})
}

func Adjacently(tagName string, contentName string) Option {
	return optFunc(func(enumCfg *enumCfg) {
		enumCfg.jsonOpts = AdjacentlyOptions{
			Tag:     tagName,
			Content: contentName,
		}
	})
}

// Externally is the default
func Externally() Option {
	return optFunc(func(enumCfg *enumCfg) {
		enumCfg.jsonOpts = ExternallyOptions{}
	})
}

func Internally() Option {
	return nil
}

func Untagged() Option {
	return nil
}
