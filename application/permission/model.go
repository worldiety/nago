package permission

import (
	"fmt"
	"iter"
	"reflect"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"sync"
)

var regexPermissionID = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z0-9_]+)*[a-z0-9_]*$`)

// ID is unique in the entire permission world. Each use case may have none or exactly one associated permission.
// Note, that a use case may be a composite of other use cases and each will keep its permission check.
// Thus, if a subject wants to execute a composite use case, each permission must be assigned to the subject.
// A Permission is usually a compile-time thing, and it makes not much sense to make it dynamic.
// However, to improve code modularity and decrease coupling, this package provides a global list of permissions
// where other packages or functions can register their individual permissions. This has also the advantage, that
// after compiling all used packages together, the runtime has the exact set of all actually reachable
// permissions in a specific application.
type ID string

func (id ID) Valid() bool {
	return regexPermissionID.MatchString(string(id))
}

// Auditable is used to bootstrap without package cycles.
type Auditable interface {
	// Audit checks if this identity, subject or context has the actual use case permission and may save the positive or
	// negative result in the audit log. An error indicates, that the Subject has not the given permission. The error
	// may just be promoted into error receiving ui components like [alert.BannerError] or [alert.ShowBannerError].
	Audit(permission ID) error

	// HasPermission checks, if the Subject or context has the given permission. A regular use case
	// should use the [Auditable.Audit]. However, this may be used e.g. by the UI to show or hide specific aspects.
	HasPermission(permission ID) bool

	Permissions() iter.Seq[ID]
}

type Permission struct {
	ID ID `json:"id"`
	// Name is the unlocalized fallback or default human-readable name of the permission.
	Name string `json:"name"`
	// Description is the unlocalized fallback or default human-readable description of the permission.
	Description string `json:"desc"`
}

func (p Permission) String() string {
	return p.Name
}

func (p Permission) WithIdentity(id ID) Permission {
	p.ID = id
	return p
}

func (p Permission) Identity() ID {
	return p.ID
}

type permissionContext struct {
	Permission      Permission
	debugDeclaredAt string
}

var globalPermissions = map[ID]permissionContext{}
var mutex sync.RWMutex

// Make is like [Register] but just requires a unique ID as minimum and is a convenience wrapper.
// This may be fine, if the name and description may be obtained automatically e.g. an AST parser and code generator.
// See also [Declare] and [Register].
func Make[UseCase any](id ID) ID {
	return register[UseCase](Permission{ID: id}, 3)
}

// Declare is like [Make] but with 3 parameters. See also [Make] and [Register].
func Declare[UseCase any](id ID, name string, description string) ID {
	return register[UseCase](Permission{ID: id, Name: name, Description: description}, 3)
}

// SetName sets the default permission name or if undefined, ignores it. Ignoring is fine, e.g. because
// an AST preprocessor may include documentation for otherwise unreachable code.
func SetName(id ID, name string) {
	mutex.Lock()
	defer mutex.Unlock()

	if p, ok := globalPermissions[id]; ok {
		p.Permission.Name = name
		globalPermissions[id] = p
	}
}

// SetDescription sets the default permission description or if no such id is defined, ignores it.
// Ignoring is fine, e.g. because
// an AST preprocessor may include documentation for otherwise unreachable code and hence such permission
// has never been registered.
func SetDescription(id ID, desc string) {
	mutex.Lock()
	defer mutex.Unlock()

	if p, ok := globalPermissions[id]; ok {
		p.Permission.Description = desc
		globalPermissions[id] = p
	}
}

// Register connects the given permission and the associated UseCase type parameter. It also applies a simple
// sanity check, if another Permission and UseCase have already been declared together. If so, Register panics.
// See also [Make] or [Declare] as a shortcuts. This allows to introduce more fields to Permission in the future.
func Register[UseCase any](permission Permission) ID {
	return register[UseCase](permission, 3)
}

func register[UseCase any](permission Permission, skip int) ID {
	t := reflect.TypeFor[UseCase]()
	if t.Kind() != reflect.Func {
		panic(fmt.Sprintf("a UseCase type must be a named func type but found: %s", t.Kind()))
	}

	if !permission.ID.Valid() {
		panic(fmt.Sprintf("permission id %s is not valid", permission.ID))
	}

	mutex.Lock()
	defer mutex.Unlock()

	if existing, ok := globalPermissions[permission.ID]; ok {
		panic(fmt.Errorf("permission '%s' was already registered at: %s\n", permission.ID, existing.debugDeclaredAt))
	}

	var src strings.Builder
	var pcs [1]uintptr
	n := runtime.Callers(skip, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		src.WriteString(fmt.Sprintf("%s:%d\n", frame.File, frame.Line))
		if !more {
			break
		}
	}

	globalPermissions[permission.ID] = permissionContext{
		Permission:      permission,
		debugDeclaredAt: src.String(),
	}

	return permission.ID
}

// All returns a snapshot of the current global set of known permissions and returns them ordered asc by name.
func All() iter.Seq[Permission] {
	mutex.RLock()
	tmp := make([]Permission, 0, len(globalPermissions))
	for _, context := range globalPermissions {
		tmp = append(tmp, context.Permission)
	}
	mutex.RUnlock()

	slices.SortFunc(tmp, func(a, b Permission) int {
		// always sort by natural ID order, they are more stable than translated and non-prefixed natural texts
		return strings.Compare(string(a.ID), string(b.ID))
	})

	return slices.Values(tmp)
}
