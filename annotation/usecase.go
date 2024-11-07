package annotation

import (
	"fmt"
	"golang.org/x/text/language"
	"reflect"
)

var usecases = map[reflect.Type]*UsecaseBuilder{}

type UsecaseBuilder struct {
	typ        reflect.Type
	doc        []DocElem
	names      []string
	permission perm
}

// Synonyms defines additional aliases or alternative names for the default language.
func (b *UsecaseBuilder) Synonyms(names ...string) *UsecaseBuilder {
	b.names = names
	return b
}

// proposal: LocalizedSynonyms allows to define additional aliases or alternative names for the default language.
func (b *UsecaseBuilder) LocalizedSynonyms(tag language.Tag, names ...string) *UsecaseBuilder {
	panic("signature proposal only")
}

// proposal: Localize allows to define a localized glossary text.
func (b *UsecaseBuilder) Localize(tag language.Tag, doc string) *UsecaseBuilder {
	panic("signature proposal only")
}

// Usecase documents a function type as an Usecase. It panics, if T is not a (named) function type.
// This cannot be narrowed through parameter constraints. Use the companies official language for the default
// documentation. Note, that there may be conflicts between Unternehmenssprache and Betriebssprache.
func Usecase[T any](doc DocElem, documentation ...DocElem) *UsecaseBuilder {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Func {
		panic(fmt.Sprintf("a usecase must be a named func type but found: %s", t.Kind()))
	}

	if _, ok := usecases[t]; ok {
		panic(fmt.Sprintf("a usecase can only be used once: %s", t))
	}

	b := &UsecaseBuilder{doc: append([]DocElem{doc}, documentation...), typ: t}
	usecases[t] = b
	return b
}
