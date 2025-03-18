package template

import (
	"go.wdy.de/nago/auth"
	"iter"
	"slices"
)

func NewFindAll(repository Repository) FindAll {
	return func(subject auth.Subject, tags []Tag) iter.Seq2[Project, error] {

		return func(yield func(Project, error) bool) {
		nextProject:
			for project, err := range repository.All() {
				if err != nil {
					if !yield(project, err) {
						return
					}
				}

				if err := subject.AuditResource(repository.Name(), string(project.ID), PermFindAll); err != nil {
					continue nextProject
				}

				if len(tags) > 0 {
					for _, tag := range tags {
						if !slices.Contains(project.Tags, tag) {
							continue nextProject
						}
					}
				}

				if !yield(project, nil) {
					return
				}

			}
		}
	}
}
