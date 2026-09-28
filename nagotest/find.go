// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagotest

import (
	"fmt"
	"reflect"
	"strings"

	"go.wdy.de/nago/presentation/proto"
)

// A Node is a component within a rendered tree together with its ancestors.
type Node struct {
	// Component is the rendered protocol component, e.g. *proto.TextView.
	Component proto.Component
	// Parents contains all ancestors, the root is first and the direct parent is last.
	Parents []proto.Component
}

// A Matcher decides, if a node is selected.
type Matcher struct {
	desc  string
	match func(n Node) bool
}

func (m Matcher) String() string {
	return m.desc
}

// Where creates a custom matcher, described by desc.
func Where(desc string, match func(n Node) bool) Matcher {
	return Matcher{desc: desc, match: match}
}

// Text matches text views which show exactly the given text.
func Text(text string) Matcher {
	return Where(fmt.Sprintf("Text(%q)", text), func(n Node) bool {
		tv, ok := n.Component.(*proto.TextView)
		return ok && string(tv.Value) == text
	})
}

// TextContains matches text views whose text contains the given text.
func TextContains(text string) Matcher {
	return Where(fmt.Sprintf("TextContains(%q)", text), func(n Node) bool {
		tv, ok := n.Component.(*proto.TextView)
		return ok && strings.Contains(string(tv.Value), text)
	})
}

// Label matches components whose label or accessibility label is exactly the given text, e.g. text fields
// or buttons.
func Label(label string) Matcher {
	return Where(fmt.Sprintf("Label(%q)", label), func(n Node) bool {
		return strField(n.Component, "Label") == label || strField(n.Component, "AccessibilityLabel") == label
	})
}

// ID matches the component with the given ID, see e.g. ui.TStack.ID.
func ID(id string) Matcher {
	return Where(fmt.Sprintf("ID(%q)", id), func(n Node) bool {
		return strField(n.Component, "Id") == id || strField(n.Component, "ID") == id
	})
}

// Type matches all components of type T, e.g. Type[*proto.TextField]().
func Type[T proto.Component]() Matcher {
	var zero T
	return Where(fmt.Sprintf("Type[%T]", zero), func(n Node) bool {
		_, ok := n.Component.(T)
		return ok
	})
}

// TypeName matches all components whose protocol type has the given name, e.g. "TextField". This is intended
// for declarative scenarios, prefer [Type] in Go.
func TypeName(name string) Matcher {
	return Where(fmt.Sprintf("TypeName(%q)", name), func(n Node) bool {
		return reflect.Indirect(reflect.ValueOf(n.Component)).Type().Name() == name
	})
}

// And matches if all given matchers match.
func And(matchers ...Matcher) Matcher {
	var desc []string
	for _, m := range matchers {
		desc = append(desc, m.desc)
	}

	return Where("And("+strings.Join(desc, ", ")+")", func(n Node) bool {
		for _, m := range matchers {
			if !m.match(n) {
				return false
			}
		}

		return true
	})
}

// Containing matches components which have a descendant matching m, e.g. the table row containing a text.
func Containing(m Matcher) Matcher {
	return Where("Containing("+m.desc+")", func(n Node) bool {
		found := false
		walk(n.Component, nil, func(c Node) bool {
			if c.Component != n.Component && m.match(c) {
				found = true
				return false
			}

			return true
		})

		return found
	})
}

// A Selection is the result of a search within the tree of a [Window]. Its assertions fail the test.
type Selection struct {
	w       *Window
	matcher Matcher
	nodes   []Node
}

// Find selects exactly one node matching m and fails the test otherwise.
func (w *Window) Find(m Matcher) Selection {
	w.t.Helper()

	return w.FindAll(m).Exactly(1)
}

// FindAll selects all nodes matching m in document order.
func (w *Window) FindAll(m Matcher) Selection {
	return Selection{w: w, matcher: m, nodes: find(w.Tree(), nil, m)}
}

// Find selects exactly one node matching m within the subtrees of this selection.
func (s Selection) Find(m Matcher) Selection {
	s.w.t.Helper()

	return s.FindAll(m).Exactly(1)
}

// FindAll selects all nodes matching m within the subtrees of this selection.
func (s Selection) FindAll(m Matcher) Selection {
	var nodes []Node
	for _, n := range s.nodes {
		for _, c := range find(n.Component, n.Parents, m) {
			if c.Component != n.Component {
				nodes = append(nodes, c)
			}
		}
	}

	return Selection{w: s.w, matcher: Where(s.matcher.desc+" > "+m.desc, m.match), nodes: nodes}
}

// Exactly asserts that n nodes have been selected.
func (s Selection) Exactly(n int) Selection {
	s.w.t.Helper()

	if len(s.nodes) != n {
		s.w.t.Fatalf("nagotest: %s: expected %d nodes but found %d", s.matcher, n, len(s.nodes))
	}

	return s
}

// None asserts that nothing has been selected.
func (s Selection) None() {
	s.w.t.Helper()

	s.Exactly(0)
}

// Len returns the amount of selected nodes.
func (s Selection) Len() int {
	return len(s.nodes)
}

// Nodes returns all selected nodes.
func (s Selection) Nodes() []Node {
	return s.nodes
}

// At selects the i-th node.
func (s Selection) At(i int) Selection {
	s.w.t.Helper()

	if i < 0 || i >= len(s.nodes) {
		s.w.t.Fatalf("nagotest: %s: index %d out of range, found %d nodes", s.matcher, i, len(s.nodes))
	}

	return Selection{w: s.w, matcher: s.matcher, nodes: s.nodes[i : i+1]}
}

// Node returns the single selected node and fails the test otherwise.
func (s Selection) Node() Node {
	s.w.t.Helper()

	return s.Exactly(1).nodes[0]
}

// Text returns the texts of all text views within the single selected node, joined by a space.
func (s Selection) Text() string {
	s.w.t.Helper()

	var texts []string
	walk(s.Node().Component, nil, func(n Node) bool {
		if tv, ok := n.Component.(*proto.TextView); ok && tv.Value != "" {
			texts = append(texts, string(tv.Value))
		}

		return true
	})

	return strings.Join(texts, " ")
}

// Visible asserts that all selected nodes and their ancestors are visible.
func (s Selection) Visible() Selection {
	s.w.t.Helper()

	for _, n := range s.nodes {
		if c := firstOf(n, "Invisible"); c != nil {
			s.w.t.Fatalf("nagotest: %s: %T is invisible", s.matcher, c)
		}
	}

	return s
}

// Enabled asserts that all selected nodes and their ancestors are not disabled.
func (s Selection) Enabled() Selection {
	s.w.t.Helper()

	for _, n := range s.nodes {
		if c := firstOf(n, "Disabled"); c != nil {
			s.w.t.Fatalf("nagotest: %s: %T is disabled", s.matcher, c)
		}
	}

	return s
}

// firstOf returns the node's component or its nearest ancestor which has the given bool field set.
func firstOf(n Node, name string) proto.Component {
	if boolField(n.Component, name) {
		return n.Component
	}

	for i := len(n.Parents) - 1; i >= 0; i-- {
		if boolField(n.Parents[i], name) {
			return n.Parents[i]
		}
	}

	return nil
}

func find(root proto.Component, parents []proto.Component, m Matcher) []Node {
	var res []Node
	walk(root, parents, func(n Node) bool {
		if m.match(n) {
			res = append(res, n)
		}

		return true
	})

	return res
}

var componentType = reflect.TypeFor[proto.Component]()

// walk visits the root and all nested components in document order, as long as visit returns true.
func walk(root proto.Component, parents []proto.Component, visit func(n Node) bool) {
	if isNil(root) {
		return
	}

	var visitValue func(v reflect.Value, parents []proto.Component) bool
	visitComponent := func(c proto.Component, parents []proto.Component) bool {
		if !visit(Node{Component: c, Parents: parents}) {
			return false
		}

		// full slice expression, so that siblings never share the backing array
		return visitValue(reflect.ValueOf(c), append(parents[:len(parents):len(parents)], c))
	}

	visitValue = func(v reflect.Value, parents []proto.Component) bool {
		switch v.Kind() {
		case reflect.Interface, reflect.Pointer:
			if v.IsNil() {
				return true
			}

			return visitValue(v.Elem(), parents)
		case reflect.Struct:
			for i := range v.NumField() {
				f := v.Field(i)
				if !f.CanInterface() {
					continue
				}

				if f.Kind() == reflect.Interface || f.Kind() == reflect.Pointer {
					if f.IsNil() {
						continue
					}

					if c, ok := f.Interface().(proto.Component); ok && f.Type().Implements(componentType) {
						if !visitComponent(c, parents) {
							return false
						}

						continue
					}
				}

				if !visitValue(f, parents) {
					return false
				}
			}
		case reflect.Slice, reflect.Array:
			for i := range v.Len() {
				e := v.Index(i)
				if e.Kind() == reflect.Interface && !e.IsNil() {
					if c, ok := e.Interface().(proto.Component); ok {
						if !visitComponent(c, parents) {
							return false
						}

						continue
					}
				}

				if !visitValue(e, parents) {
					return false
				}
			}
		default:
			// scalars, strings and maps never contain components
		}

		return true
	}

	visitComponent(root, parents)
}

func isNil(c proto.Component) bool {
	if c == nil {
		return true
	}

	v := reflect.ValueOf(c)
	return v.Kind() == reflect.Pointer && v.IsNil()
}

func field(c any, name string) (reflect.Value, bool) {
	if c == nil {
		return reflect.Value{}, false
	}

	v := reflect.ValueOf(c)
	if v.Kind() == reflect.Pointer && v.IsNil() {
		return reflect.Value{}, false
	}

	v = reflect.Indirect(v)
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	f := v.FieldByName(name)
	return f, f.IsValid()
}

func strField(c any, name string) string {
	f, ok := field(c, name)
	if !ok || f.Kind() != reflect.String {
		return ""
	}

	return f.String()
}

func boolField(c proto.Component, name string) bool {
	f, ok := field(c, name)
	return ok && f.Kind() == reflect.Bool && f.Bool()
}

func ptrField(c proto.Component, name string) proto.Ptr {
	f, ok := field(c, name)
	if !ok || f.Type() != reflect.TypeFor[proto.Ptr]() {
		return 0
	}

	return proto.Ptr(f.Uint())
}
