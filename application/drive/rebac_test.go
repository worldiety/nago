// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import (
	"context"
	"io"
	"iter"
	"strings"
	"testing"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob/fs"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/events"
	"golang.org/x/text/language"
)

// newTestRDB creates an in-memory ReBAC database wired like the real application: the drive per-file static
// rules are registered (user/group -> file for every ACL permission) and the group member resolver is added,
// so that a resource permission granted to a group is honored for its members (see Configurator.RDB and
// application/drive/cfg).
func newTestRDB(t *testing.T) *rebac.DB {
	t.Helper()
	rdb, err := rebac.NewDB(mem.NewBlobStore("rebac"))
	if err != nil {
		t.Fatalf("cannot create rebac db: %v", err)
	}

	// resolvers used by the real wiring for user membership resolution
	rdb.AddResolver(rebac.NewSourceMemberResolver(user.Namespace, role.Namespace))
	rdb.AddResolver(rebac.NewSourceMemberResolver(user.Namespace, group.Namespace))

	// membership relations need a rule as well so we can wire test group memberships
	rdb.RegisterStaticRule(rebac.StaticRule{Source: group.Namespace, Relation: rebac.Member, Target: user.Namespace})

	for _, ns := range []rebac.Namespace{user.Namespace, group.Namespace} {
		for _, pid := range ACLPermissions {
			rdb.RegisterStaticRule(rebac.StaticRule{Source: ns, Relation: rebac.Relation(pid), Target: FileNamespace})
		}
	}

	return rdb
}

func newTestUseCases(t *testing.T) (UseCases, Repository, *rebac.DB) {
	t.Helper()
	repo := Repository(json.NewSloppyJSONRepository[File, FID](mem.NewBlobStore(string(FileNamespace))))
	globalRoots := NamedRootRepository(json.NewSloppyJSONRepository[NamedRoot, string](mem.NewBlobStore("global")))
	userRoots := UserRootRepository(json.NewSloppyJSONRepository[UserRoots, user.ID](mem.NewBlobStore("userroots")))
	blobs, err := fs.NewBlobStore(t.TempDir())
	if err != nil {
		t.Fatalf("cannot create fs blob store: %v", err)
	}
	rdb := newTestRDB(t)
	uc := NewUseCases(events.NewEventBus(), repo, globalRoots, userRoots, blobs, rdb)
	return uc, repo, rdb
}

// newRoot creates a fresh global drive root as SU and returns its stat'd file (with repo attached).
func newRoot(t *testing.T, uc UseCases) File {
	t.Helper()
	drv, err := uc.OpenDrive(user.SU(), OpenDriveOptions{
		Namespace: NamespaceGlobal,
		Name:      "test",
		Create:    true,
		Mode:      0700,
	})
	if err != nil {
		t.Fatalf("open drive: %v", err)
	}

	optRoot, err := uc.Stat(user.SU(), drv.Root)
	if err != nil || optRoot.IsNone() {
		t.Fatalf("stat root: %v", err)
	}
	return optRoot.Unwrap()
}

// statFile stats the given file as SU and fails if it is missing.
func statFile(t *testing.T, uc UseCases, fid FID) File {
	t.Helper()
	optFile, err := uc.Stat(user.SU(), fid)
	if err != nil || optFile.IsNone() {
		t.Fatalf("stat %s: %v", fid, err)
	}
	return optFile.Unwrap()
}

// addGroupMember wires a test membership triple (group has-member user) so that the group member resolver can
// resolve group grants for that user.
func addGroupMember(t *testing.T, rdb *rebac.DB, gid group.ID, uid user.ID) {
	t.Helper()
	err := rdb.Put(rebac.Triple{
		Source:   rebac.Entity{Namespace: group.Namespace, Instance: rebac.Instance(gid)},
		Relation: rebac.Member,
		Target:   rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance(uid)},
	})
	if err != nil {
		t.Fatalf("cannot add group member: %v", err)
	}
}

func stringReader(s string) io.Reader { return strings.NewReader(s) }

// fakeSubject is a minimal auth.Subject for tests. The security-relevant methods mirror the real viewImpl
// semantics: HasResourcePermission resolves (not just contains) against the rebac db, so grants inherited via
// group membership are honored - which is exactly the behavior under test.
type fakeSubject struct {
	id  user.ID
	rdb *rebac.DB
}

func (s fakeSubject) ID() user.ID { return s.id }

func (s fakeSubject) HasPermission(p permission.ID) bool {
	ok, _ := s.rdb.Resolve(rebac.Triple{
		Source:   rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance(s.id)},
		Relation: rebac.Relation(p),
		Target:   rebac.Entity{Namespace: rebac.Global, Instance: rebac.AllInstances},
	})
	return ok
}

func (s fakeSubject) Audit(p permission.ID) error {
	if s.HasPermission(p) {
		return nil
	}
	return user.PermissionDeniedErr
}

func (s fakeSubject) HasResourcePermission(name rebac.Namespace, id rebac.Instance, p permission.ID) bool {
	if s.HasPermission(p) {
		return true
	}
	ok, _ := s.rdb.Resolve(rebac.Triple{
		Source:   rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance(s.id)},
		Relation: rebac.Relation(p),
		Target:   rebac.Entity{Namespace: name, Instance: id},
	})
	return ok
}

func (s fakeSubject) AuditResource(name rebac.Namespace, id rebac.Instance, p permission.ID) error {
	if s.HasResourcePermission(name, id, p) {
		return nil
	}
	return user.PermissionDeniedErr
}

func (s fakeSubject) HasGroup(id group.ID) bool {
	ok, _ := s.rdb.Contains(rebac.Triple{
		Source:   rebac.Entity{Namespace: group.Namespace, Instance: rebac.Instance(id)},
		Relation: rebac.Member,
		Target:   rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance(s.id)},
	})
	return ok
}

func (s fakeSubject) Groups() iter.Seq[group.ID] {
	return func(yield func(group.ID) bool) {
		it := s.rdb.Query(rebac.Select().
			Where().Source().IsNamespace(group.Namespace).
			Where().Relation().Has(rebac.Member).
			Where().Target().Is(user.Namespace, rebac.Instance(s.id)))
		for triple, err := range it {
			if err != nil {
				return
			}
			if !yield(group.ID(triple.Source.Instance)) {
				return
			}
		}
	}
}

func (s fakeSubject) Roles() iter.Seq[role.ID] { return func(yield func(role.ID) bool) {} }
func (s fakeSubject) HasRole(id role.ID) bool  { return false }
func (s fakeSubject) Valid() bool              { return s.id != "" }
func (s fakeSubject) Name() string             { return string(s.id) }
func (s fakeSubject) Firstname() string        { return "" }
func (s fakeSubject) Lastname() string         { return "" }
func (s fakeSubject) Email() string            { return "" }
func (s fakeSubject) Avatar() string           { return "" }
func (s fakeSubject) Language() language.Tag   { return language.English }
func (s fakeSubject) Bundle() *i18n.Bundle     { return nil }
func (s fakeSubject) Context() context.Context { return context.Background() }

// TestGrantAndRevokeUser verifies the convenience grant/revoke use cases for a single user and that the grants
// materialize and are enforced by File.CanRead / CanWrite.
func TestGrantAndRevokeUser(t *testing.T) {
	uc, _, rdb := newTestUseCases(t)
	owner := user.SU()
	root := newRoot(t, uc)

	// a non-owner, non-member user must not be able to read initially
	bob := fakeSubject{id: "bob", rdb: rdb}
	if root.CanRead(bob) {
		t.Fatal("bob must not read before grant")
	}

	// grant read to bob
	if err := uc.GrantFileAccess(owner, root.ID, UserGrantee("bob"), PermsReader...); err != nil {
		t.Fatalf("grant: %v", err)
	}

	root = statFile(t, uc, root.ID)
	if !root.CanRead(bob) {
		t.Fatal("bob must read after grant")
	}
	if root.CanWrite(bob) {
		t.Fatal("bob must not write with reader grant")
	}

	// upgrade bob to writer
	if err := uc.GrantFileAccess(owner, root.ID, UserGrantee("bob"), PermsWriter...); err != nil {
		t.Fatalf("grant writer: %v", err)
	}
	if !root.CanWrite(bob) {
		t.Fatal("bob must write after writer grant")
	}

	// list grants
	grants, err := uc.ReadFileGrants(owner, root.ID)
	if err != nil {
		t.Fatalf("read grants: %v", err)
	}
	if len(grants) != 1 || grants[0].Grantee.User != "bob" {
		t.Fatalf("unexpected grants: %+v", grants)
	}

	// revoke write again
	if err := uc.RevokeFileAccess(owner, root.ID, UserGrantee("bob"), PermPut); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if root.CanWrite(bob) {
		t.Fatal("bob must not write after revoke")
	}
	if !root.CanRead(bob) {
		t.Fatal("bob must still read after only write revoke")
	}
}

// TestGroupGrantResolvesForMember is the key test for the group support: a permission granted to a group must
// be honored for a member of that group through the source-member resolver.
func TestGroupGrantResolvesForMember(t *testing.T) {
	uc, _, rdb := newTestUseCases(t)
	owner := user.SU()
	root := newRoot(t, uc)

	// alice is a member of group "team"
	addGroupMember(t, rdb, "team", "alice")
	alice := fakeSubject{id: "alice", rdb: rdb}

	if root.CanRead(alice) {
		t.Fatal("alice must not read before group grant")
	}

	// grant read+write to the group
	if err := uc.GrantFileAccess(owner, root.ID, GroupGrantee("team"), PermsWriter...); err != nil {
		t.Fatalf("grant group: %v", err)
	}

	root = statFile(t, uc, root.ID)
	if !root.CanRead(alice) {
		t.Fatal("alice must read via group membership")
	}
	if !root.CanWrite(alice) {
		t.Fatal("alice must write via group membership")
	}

	// a user who is not a member must not gain access
	carol := fakeSubject{id: "carol", rdb: rdb}
	if root.CanRead(carol) {
		t.Fatal("carol must not read, not a group member")
	}
}

// TestInheritGrantsOnMkDirAndPut verifies that per-file ACL grants are inherited (snapshot-copied) from the
// parent onto newly created children, analogous to owner/group/mode inheritance.
func TestInheritGrantsOnMkDirAndPut(t *testing.T) {
	uc, _, _ := newTestUseCases(t)
	owner := user.SU()
	root := newRoot(t, uc)

	// grant read to the group on the root
	if err := uc.GrantFileAccess(owner, root.ID, GroupGrantee("team"), PermsReader...); err != nil {
		t.Fatalf("grant group: %v", err)
	}

	// create a subdirectory - it must inherit the group grant
	sub, err := uc.MkDir(owner, root.ID, "sub", MkDirOptions{Mode: 0700})
	if err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}

	grants, err := uc.ReadFileGrants(owner, sub.ID)
	if err != nil {
		t.Fatalf("read grants sub: %v", err)
	}
	if len(grants) != 1 || grants[0].Grantee.Group != "team" {
		t.Fatalf("subdir did not inherit group grant: %+v", grants)
	}

	// create a file inside root via Put - it must inherit the group grant too. Note: Put derives the stored
	// filename from OriginalFilename (or the blob key), not from the name argument, so we locate the file as
	// the single non-directory entry that appeared in root.
	if err := uc.Put(owner, root.ID, "file.txt", stringReader("hello"), PutOptions{Mode: 0600, OriginalFilename: "file.txt"}); err != nil {
		t.Fatalf("put: %v", err)
	}

	// find the created file id (the non-directory entry in root)
	rootStat := statFile(t, uc, root.ID)
	var fileID FID
	for fid := range rootStat.Entries.All() {
		f := statFile(t, uc, fid)
		if !f.IsDir() {
			fileID = fid
		}
	}
	if fileID == "" {
		t.Fatal("created file not found in root")
	}

	fileGrants, err := uc.ReadFileGrants(owner, fileID)
	if err != nil {
		t.Fatalf("read grants file: %v", err)
	}
	if len(fileGrants) != 1 || fileGrants[0].Grantee.Group != "team" {
		t.Fatalf("file did not inherit group grant: %+v", fileGrants)
	}
}

// TestDeleteRevokesGrants verifies that all rebac grants targeting a file are removed when the file is deleted.
func TestDeleteRevokesGrants(t *testing.T) {
	uc, _, rdb := newTestUseCases(t)
	owner := user.SU()
	root := newRoot(t, uc)

	sub, err := uc.MkDir(owner, root.ID, "sub", MkDirOptions{Mode: 0700})
	if err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}

	if err := uc.GrantFileAccess(owner, sub.ID, UserGrantee("bob"), PermsWriter...); err != nil {
		t.Fatalf("grant: %v", err)
	}

	// sanity: triple exists
	target := rebac.Entity{Namespace: FileNamespace, Instance: rebac.Instance(sub.ID)}
	src := rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance("bob")}
	if ok, _ := rdb.Contains(rebac.Triple{Source: src, Relation: rebac.Relation(PermPut), Target: target}); !ok {
		t.Fatal("expected grant before delete")
	}

	if err := uc.Delete(owner, sub.ID, DeleteOptions{Recursive: true}); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if ok, _ := rdb.Contains(rebac.Triple{Source: src, Relation: rebac.Relation(PermPut), Target: target}); ok {
		t.Fatal("grant should be revoked after delete")
	}
}

// TestGrantRequiresWritePermission verifies that a subject which may not write a file cannot change its ACL.
func TestGrantRequiresWritePermission(t *testing.T) {
	uc, _, rdb := newTestUseCases(t)
	root := newRoot(t, uc)

	// mallory has no rights on root
	mallory := fakeSubject{id: "mallory", rdb: rdb}
	if err := uc.GrantFileAccess(mallory, root.ID, UserGrantee("mallory"), PermsWriter...); err == nil {
		t.Fatal("expected permission denied when granting without write rights")
	}
}

// TestCanDeleteUsesParentWritePermission verifies the unix delete semantics: write permission on the parent
// directory is sufficient to delete a child, and - conversely - mere write permission on the file itself is
// NOT (that used to be a bug where CanDelete loaded the file instead of its parent).
func TestCanDeleteUsesParentWritePermission(t *testing.T) {
	uc, _, rdb := newTestUseCases(t)
	owner := user.SU()
	root := newRoot(t, uc)

	sub, err := uc.MkDir(owner, root.ID, "sub", MkDirOptions{Mode: 0700})
	if err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}

	if err := uc.Put(owner, sub.ID, "file.txt", stringReader("hello"), PutOptions{Mode: 0600, OriginalFilename: "file.txt"}); err != nil {
		t.Fatalf("put: %v", err)
	}

	// locate the created file
	subStat := statFile(t, uc, sub.ID)
	var fileID FID
	for fid := range subStat.Entries.All() {
		if f := statFile(t, uc, fid); !f.IsDir() {
			fileID = fid
		}
	}
	if fileID == "" {
		t.Fatal("created file not found")
	}

	bob := fakeSubject{id: "bob", rdb: rdb}

	// case 1: bob may only write the file itself (PermPut grant on the file), but has no rights on the
	// parent directory and no PermDelete. He must NOT be able to delete it.
	if err := uc.GrantFileAccess(owner, fileID, UserGrantee("bob"), PermsWriter...); err != nil {
		t.Fatalf("grant writer on file: %v", err)
	}

	file := statFile(t, uc, fileID)
	if !file.CanWrite(bob) {
		t.Fatal("precondition: bob should be able to write the file")
	}
	if file.CanDelete(bob) {
		t.Fatal("bob must NOT be able to delete: write on the file itself is not enough, needs parent write or PermDelete")
	}

	// case 2: now grant bob write on the parent directory. He must be able to delete the child.
	if err := uc.GrantFileAccess(owner, sub.ID, UserGrantee("bob"), PermsWriter...); err != nil {
		t.Fatalf("grant writer on parent: %v", err)
	}

	file = statFile(t, uc, fileID)
	if !file.CanDelete(bob) {
		t.Fatal("bob must be able to delete once he can write the parent directory")
	}
}

