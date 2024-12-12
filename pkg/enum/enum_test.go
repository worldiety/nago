package enum_test

import (
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/pkg/enum"
	"testing"
)

type Currency interface {
	currency()
}

type Dollar int

func (Dollar) currency() {}

type EuroCent int

func (EuroCent) currency() {}

type WdyCoin struct {
}

func (WdyCoin) currency() {}

var CurrencyEnum = enum.Declare[Currency, struct {
	Dollar `tagValue:"$"`
	EuroCent
	WdyCoin
	_ any //`encoding:"adjacent"`
}]()

func ExampleDeclare() {
	enum.Declare[Currency, struct {
		Dollar `tagValue:"$"`
		EuroCent
		WdyCoin
		_ any //`encoding:"adjacent"`
	}]()

	var currency enum.Box[Currency]
	currency = enum.Make[Currency](EuroCent(3))
	fmt.Println(currency)
	// Output: {"EuroCent":3}
}

func TestDeclare(t *testing.T) {

	var currency enum.Box[Currency]

	fmt.Println(currency.Ordinal())
	currency = enum.Make[Currency](EuroCent(3))
	fmt.Println(currency.Ordinal())

	buf, err := json.Marshal(currency)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(buf))

	var other enum.Box[Currency]
	err = json.Unmarshal(buf, &other)
	if err != nil {
		t.Fatal(err)
	}

	if other != currency {
		t.Fatal("should be equal")
	}
	fmt.Println(other)
}
