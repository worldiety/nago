package token

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"time"
)

type ID string
type Token struct {
	ID   ID     `json:"id,omitempty"`
	Name string `json:"name,omitempty"`

	Salt      []byte             `json:"salt,omitempty"`
	Algorithm user.HashAlgorithm `json:"algorithm,omitempty"`
	TokenHash []byte             `json:"tokenHash,omitempty"`
	CreatedAt time.Time          `json:"createdAt,omitempty"`
	// ValidUntil can be set to zero for unlimited lifetime
	ValidUntil time.Time `json:"validUntil,omitempty"`

	// Impersonation has priority thus if valid, other Groups, Roles, Permissions and Resources are ignored.
	Impersonation option.Opt[user.ID] `json:"impersonation"`

	// Other permissions rules

	Groups      []group.ID                        `json:"groups,omitempty"`
	Roles       []role.ID                         `json:"roles,omitempty"`
	Permissions []permission.Permission           `json:"permissions,omitempty"`
	Licenses    []license.ID                      `json:"licenses,omitempty"`
	Resources   map[user.Resource][]permission.ID `json:"resources,omitempty" json:"resources,omitempty"`
}
