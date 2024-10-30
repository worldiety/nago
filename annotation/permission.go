package annotation

import (
	"fmt"
	"go.wdy.de/nago/pkg/xmaps"
	"go.wdy.de/nago/pkg/xstrings"
	"slices"
	"strings"
)

var permissions = xmaps.NewConcurrentMap[string, *UsecaseBuilder]()

// Permission is the basic contract for the permissions repository, which is used by higher level implementations.
type Permission interface {
	Identity() string
	Name() string
	Desc() string
}

type perm struct {
	id   string
	name string
	desc string
}

func (b perm) Identity() string {
	return b.id
}

func (b perm) Name() string {
	return b.name
}

func (b perm) Desc() string {
	return b.desc
}

// Permissions returns all declared permissions in stable order.
// See [Usecase] or [UsecaseBuilder.Permission] for details.
func Permissions() []Permission {
	var res []Permission
	for _, p := range permissions.All() {
		res = append(res, p.permission)
	}

	slices.SortFunc(res, func(a, b Permission) int {
		return strings.Compare(a.Identity(), b.Identity())
	})

	return res
}

// proposal: PermissionBuilder create a permission builder to configure multiple localizations.
func (b *UsecaseBuilder) PermissionBuilder(id, name string, docs ...DocElem) any {
	panic("signature proposal only")
}

// Permission is a terminal action for the builder and returns a Permission. Typically, most use cases have an
// associated permission, which allows exact authorization management based on the permission identifiers.
// If the given permission id has already been declared, a panic is thrown. Use the official companies language.
func (b *UsecaseBuilder) Permission(id, name string, docs ...DocElem) Permission {
	existing, loaded := permissions.LoadOrStore(id, b)

	if loaded {
		// this is a programming error
		panic(fmt.Errorf("default permission '%s' already defined for use case %v", id, existing.typ))
	}

	b.permission = perm{
		id:   id,
		name: name,
		desc: string(xstrings.Join(docs, "")),
	}

	return b.permission
}
