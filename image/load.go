package image

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
)

// LoadSrcSet just returns the according source set, if any.
type LoadSrcSet func(user auth.Subject, id ID) (std.Option[SrcSet], error)

func NewLoadSrcSet(repository Repository) LoadSrcSet {
	return func(user auth.Subject, id ID) (std.Option[SrcSet], error) {
		// TODO solve permission questions
		optSet, err := repository.FindByID(id)
		if err != nil {
			return std.None[SrcSet](), fmt.Errorf("failed to load SrcSet from repo: %w", err)
		}

		return optSet, nil
	}
}
