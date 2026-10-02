---
title: Drive Management
---

Drive Management is a file storage with folders, versions and access control, comparable to a network share.
Your code works with files through use cases; the UI component `uidrive.Drive` and the page
`uidrive.PageDrive` add a file browser with upload, download, preview, rename and move.

## Enable

```go
import cfgdrive "go.wdy.de/nago/application/drive/cfg"

drives := std.Must(cfgdrive.Enable(cfg)) // cfgdrive.Management
```

`cfgdrive.Management` has the fields `UseCases drive.UseCases` and `Pages uidrive.Pages`. The system registers
no page and no admin card; mount the file browser yourself:

```go
cfg.RootViewWithDecoration("files", func(wnd core.Window) core.View {
	return uidrive.PageDrive(wnd, drives.UseCases)
})
```

`PageDrive` opens the drive `nago.drive` of the current user, or the folder given by the query parameter
`fid`.

## Drives and access control

A drive is a named root folder. `OpenDrive` opens or creates it, either private to a user
(`drive.NamespacePrivate`) or global (`drive.NamespaceGlobal`). Every file has an owner, a group and a
Unix-like file mode (`drive.OtherRead`, `drive.OtherWrite`, ...). In addition, `GrantFileAccess` grants single
users or groups access to a file; these grants are stored in [ReBAC](../rebac_management/) and drive files
show up as a resource in its editor.

## Use cases

| Use case           | Description                                                                    |
|--------------------|--------------------------------------------------------------------------------|
| `OpenDrive`        | Opens or creates a drive.                                                      |
| `ReadDrives`       | Lists the drives visible to a user.                                            |
| `FindDrive`        | Finds the drive of a file.                                                     |
| `Stat`             | Reads the metadata of a file.                                                  |
| `MkDir`            | Creates a folder, or returns an existing one.                                  |
| `Put`              | Creates a file or adds a new version.                                          |
| `Get`              | Opens a version of a file, the latest by default.                              |
| `Zip`              | Zips files on the fly.                                                         |
| `WalkDir`          | Walks a folder tree.                                                           |
| `Rename`           | Renames a file.                                                                |
| `Move`             | Moves a file into another folder and keeps its ID, history and grants.         |
| `Delete`           | Deletes a file, optionally recursive.                                          |
| `GrantFileAccess`, `RevokeFileAccess`, `ReadFileGrants` | Manage the access grants of a file.       |

Files are downloaded through `/api/nago/v1/drive/file?fid=<id>`, which checks the read permission.

## Permissions

| Permission            | Allows to                       |
|-----------------------|---------------------------------|
| `nago.drive.open_file`| open files                      |
| `nago.drive.mkdir`    | create folders                  |
| `nago.drive.put`      | create or update files          |
| `nago.drive.rename`   | rename files                    |
| `nago.drive.delete`   | delete files and folders        |

{{< callout type="warning" >}}
These permissions apply to all drives when granted globally, which equals root access. Prefer the owner,
group and file mode, or grant them per file.
{{< /callout >}}

## Related

- [Tutorial: drive](/docs/examples/tutorial-76-drive/)
- [Tutorial: AI](/docs/examples/tutorial-77-ai/) gives the assistant access to a drive.
