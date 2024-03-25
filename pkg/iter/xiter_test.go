package iter_test

import (
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/pkg/slices"
	"reflect"
	"strconv"
	"testing"
)

func TestFilter(t *testing.T) {
	tmp := []int{1, 2, 3, 4, 5, 6}
	even := slices.Collect(iter.Filter(func(v int) bool { return v%2 == 0 }, slices.Values(tmp)))
	if !reflect.DeepEqual([]int{2, 4, 6}, even) {
		t.Fatal(even)
	}
}

func TestMap(t *testing.T) {
	tmp := []int{1, 2, 3}
	str := slices.Collect(iter.Map(func(v int) string { return strconv.Itoa(v) }, slices.Values(tmp)))
	if !reflect.DeepEqual([]string{"1", "2", "3"}, str) {
		t.Fatal(str)
	}
}
