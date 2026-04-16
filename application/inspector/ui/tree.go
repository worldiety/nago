// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiinspector

import (
	"strings"

	"go.wdy.de/nago/application/backup"
	"go.wdy.de/nago/application/inspector"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui/treeview"
)

// createTree builds a hierarchy of treeview.Node from the given stores.
// The store Name (e.g. "nago.iam.user") is split on "." to form intermediate
// folder nodes. Leaf nodes carry a *inspector.Store pointer as Data; intermediate
// nodes have Data == nil.
// The stores slice is assumed to already be sorted by Name.
func createTree(stores []inspector.Store) *treeview.Node[*inspector.Store, string] {
	root := &treeview.Node[*inspector.Store, string]{}

	for i := range stores {
		store := &stores[i]
		segments := strings.Split(store.Name, ".")
		current := root

		for j, seg := range segments {
			id := strings.Join(segments[:j+1], ".")

			if j == len(segments)-1 {
				// leaf node – carries the store pointer
				icon := icons.File
				if store.Stereotype == backup.StereotypeDocument {
					icon = icons.Database
				}

				current.Children = append(current.Children, &treeview.Node[*inspector.Store, string]{
					ID:    id,
					Label: seg,
					Icon:  icon,
					Data:  store,
				})
			} else {
				// intermediate folder node
				current = findOrCreateFolderNode(current, id, seg)
			}
		}
	}

	return root
}

// findOrCreateFolderNode returns the existing child with the given id,
// or appends a new expandable folder node and returns it.
func findOrCreateFolderNode(
	parent *treeview.Node[*inspector.Store, string],
	id, label string,
) *treeview.Node[*inspector.Store, string] {
	for _, child := range parent.Children {
		if child.ID == id {
			return child
		}
	}

	node := &treeview.Node[*inspector.Store, string]{
		ID:         id,
		Label:      label,
		Icon:       icons.Folder,
		Expandable: true,
	}
	parent.Children = append(parent.Children, node)
	return node
}
