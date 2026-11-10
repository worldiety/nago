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

func Generic[T any](data bin.Bytes) (T, error) {
	return json.Decode[T](data) // want GS047
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
