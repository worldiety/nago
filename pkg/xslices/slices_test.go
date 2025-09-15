// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xslices

import (
	"reflect"
	"testing"
)

func TestPrefixSearch(t *testing.T) {
	tests := []struct {
		name     string
		data     []string
		prefix   string
		expected []string
	}{
		{
			name:     "simple match",
			data:     []string{"apple", "applet", "application", "banana", "band", "cat"},
			prefix:   "app",
			expected: []string{"apple", "applet", "application"},
		},
		{
			name:     "prefix with single match",
			data:     []string{"alpha", "beta", "gamma"},
			prefix:   "be",
			expected: []string{"beta"},
		},
		{
			name:     "prefix with no matches",
			data:     []string{"alpha", "beta", "gamma"},
			prefix:   "z",
			expected: []string{},
		},
		{
			name:     "empty prefix returns everything",
			data:     []string{"alpha", "beta", "gamma"},
			prefix:   "",
			expected: []string{"alpha", "beta", "gamma"},
		},
		{
			name:     "prefix matches whole word",
			data:     []string{"alpha", "beta", "gamma"},
			prefix:   "gamma",
			expected: []string{"gamma"},
		},
		{
			name:     "prefix longer than any word",
			data:     []string{"alpha", "beta", "gamma"},
			prefix:   "gammadelta",
			expected: []string{},
		},
		{
			name:     "words with shared prefix",
			data:     []string{"car", "card", "care", "carp", "cat"},
			prefix:   "car",
			expected: []string{"car", "card", "care", "carp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrefixSearch(tt.data, tt.prefix)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("PrefixSearch(%q) = %v, expected %v", tt.prefix, result, tt.expected)
			}
		})
	}
}
