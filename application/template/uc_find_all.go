package template

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
	"slices"
)

func NewFindAll(repository Repository) FindAll {
	return func(subject auth.Subject, tags []Tag) iter.Seq2[Project, error] {
		if err := subject.Audit(PermFindAll); err != nil {
			return xiter.WithError[Project](err)
		}

		return func(yield func(Project, error) bool) {
		nextProject:
			for project, err := range repository.All() {
				if err != nil {
					if !yield(project, err) {
						return
					}
				}

				if len(project.ReadableBy) == 0 {
					if !yield(project, nil) {
						return
					}
				}

				for _, id := range project.ReadableBy {
					if len(tags) > 0 {
						for _, tag := range tags {
							if !slices.Contains(project.Tags, tag) {
								continue nextProject
							}
						}
					}
					if subject.HasGroup(id) {
						if !yield(project, nil) {
							return
						}
					}
				}
			}
		}
	}
}
