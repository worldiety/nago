---
title: Persistence
weight: 6
---

Nago stores everything in local files below the data directory of the application. You need no database server;
a Nago application is backed up and moved by copying a directory.

## The data directory

`cfg.DataDir()` returns the directory. Nago determines it in this order:

1. the directory set with `cfg.SetDataDir`,
2. the first entry of the environment variable `STATE_DIRECTORY`, which systemd sets for services with
   `StateDirectory=`,
3. `~/.nago/<application ID>`, e.g. `~/.nago/com.example.myapp`.

The directory is created with mode `0700`, so only the owner can access it. Its layout:

| Path                 | Content                                                                    |
|----------------------|----------------------------------------------------------------------------|
| `tdb/`               | the transactional database which holds all entity stores                   |
| `files/<bucket>/`    | one directory per file store                                               |
| `ndb/`               | the default ndb database, once you use `cfg.NDB()`                         |
| `.masterkey`         | the generated master key, unless `NAGO_MASTER_KEY` is set                  |
| `adm/once-after-cfg` | administrative commands which are applied once on the next start           |

`cfg.Directory(name)` allocates a further directory below the data directory for your own purposes.

## Blob stores

The basic storage abstraction is the `blob.Store` from `go.wdy.de/nago/pkg/blob`, a key-value store for byte
streams. The configurator provides two kinds:

- `cfg.EntityStore(bucket)` for many small values up to a few kilobytes, e.g. serialized entities. All entity
  stores share the transactional `tdb` database.
- `cfg.FileStore(bucket)` for large blobs from hundreds of kilobytes to gigabytes. Each blob is a file in the
  file system.

Both are read and written with helpers like `blob.Write` and `blob.Read`, see
[tutorial-25-blobstore](/docs/examples/tutorial-25-blobstore/).

## Repositories

For your entities you usually want a typed repository instead of raw bytes. `application.JSONRepository` stores
values as JSON in an entity store and returns a `data.Repository[E, ID]` from `go.wdy.de/nago/pkg/data`:

```go
type PersonID string

type Person struct {
	ID        PersonID `json:"id"`
	Firstname string   `json:"firstname"`
}

func (p Person) Identity() PersonID { return p.ID }

// within Configure
persons, err := application.JSONRepository[Person, PersonID](cfg, "com.example.myapp.person")
```

An entity only has to implement `Identity()`. The repository offers `Save`, `FindByID` (returning an
`option.Opt`), `All`, `FindAllByID`, `DeleteByID`, `Count` and more; iterators are `iter.Seq2[E, error]`. Pass
the repository to the constructors of your [use cases](../use-cases-and-permissions/).

The JSON is the persisted format, so rename fields with care and keep the `json` tags stable.

## Entities with generated UI

If you need plain CRUD for a type, `cfgent.Enable` from `go.wdy.de/nago/application/ent/cfg` creates the
repository, the use cases, the permissions and admin pages for listing and editing it:

```go
option.Must(cfgent.Enable(cfg, "testdomain.person", "Person", cfgent.Options[Person, PersonID]{}))
```

The second argument is the prefix of the generated permission ids. The entity additionally needs
`WithIdentity(id) E`, and its ID must be a string type. Struct tags like `label`, `visible` or `source` control the
generated forms. See [tutorial-21-entities](/docs/examples/tutorial-21-entities/) and the
[auto form](/docs/components/composite/auto_form/).

Projects in the [architecture style](/docs/architecture/persistence/#why-generic-crud-is-not-available) checked by
speclink do not use `cfgent`, because generated use cases cannot be traced to requirements.

## ndb: event streams and time series

`go.wdy.de/nago/pkg/ndb` is Nago's storage engine for append-only data. `cfg.NDB()` opens the shared database at
`DataDir()/ndb`, `cfg.OpenNDB(path)` opens further ones; both are closed automatically on shutdown. A database
holds named engines of a kind:

- `msgstore` – ordered message streams, the backend for event sourcing with
  [`application/evs`](https://github.com/worldiety/nago/tree/main/application/evs) (decide/evolve handlers and
  projections),
- `tsdb` – compressed time series.

See [tutorial-103-ndb](/docs/examples/tutorial-103-ndb/) and
[tutorial-85-eventsourcing-decide-evolve](/docs/examples/tutorial-85-eventsourcing-decide-evolve/).

## Encryption and backups

Sensitive built-in data, e.g. secrets and sessions, is stored encrypted with the master key of the application,
see [Configuration and deployment](../deployment/#master-key). The
[backup system](/docs/systems/backup_management/) exports and restores the blob stores; encrypted stores stay
encrypted and the master key is not part of the backup, so keep it separately.

## Related

- [Persistence patterns](/docs/architecture/persistence/) – when to use a repository, event sourcing or projections
- [tutorial-57-adm-stores](/docs/examples/tutorial-57-adm-stores/) – inspecting stores in the admin center
- [Inspector](/docs/systems/inspector_management/)
