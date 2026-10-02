---
title: Resource Based Access (ReBAC)
linkTitle: ReBAC
---

Nago stores who may do what as relations between entities, so-called triples: *source*, *relation*,
*target*. For example *user 42* is *member* of *role librarian*, or *group finance* has the permission
*nago.drive.put* on *drive file 7*. Global permissions are relations to the target `global` with the instance
`*`. Roles, groups, direct permissions, token rights and drive file grants are all stored in this database.
Members inherit the relations of their roles and groups.

The ReBAC database is always available. The optional ReBAC module adds an admin editor, in which
administrators grant users, roles and groups permissions on single resources of every registered type.

## Use the database

```go
rdb := std.Must(cfg.RDB()) // *rebac.DB

err := rdb.Put(rebac.Triple{
	Source:   rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance(uid)},
	Relation: rebac.Relation(PermEditBook),
	Target:   rebac.Entity{Namespace: "my.app.book", Instance: rebac.Instance(bookID)},
})
```

`Contains`, `Query`, `Delete` and `DeleteByQuery` read and remove triples.

To show your own entities in the editor, implement `rebac.Resources` (or use `rebac.NewRepositoryResources`
for a repository) and register it with `rdb.RegisterResources`. `rdb.RegisterStaticRule` declares which
combinations of source, relation and target are allowed.

## Enable the editor

```go
import cfgrebac "go.wdy.de/nago/application/rebac/cfg"

mod := std.Must(cfgrebac.Enable(cfg)) // cfgrebac.Module
```

`cfgrebac.Module` has the fields `DB *rebac.DB`, `UseCases ucrebac.UseCases` and `Pages uirebac.Pages` with
the path `Editor`.

## Use cases

| Use case           | Description                                                        |
|--------------------|--------------------------------------------------------------------|
| `FindAllResources` | Lists all registered resource types.                               |
| `WithReBAC`        | Gives audited access to the database, used by the editor.          |

## Permissions

| Permission                     | Allows to                                 |
|--------------------------------|-------------------------------------------|
| `nago.rebac.resources.find_all`| list resource types                       |
| `nago.rebac.resources.with_db` | read and change grants in the editor      |

## UI

The admin center group *Resources and Grants* has a card per resource type, for example users, roles, groups,
drive files or AI sessions. Each opens `admin/rebac/editor?resources=<namespace>`.

## Related

- [Tutorial: resource based access](/docs/examples/tutorial-88-resource-based-access/)
- [Permission Management](../permission_management/)
