package enum

import (
	"encoding/json"
	"fmt"
	"iter"
	"reflect"
	"slices"
	"strconv"
	"sync"
)

var declaredEnumTypes = map[reflect.Type]enum{}
var mutex sync.RWMutex

type encodingType int

const (
	encodeExternally encodingType = iota
	encodeAdjacent
)

type enum struct {
	variants                  []reflect.Type // enumeration = declaration order
	variantTypeFromString     map[string]reflect.Type
	variantStringFromType     map[reflect.Type]string
	adjacentNameDiscriminator string
	adjacentNameContent       string
	encoding                  encodingType
}

type Enum[Iface any] struct {
	enum enum
}

func (b Enum[Iface]) Types() iter.Seq[reflect.Type] {
	return slices.Values(b.enum.variants)
}

// Declare specifies at runtime a sum type based on a (marker) interface type. The actual members are defined
// through a (anonymous) struct type whose fields denote the enumerated members of the tagged union. This declaration
// must ever occur once per Interface type and as early as possible, thus probably at package level.
// Afterward, any [Box] type can be used to accept only one of the allowed types, even for json unmarshalling.
// Note, that you can use the following field tags for customization:
//   - a _ field may contain the tags encoding, tag and content. Encoding may be of one adjacent|external.
//   - tagValue for a different name, instead of the bare type name
func Declare[Interface any, MatchT any]() Enum[Interface] {
	mutex.Lock()
	defer mutex.Unlock()

	sumT := reflect.TypeFor[Interface]()
	if sumT.Kind() != reflect.Interface {
		panic(fmt.Errorf("type %v must be an interface type", sumT))
	}

	enumT := reflect.TypeFor[MatchT]()
	if enumT.Kind() != reflect.Struct {
		panic(fmt.Errorf("enum stereotype %v must be a struct type", enumT))
	}

	e := enum{
		variantStringFromType: make(map[reflect.Type]string),
		variantTypeFromString: make(map[string]reflect.Type),
	}
	for _, field := range reflect.VisibleFields(enumT) {
		if field.Name == "_" {
			switch field.Tag.Get("encoding") {
			case "adjacent":
				e.encoding = encodeAdjacent
				e.adjacentNameDiscriminator = field.Tag.Get("tag")
				e.adjacentNameContent = field.Tag.Get("content")
				if e.adjacentNameDiscriminator == "" {
					e.adjacentNameDiscriminator = "tag"
				}

				if e.adjacentNameContent == "" {
					e.adjacentNameContent = "content"
				}
			case "":
				fallthrough
			case "external":
				e.encoding = encodeExternally
			default:
				panic(fmt.Errorf("unknown encoding: %v", field.Tag.Get("encoding")))
			}
		}

		if !field.IsExported() {
			continue
		}

		if isInvalidKind(field.Type) {
			panic(fmt.Errorf("the kind %v of %v is not allowed as an enum variant", field.Type.Kind(), field.Type))
		}

		if !field.Type.Implements(sumT) {
			panic(fmt.Errorf("type %v does not implement %v but enum declaration requires it", field.Type, sumT))
		}

		name := field.Name
		if n := field.Tag.Get("tagValue"); n != "" {
			name = n
		}

		if name == "" {
			name = field.Type.Name()
		}

		if t, ok := e.variantTypeFromString[name]; ok {
			panic(fmt.Errorf("the variant name '%s' is already declared by %v", name, t))
		}

		if n, ok := e.variantStringFromType[field.Type]; ok {
			panic(fmt.Errorf("the same variant type '%s' is already declared by %v", field.Type, n))
		}

		e.variantTypeFromString[name] = field.Type
		e.variantStringFromType[field.Type] = name
		e.variants = append(e.variants, field.Type)
	}

	declaredEnumTypes[sumT] = e

	return Enum[Interface]{enum: e}
}

func Make[Interface any](v Interface) Box[Interface] {
	ifaceType := reflect.TypeFor[Interface]()
	if ifaceType.Kind() != reflect.Interface {
		panic(fmt.Errorf("type %v must be an interface type", ifaceType))
	}

	b := Box[Interface]{v: v}
	if !b.Valid() {
		panic(fmt.Errorf("type %T is not a valid member of %v", v, ifaceType))
	}

	return b
}

type Box[Interface any] struct {
	v Interface
}

func (b *Box[Interface]) UnmarshalJSON(bytes []byte) error {
	ifaceType := reflect.TypeFor[Interface]()
	if ifaceType.Kind() != reflect.Interface {
		panic(fmt.Errorf("type %v must be an interface type", ifaceType))
	}

	mutex.RLock()
	e, ok := declaredEnumTypes[ifaceType]
	mutex.RUnlock()

	if !ok {
		panic(fmt.Errorf("type %v is not declared as enum", ifaceType))
	}

	switch e.encoding {
	case encodeAdjacent:
		return b.UnmarshalJSONAdjacentlyTagged(bytes, e.adjacentNameDiscriminator, e.adjacentNameContent)
	case encodeExternally:
		return b.UnmarshalJSONExternallyTagged(bytes)
	default:
		panic(fmt.Errorf("unknown encoding: %v", e.encoding))
	}
}

func (b Box[Interface]) MarshalJSON() ([]byte, error) {
	ifaceType := reflect.TypeFor[Interface]()
	if ifaceType.Kind() != reflect.Interface {
		panic(fmt.Errorf("type %v must be an interface type", ifaceType))
	}

	mutex.RLock()
	e, ok := declaredEnumTypes[ifaceType]
	mutex.RUnlock()

	if !ok {
		panic(fmt.Errorf("type %v is not declared as enum", ifaceType))
	}

	switch e.encoding {
	case encodeAdjacent:
		return b.MarshalJSONAdjacentlyTagged(e.adjacentNameDiscriminator, e.adjacentNameContent)
	case encodeExternally:
		return b.MarshalJSONExternallyTagged()
	default:
		panic(fmt.Errorf("unknown encoding: %v", e.encoding))
	}
}

// MarshalJSONExternallyTagged writes an externally tagged representation which is compatible with the
// following rust serde configuration:
//
//	#[derive(Serialize, Deserialize)]
func (b Box[Interface]) MarshalJSONExternallyTagged() ([]byte, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	ifaceType := reflect.TypeFor[Interface]()
	decl, ok := declaredEnumTypes[ifaceType]
	if !ok {
		return nil, fmt.Errorf("type %v is not declared as enum", ifaceType)
	}

	if !b.Valid() {
		return nil, fmt.Errorf("type %T is not a valid member of %v", b.v, ifaceType)
	}

	vType := reflect.TypeOf(b.v)

	return json.Marshal(map[string]any{
		decl.variantStringFromType[vType]: b.v,
	})
}

// MarshalJSONAdjacentlyTagged writes an adjacently tagged representation which is compatible with the
// following rust serde configuration:
//
//	#[derive(Serialize, Deserialize)]
//	#[serde(tag = "t", content = "c")]
func (b Box[Interface]) MarshalJSONAdjacentlyTagged(tag, content string) ([]byte, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	ifaceType := reflect.TypeFor[Interface]()
	decl, ok := declaredEnumTypes[ifaceType]
	if !ok {
		return nil, fmt.Errorf("type %v is not declared as enum", ifaceType)
	}

	if !b.Valid() {
		return nil, fmt.Errorf("type %T is not a valid member of %v", b.v, ifaceType)
	}

	vType := reflect.TypeOf(b.v)

	return json.Marshal(map[string]any{
		tag:     decl.variantStringFromType[vType],
		content: b.v,
	})
}

// Ordinal either returns a zero based index of the contained enum variant or -1.
func (b Box[Interface]) Ordinal() int {
	if b.IsZero() {
		return -1
	}

	mutex.RLock()
	defer mutex.RUnlock()

	ifaceType := reflect.TypeFor[Interface]()
	decl, ok := declaredEnumTypes[ifaceType]

	if !ok {
		return -1
	}

	vType := reflect.TypeOf(b.v)

	for i, variant := range decl.variants {
		if vType == variant {
			return i
		}
	}

	return -1
}

// Valid returns true, if the boxed value is not nil and the value is one of the declared types.
func (b Box[Interface]) Valid() bool {
	if b.IsZero() {
		return false
	}

	vType := reflect.TypeOf(b.v)
	if vType == nil {
		return false
	}

	if isInvalidKind(vType) {
		panic(fmt.Errorf("type %v is not allowed as an enum variant", vType))
	}

	return true
}

// Unwrap returns the boxed value and panics if the [Box.IsZero] returns true.
// Use this to apply a type switch on your expected interface members.
func (b Box[Interface]) Unwrap() Interface {
	if b.IsZero() {
		panic(fmt.Errorf("enum box is zero"))
	}

	return b.v
}

// IsZero returns true, if the value is nil.
func (b Box[Interface]) IsZero() bool {
	return any(b.v) == nil
}

func (b *Box[Interface]) UnmarshalJSONAdjacentlyTagged(bytes []byte, tag string, content string) error {
	mutex.RLock()
	defer mutex.RUnlock()

	ifaceType := reflect.TypeFor[Interface]()
	decl, ok := declaredEnumTypes[ifaceType]
	if !ok {
		return fmt.Errorf("type %v is not declared as enum", ifaceType)
	}

	tmp := map[string]json.RawMessage{
		tag:     nil,
		content: nil,
	}

	if err := json.Unmarshal(bytes, &tmp); err != nil {
		return fmt.Errorf("cannot decode adjacent container: %w", err)
	}

	tagValue, err := strconv.Unquote(string(tmp[tag]))
	if err != nil {
		return fmt.Errorf("cannot unquote adjacent tag value '%s': %w", string(tmp[tag]), err)
	}
	vType, ok := decl.variantTypeFromString[tagValue]
	if !ok {
		return fmt.Errorf("tag name '%v' is not defined as an enum variant", tagValue)
	}

	ptrToVal := reflect.New(vType).Interface()
	if err := json.Unmarshal(tmp[content], &ptrToVal); err != nil {
		return fmt.Errorf("type '%v' cannot be unmarshalled", vType)
	}

	b.v = reflect.ValueOf(ptrToVal).Elem().Interface().(Interface)

	return nil
}

func (b *Box[Interface]) UnmarshalJSONExternallyTagged(bytes []byte) error {
	mutex.RLock()
	defer mutex.RUnlock()

	ifaceType := reflect.TypeFor[Interface]()
	decl, ok := declaredEnumTypes[ifaceType]
	if !ok {
		return fmt.Errorf("type %v is not declared as enum", ifaceType)
	}

	tmp := map[string]json.RawMessage{}

	if err := json.Unmarshal(bytes, &tmp); err != nil {
		return fmt.Errorf("cannot decode adjacent container: %w", err)
	}

	var key string
	var val json.RawMessage
	for k, v := range tmp {
		key = k
		val = v
	}

	vType, ok := decl.variantTypeFromString[key]
	if !ok {
		return fmt.Errorf("tag name '%v' is not defined as an enum variant", key)
	}

	ptrToVal := reflect.New(vType).Interface()
	if err := json.Unmarshal(val, &ptrToVal); err != nil {
		return fmt.Errorf("type '%v' cannot be unmarshalled", vType)
	}

	b.v = reflect.ValueOf(ptrToVal).Elem().Interface().(Interface)

	return nil
}

func isInvalidKind(vType reflect.Type) bool {
	// interfaces are not serializable by definition
	return vType.Kind() == reflect.Interface ||

		// we disallow also pointers here, because it may cause multiple unwanted side effects
		//  - we want to encourage an immutable programming style
		//  - spreading pointers in (json) serialization breaks invariants by definition, because content is flattened
		//  - intention of using a pointer as an optional marker is not clear, use an optional type instead
		//  - causing additional indirections and escapes are unnecessary for optional cases
		vType.Kind() == reflect.Ptr ||

		// channels are not serializable by definition
		vType.Kind() == reflect.Chan ||

		// functions are not serializable by definition
		vType.Kind() == reflect.Func
}
