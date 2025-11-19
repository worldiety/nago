package ent

import (
	"fmt"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
)

type Permissions struct {
	Create             permission.ID
	FindByID           permission.ID
	FindAll            permission.ID
	FindAllIdentifiers permission.ID
	DeleteByID         permission.ID
	Update             permission.ID
	Prefix             permission.ID
	EntityName         string
}

// DeclarePermissions is a factory to create a bunch of permissions. Use it at package level, so that permission
// identifiers are available already at package initialization time to avoid working with empty
// permission identifiers accidentally.
// The following identifier naming rules apply:
//   - [Create]: <prefix>.create
//   - [FindByID]: <prefix>.find_by_id
//   - [FindAll]: <prefix>.find_all
//   - [FindAllIdentifiers]: <prefix>.find_all_identifiers
//   - [Update]: <prefix>.update
//   - [DeleteByID]: <prefix>.delete_by_id
func DeclarePermissions[T Aggregate[T, ID], ID data.IDType](prefix permission.ID, entityName string) Permissions {
	if !prefix.Valid() {
		panic(fmt.Errorf("invalid prefix: %s", prefix))
	}

	return Permissions{
		Create:             permission.DeclareCreate[Create[T, ID]](prefix+".create", entityName),
		FindByID:           permission.DeclareFindByID[FindByID[T, ID]](prefix+".find_by_id", entityName),
		FindAll:            permission.DeclareFindAll[FindAll[T, ID]](prefix+".find_all", entityName),
		FindAllIdentifiers: permission.DeclareFindAllIdentifiers[FindAllIdentifiers[T, ID]](prefix+".find_all_identifiers", entityName),
		DeleteByID:         permission.DeclareDeleteByID[DeleteByID[T, ID]](prefix+".delete_by_id", entityName),
		Update:             permission.DeclareUpdate[Update[T, ID]](prefix+".update", entityName),
		Prefix:             prefix,
		EntityName:         entityName,
	}
}
