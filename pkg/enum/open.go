package enum

import (
	"fmt"
	"reflect"
	"slices"
)

// Variant adds or updates another type variant to an Enum declaration. If no such Enum Declaration exists, a new one
// is created. Panics, if the Enum is sealed. Also, when adding a new variant, all global settings may be changed.
// This is intentional, e.g. to introduce future workarounds. Currently, always nil is returned, to allow
// simple execution at global variable declaration time.
func Variant[Interface any, Member any](opts ...Option) any {
	decl, ok := DeclarationOf(reflect.TypeFor[Interface]())
	if !ok {
		_ = Declare[Interface, func(func(any))](opts...)
		decl, ok = DeclarationOf(reflect.TypeFor[Interface]())
		if !ok {
			panic("can't add variant to an enum")
		}
	}

	mutex.Lock()
	defer mutex.Unlock()

	// add variant
	eType := reflect.TypeFor[Interface]()
	mType := reflect.TypeFor[Member]()
	if !mType.Implements(eType) {
		panic(fmt.Errorf("member type %v does not implement %v", mType, eType))
	}

	decl.variants = slices.DeleteFunc(decl.variants, func(r reflect.Type) bool {
		return mType == r
	})

	name := mType.Name()
	decl.variants = append(decl.variants, mType)
	decl.cfg.fromTypeToName[mType] = name
	decl.cfg.fromNameToType[name] = mType

	// rewrite all settings
	for _, opt := range opts {
		opt.apply(&decl.cfg)
	}

	globalDeclContext.decls[eType] = decl

	return nil
}
