package workspace

import (
	"go.wdy.de/nago/annotation"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

var permSave = annotation.Permission[Save]("de.worldiety.nago.workspace.save")

type Save func(subject auth.Subject, ws Workspace) error

func NewSave(repo Repository) Save {
	return func(subject auth.Subject, ws Workspace) error {
		if err := subject.Audit(permSave.Identity()); err != nil {
			return err
		}

		if ws.ID == "" {
			ws.ID = data.RandIdent[ID]()
		}

		return repo.Save(ws)
	}
}
