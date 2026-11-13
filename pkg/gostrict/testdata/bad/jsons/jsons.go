package jsons

import (
	"go.wdy.de/nago/pkg/gostrict/std/bin"
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/json"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

type Node struct {
	Next *Node
}

type Keys struct {
	ByID dict.Map[int32, string]
}

type Fine struct {
	Name  string            `json:"name"`
	Items list.List[string] `json:"items"`
}

func Pointer(data bin.Bytes) (Node, error) {
	return json.Decode[Node](data) // want GS047
}

func IntKeys(data bin.Bytes) (Keys, error) {
	return json.Decode[Keys](data) // want GS047
}

// A generic function may read a type parameter; each instantiation must name a type, which json can read.
func decodeAs[T any](data bin.Bytes) (T, error) {
	return json.Decode[T](data)
}

func Generic(data bin.Bytes) (Fine, error) {
	return decodeAs[Fine](data)
}

func GenericPointer(data bin.Bytes) (Node, error) {
	return decodeAs[Node](data) // want GS047
}

// The demand passes through further generic functions.
func twice[U any](data bin.Bytes) (U, error) {
	return decodeAs[U](data)
}

func GenericKeys(data bin.Bytes) (Keys, error) {
	return twice[Keys](data) // want GS047
}

func encodeAny[V any](v V) (bin.Bytes, error) {
	return json.Encode(list.Of(v))
}

func Inferred(n Node) (bin.Bytes, error) {
	return encodeAny(n) // want GS047
}

func InferredFine(f Fine) (bin.Bytes, error) {
	return encodeAny(f)
}

// A method of a generic type, which writes its type parameter, demands it of every instantiation of the type.
type Box[T any] struct {
	V T
}

func (b Box[T]) Write() (bin.Bytes, error) {
	return json.Encode(b.V)
}

func BoxOfFine(f Fine) (bin.Bytes, error) {
	return Box[Fine]{V: f}.Write()
}

func BoxOfNode(n Node) (bin.Bytes, error) {
	return Box[Node]{V: n}.Write() // want GS047
}

func Allowed(data bin.Bytes) (Fine, error) {
	return json.Decode[Fine](data)
}

func Write(n Node) (bin.Bytes, error) {
	return json.Encode(n) // want GS047
}

func WriteFine(f Fine) (bin.Bytes, error) {
	return json.Encode(f)
}
