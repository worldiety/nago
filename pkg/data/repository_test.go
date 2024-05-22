package data

import (
	"testing"
)

func TestRandIdent(t *testing.T) {
	t.Log(RandIdent[string]())
}

func TestStoid(t *testing.T) {
	type ID1 string
	id1, err := Stoid[ID1]("hello")
	if err != nil {
		t.Fatal(err)
	}
	if string(id1) != "hello" {
		t.Fatal(id1)
	}

	type ID2 int32
	id2, err := Stoid[ID2]("5")
	if err != nil {
		t.Fatal(err)
	}
	if int(id2) != 5 {
		t.Fatal(id2)
	}
}
