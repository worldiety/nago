// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"testing"

	"github.com/expr-lang/expr"
	"github.com/worldiety/jsonptr"
	"github.com/worldiety/option"
)

func TestExpression(t *testing.T) {
	env := map[string]any{
		"obj": jsonptr.NewObj(map[string]jsonptr.Value{
			"foo": jsonptr.String("bar"),
		}),
		"NULL": jsonptr.Null{},
	}

	prg := option.Must(expr.Compile(`obj.GetOr("foo2",NULL).String() == "bar"`, expr.Env(env)))
	out := option.Must(expr.Run(prg, env))
	t.Log(out)
}

// TestFormCardCloneKeepsWrapperAndDeepCopiesChildren guards two clone bugs:
//   - FormCard must clone to a *FormCard (not degrade to *baseViewGroup) and keep
//     its label/supportingText.
//   - The view tree must be deep-cloned so child nodes are not shared with the
//     source (mutating a clone's child must not affect the original).
func TestFormCardCloneKeepsWrapperAndDeepCopiesChildren(t *testing.T) {
	card := NewFormCard("card-1")
	card.SetLabel("Title")
	card.SetSupportingText("Subtitle")
	card.Insert(NewFormText("txt-1", "hello", ""), "")

	clone := card.Clone()

	cc, ok := clone.(*FormCard)
	if !ok {
		t.Fatalf("clone lost concrete type: got %T, want *FormCard", clone)
	}
	if cc.Label() != "Title" || cc.SupportingText() != "Subtitle" {
		t.Fatalf("clone dropped wrapper fields: label=%q supporting=%q", cc.Label(), cc.SupportingText())
	}

	// Locate the cloned child and mutate it; the original must be untouched.
	origChild, _ := FindElementByID(card, "txt-1")
	cloneChild, _ := FindElementByID(cc, "txt-1")
	if origChild == cloneChild {
		t.Fatal("child view node is shared between original and clone (shallow clone)")
	}
	cloneChild.SetVisibleExpr("changed")
	if origChild.VisibleExpr() == "changed" {
		t.Fatal("mutating the clone's child affected the original (shallow clone)")
	}
}

