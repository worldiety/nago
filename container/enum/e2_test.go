package enum

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestE2_serEmpty(t *testing.T) {
	var empty E2[string, bool]
	buf, err := json.Marshal(empty)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(buf))

	var read E2[string, bool]
	if err := json.Unmarshal(buf, &read); err != nil {
		t.Fatal(err)
	}

	if empty != read {
		t.Fatalf("should be equal: %v vs %v", empty, read)
	}
}

func Test2_serPrim(t *testing.T) {
	toWrite := E2[string, int]{}.With1("hello world")
	buf, err := json.Marshal(toWrite)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(buf))

	var read E2[string, int]
	if err := json.Unmarshal(buf, &read); err != nil {
		t.Fatal(err)
	}

	if toWrite != read {
		t.Fatalf("should be equal: %v vs %v", toWrite, read)
	}

	if !Match2(toWrite, func(t1 string) bool {
		return true
	}, func(i int) bool {
		return false
	}) {
		t.Fatal("invalid match")
	}
}

func Test2_serPrim2(t *testing.T) {
	toWrite := E2[string, int]{}.With2(1234)
	buf, err := json.Marshal(toWrite)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(buf))

	var read E2[string, int]
	if err := json.Unmarshal(buf, &read); err != nil {
		t.Fatal(err)
	}

	if toWrite != read {
		t.Fatalf("should be equal: %v vs %v", toWrite, read)
	}

	if !Match2(toWrite, func(t1 string) bool {
		return false
	}, func(i int) bool {
		return true
	}) {
		t.Fatal("invalid match")
	}
}

type NotFound string

func (e NotFound) Error() string {
	return fmt.Sprintf("not found: %s", string(e))
}

type Other error

func TestUnwrap(t *testing.T) {
	type FindErrorEnum = E2[NotFound, Other]
	type FindError Error[FindErrorEnum]

	err := IntoErr(FindErrorEnum{}.With1("1234"))
	t.Log(err)
}
