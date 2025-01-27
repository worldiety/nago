package nprotoc

import (
	"github.com/worldiety/enum"
	"iter"
	"maps"
	"slices"
)

type Declaration interface {
	decl()
}

var DeclEnum = enum.Declare[Declaration, func(
	func(Enum),
	func(Uint),
	func(Record),
	func(Project),
	func(Map),
	func(String),
	func(Array),

)](
	enum.NoZero(),
	enum.Internally("kind"),
)

type Typename string
type Enum struct {
	Doc      string     `json:"doc"`
	Kind     string     `json:"kind"`
	Variants []Typename `json:"variants"`
}

func (Enum) decl() {}

type Const struct {
	Name string `json:"name"`
	Doc  string `json:"doc"`
}

type Value string
type Uint struct {
	Doc         string          `json:"doc"`
	Kind        string          `json:"kind"`
	Id          int             `json:"id"`
	ConstValues map[Value]Const `json:"const"`
}

func (d Uint) sortedConst() iter.Seq2[Value, Const] {
	return func(yield func(Value, Const) bool) {
		for _, key := range slices.Sorted(maps.Keys(d.ConstValues)) {
			if !yield(key, d.ConstValues[key]) {
				return
			}
		}
	}

}

func (d Uint) ID() int {
	return d.Id
}

func (Uint) decl() {}

// String has intentionally no const block, because we do not (yet) optimize that. Thus, you are much more efficient
// if you use Uint consts instead.
type String struct {
	Doc  string `json:"doc"`
	Kind string `json:"kind"`
	Id   int    `json:"id"`
}

func (d String) ID() int {
	return d.Id
}

func (String) decl() {}

type Map struct {
	Doc   string   `json:"doc"`
	Kind  string   `json:"kind"`
	Id    int      `json:"id"`
	Key   Typename `json:"key"`
	Value Typename `json:"value"`
}

func (d Map) ID() int {
	return d.Id
}

func (Map) decl() {}

type Array struct {
	Doc  string   `json:"doc"`
	Kind string   `json:"kind"`
	Id   int      `json:"id"`
	Type Typename `json:"type"`
}

func (d Array) ID() int {
	return d.Id
}

func (Array) decl() {}

type FieldID int
type Record struct {
	Doc    string            `json:"doc"`
	Kind   string            `json:"kind"`
	Id     int               `json:"id"`
	Fields map[FieldID]Field `json:"fields"`
}

func (d Record) sortedFields() iter.Seq2[FieldID, Field] {
	return func(yield func(FieldID, Field) bool) {
		for _, k := range slices.Sorted(maps.Keys(d.Fields)) {
			if !yield(k, d.Fields[k]) {
				return
			}
		}
	}
}

func (d Record) fieldCount() int {
	return len(d.Fields)
}

func (d Record) ID() int {
	return d.Id
}

func (Record) decl() {}

type Field struct {
	Doc  string   `json:"doc"`
	Name string   `json:"name"`
	Type Typename `json:"type"`
}

type Project struct {
	Kind string `json:"kind"`
	Go   struct {
		Package string `json:"package"`
	} `json:"go"`
}

func (Project) decl() {}

type IdentityTypeDeclaration interface {
	ID() int
}
