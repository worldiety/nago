// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"sync"

	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
)

// LinkOwner is the window, the session and the user, for which a link was created, like a download of
// [Window.ExportFiles] or a stream of [Window.AsURI]. Such a link is a plain URL, which a browser may show or which
// may be passed on, thus only a request of the same session may follow it, only as long as the user, who was
// logged in when the link was created, is still logged in to that session, and only while the window lives. The
// session alone is not enough, because its id is kept across a logout and the next login in the same browser. A link
// created without a login is bound to the session alone. See [Application.MayFollow].
type LinkOwner struct {
	scope   *Scope
	session session.ID
	user    user.ID
}

// LinkOwner returns the owner of a link, which the window of this scope creates now: the scope, its session and the
// user of its subject, if the subject is valid. It may be called from any goroutine.
func (s *Scope) LinkOwner() LinkOwner {
	owner := LinkOwner{scope: s}
	if vs := s.virtualSession.Load(); vs != nil && *vs != nil {
		owner.session = (*vs).ID()
	}

	if subject := s.subject.Value(); subject != nil && subject.Valid() {
		owner.user = subject.ID()
	}

	return owner
}

// MayFollow reports whether a request of the given session may follow a link of the owner, see [LinkOwner]. A
// request without a session never may, and no request may after the window has been destroyed, even though the
// observers of its destruction, which forget its links, run later on its event loop.
func (a *Application) MayFollow(owner LinkOwner, requester session.ID) bool {
	if owner.session == "" || requester != owner.session || owner.scope.destroyed.Load() {
		return false
	}

	if owner.user == "" {
		return true
	}

	current := a.findVirtualSession(requester).User()
	return current.IsSome() && current.Unwrap() == owner.user
}

// downloads maps the tokens of the download links to the scopes, which hold the downloads, see [Scope.putExportFiles].
// A token is random and names neither the scope, whose id grants access to a window, nor the export.
type downloads struct {
	mutex   sync.Mutex
	byToken map[string]downloadRef
}

type downloadRef struct {
	scope *Scope
	id    string
}

func (d *downloads) put(token string, scope *Scope, id string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.byToken == nil {
		d.byToken = map[string]downloadRef{}
	}

	d.byToken[token] = downloadRef{scope: scope, id: id}
}

func (d *downloads) remove(token string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	delete(d.byToken, token)
}

func (d *downloads) get(token string) (downloadRef, bool) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	ref, ok := d.byToken[token]
	return ref, ok
}

// Download returns the download of the link with the given token, if the requesting session may follow it, see
// [Application.MayFollow]. Otherwise, it reports false, like for an unknown token. A download lives as long as its
// window shows the page, which created it, and a request does not extend the life of the window. It may be called
// from any goroutine, e.g. by an HTTP handler.
func (a *Application) Download(token string, requester session.ID) (ExportFilesOptions, bool) {
	ref, ok := a.downloads.get(token)
	if !ok {
		return ExportFilesOptions{}, false
	}

	return ref.scope.exportFilesOptions(ref.id, token, requester)
}
