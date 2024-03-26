// Package rquery contains a reflection based query API to filter structs from a sequence using
// a query language similar to jql et al.
package rquery

import (
	"reflect"
	"strconv"
	"strings"
)

// SimplePredicate creates a simple filter predicate which just splits at any space and each term must apply (and semantic).
// Matches with 2 or less matches are ignored. An empty query returns always true.
func SimplePredicate[T any](query string) func(T) bool {
	query = strings.TrimSpace(query)
	if query == "" {
		return func(a T) bool {
			return true
		}
	}

	terms := strings.Split(strings.ToLower(query), " ")
	var sTerms []string
	for _, term := range terms {
		if len(term) <= 2 {
			continue
		}

		sTerms = append(sTerms, term)
	}
	return func(a T) bool {
		matches := 0
		for _, term := range sTerms {
			if contains(a, term) {
				matches++
			}
		}

		return matches == len(sTerms)
	}
}

func contains(t any, what string) bool {
	rType := reflect.TypeOf(t)
	valOfPtr := reflect.ValueOf(t)

	switch rType.Kind() {
	case reflect.String:
		return strings.Contains(strings.ToLower(valOfPtr.String()), what)
	case reflect.Int:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		return strings.Contains(strconv.FormatInt(valOfPtr.Int(), 10), what)
	case reflect.Pointer:
		return contains(valOfPtr.Elem().Interface(), what)
	case reflect.Slice:
		for i := range valOfPtr.Len() {
			if contains(valOfPtr.Index(i).Interface(), what) {
				return true
			}
		}
	case reflect.Struct:
		for i := range rType.NumField() {
			field := rType.Field(i)
			if !field.IsExported() {
				continue
			}

			if contains(valOfPtr.Field(i).Interface(), what) {
				return true
			}
		}
	default:
		//fmt.Printf("ignoring %T", t)
	}

	return false
}
