package forbidden

import (
	"fmt" // want GS001
	"strings"

	"go.wdy.de/nago/pkg/gostrict/std/list"
)

type Base struct{ ID int32 }

type Derived struct {
	Base // want GS014
	Name string
}

type Shape interface{ Area() float64 }

func init() {} // want GS016

func Sum(xs ...int32) int32 { return 0 } // want GS015

func Count() int { return 1 } // want GS006

func Slices() {
	xs := []int32{1, 2}     // want GS003
	m := map[string]int32{} // want GS004
	_, _ = xs, m
}

func Strings(s string) int32 {
	n := len(s)           // want GS008
	c := s[0]             // want GS008
	for _, r := range s { // want GS008
		_ = r
	}
	_ = strings.Split(s, ",") // want GS002
	_, _ = n, c
	return 0
}

func Misc(v any, a, b Shape) { // want GS017
	go Misc(v, a, b)    // want GS011
	defer Misc(v, a, b) // want GS012
	switch v.(type) {   // want GS013
	}
	x := Base{}
	p := &x    // want GS009
	_ = a == b // want GS018
	_ = p
	fmt.Println("x")
	_ = list.New[int32]()
	_ = append([]int32{}, 1) // want GS003
}
