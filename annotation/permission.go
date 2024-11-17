package annotation

import (
	"fmt"
	"go.wdy.de/nago/pkg/xmaps"
	"regexp"
	"slices"
	"strings"
)

var regexPermissionID = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z0-9_]+)*[a-z0-9_]*$`)

type PermissionID string

func (id PermissionID) Valid() bool {
	return regexPermissionID.MatchString(string(id))
}

var permissions = xmaps.NewConcurrentMap[PermissionID, *UsecaseBuilder]()

// SubjectPermission is the basic contract for the permissions repository, which is used by higher level implementations.
type SubjectPermission interface {
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
func Permissions() []SubjectPermission {
	var res []SubjectPermission
	for _, p := range permissions.All() {
		res = append(res, p.permission)
	}

	slices.SortFunc(res, func(a, b SubjectPermission) int {
		return strings.Compare(a.Identity(), b.Identity())
	})

	return res
}

// proposal: PermissionBuilder create a permission builder to configure multiple localizations.
func (b *UsecaseBuilder) PermissionBuilder(id, name string, docs ...DocElem) any {
	panic("signature proposal only")
}

// Permission is a terminal action for the builder and returns a SubjectPermission. Typically, most use cases have an
// associated permission, which allows exact authorization management based on the permission identifiers.
// If the given permission id has already been declared, a panic is thrown. Use the official companies' language.
// The arguments must have at least single string, representing the stable permission id.
// A permission id looks like
//
//	<tld>.<company>.<product>.(<resource>.<permission>)|<usecase>
//
// Valid examples:
//   - de.worldiety.hako.address.delete
//   - de.worldiety.hako.bookevent
//
// Optionally, the second string represents an alternative Name, which is otherwise derived from the use case.
// Also optionally, the third string represents any explicit description, which may be otherwise derived from additional
// embedded source comments, if available.
func (b *UsecaseBuilder) Permission(args ...string) SubjectPermission {
	if len(args) == 0 {
		panic(fmt.Errorf("must at least one argument, first argument is the id"))
	}

	id := PermissionID(args[0])
	if !id.Valid() {
		panic(fmt.Errorf("invalid permission id, must be of the form <tld>.<company>.<product>.(<resource>.<permission>)|<usecase> e.g. de.worldiety.hako.address.delete: %s", id))
	}

	existing, loaded := permissions.LoadOrStore(id, b)

	if loaded {
		// this is a programming error
		panic(fmt.Errorf("default permission '%s' already defined for use case %v", id, existing.typ))
	}

	var name string
	if len(args) > 1 {
		name = args[1]
	}

	if name == "" {
		name = b.typ.String()
	}

	var desc []string
	if len(args) > 2 {
		desc = args[2:]
	}

	b.permission = perm{
		id:   string(id),
		name: name,
		desc: strings.Join(desc, ""),
	}

	return b.permission
}

// Permission is a shortcut to [UsecaseBuilder.Permission]
func Permission[UC any](args ...string) SubjectPermission {
	b := Usecase[UC]("todo")

	return b.Permission(args...)
}
