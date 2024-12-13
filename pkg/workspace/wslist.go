package workspace

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
)

var permList = annotation.Permission[List]("de.worldiety.nago.workspace.list")

// List returns all Workspaces or those narrowed by types.
type List func(subject auth.Subject, types ...Type) iter.Seq2[Workspace, error]

func NewList(repo Repository) List {
	return func(subject auth.Subject, types ...Type) iter.Seq2[Workspace, error] {
		if err := subject.Audit(permList.Identity()); err != nil {
			return xiter.WithError[Workspace](err)
		}

		return xiter.Filter2(func(k Workspace, v error) bool {
			if v != nil {
				return true
			}
			if len(types) == 0 {
				return true
			}

			for _, t := range types {
				if k.Type == t {
					return true
				}
			}

			return false
		}, repo.All())
	}
}
