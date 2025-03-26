// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

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

func TestChunks(t *testing.T) {
	var collected [][]int
	exp := [][]int{{1, 2, 3}, {4, 5, 6}}
	for chunk := range xiter.Chunks[int](slices.Values([]int{1, 2, 3, 4, 5, 6}), 3) {
		collected = append(collected, slices.Clone(chunk))
	}

	if !reflect.DeepEqual(exp, collected) {
		t.Fatal(collected)
	}

	collected = nil
	exp = [][]int{{1, 2, 3}, {4, 5, 6}, {7}}
	for chunk := range xiter.Chunks[int](slices.Values([]int{1, 2, 3, 4, 5, 6, 7}), 3) {
		collected = append(collected, slices.Clone(chunk))
	}

	if !reflect.DeepEqual(exp, collected) {
		t.Fatal(collected)
	}
}
