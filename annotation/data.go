package annotation

import (
	"fmt"
	"go.wdy.de/nago/pkg/xmaps"
	"reflect"
)

var datamodels = xmaps.NewConcurrentMap[reflect.Type, *DataModelBuilder]()

type DataModelBuilder struct {
	typ    reflect.Type
	doc    []DocElem
	names  []string
	entity bool
}

// Synonyms defines additional aliases or alternative names for the default language.
func (b *DataModelBuilder) Synonyms(names ...string) *DataModelBuilder {
	b.names = names
	return b
}

// Data documents a model type either as value or entity/aggregate based on the existence of a [data.Identity] method.
// It panics, if T is not a data type.
// This cannot be narrowed through parameter constraints. Use the companies official language for the default
// documentation. Note, that there may be conflicts between Unternehmenssprache and Betriebssprache.
func Data[T any](doc DocElem, documentation ...DocElem) *DataModelBuilder {
	t := reflect.TypeFor[T]()
	if t.Kind() == reflect.Func || t.Kind() == reflect.Interface || t.Kind() == reflect.Invalid {
		panic(fmt.Sprintf("a usecase must be a data type but found: %s", t.Kind()))
	}

	t.MethodByName("Identity")
	b := &DataModelBuilder{doc: append([]DocElem{doc}, documentation...), typ: t}
	_, loaded := datamodels.LoadOrStore(t, b)
	if loaded {
		panic(fmt.Errorf("data model has already been declared"))
	}

	_, b.entity = t.MethodByName("Identity")

	return b
}
