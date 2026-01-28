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
