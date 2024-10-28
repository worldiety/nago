package xiter_test

import (
	"go.wdy.de/nago/pkg/xiter"
	"reflect"
	"slices"
	"strconv"
	"testing"
)

func TestFilter(t *testing.T) {
	tmp := []int{1, 2, 3, 4, 5, 6}
	even := slices.Collect(xiter.Filter(func(v int) bool { return v%2 == 0 }, slices.Values(tmp)))
	if !reflect.DeepEqual([]int{2, 4, 6}, even) {
		t.Fatal(even)
	}
}

func TestMap(t *testing.T) {
	tmp := []int{1, 2, 3}
	str := slices.Collect(xiter.Map(func(v int) string { return strconv.Itoa(v) }, slices.Values(tmp)))
	if !reflect.DeepEqual([]string{"1", "2", "3"}, str) {
		t.Fatal(str)
	}
}

func TestLimit(t *testing.T) {
	tmp := []int{1, 2, 3, 4, 5, 6}
	r := slices.Collect(xiter.Limit(slices.Values(tmp), 2))
	if !reflect.DeepEqual([]int{1, 2}, r) {
		t.Fatal(r)
	}
}
