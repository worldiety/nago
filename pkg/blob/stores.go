// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package blob

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/xslices"
	"iter"
	"reflect"
)

type StoreType int

const (
	UnspecifiedStore StoreType = iota
	EntityStore
	FileStore
)

type ContentType struct {
	Mime string // e.g. application/json
	Type reflect.Type
}

type StoreInfo struct {
	Name         string
	Type         StoreType
	ContentTypes xslices.Slice[ContentType]
}

type OpenStoreOptions struct {
	Type StoreType
}

// Stores defines a common meta-interface for databases or other sources to interact with store instances on a meta-
// level.
type Stores interface {
	// All lists all available store names, which can be potentially opened.
	All() iter.Seq2[string, error]

	// Stat gathers information about the named store.
	Stat(name string) (option.Opt[StoreInfo], error)

	// Open tries to open or create the named store. If it exists but the store options do not match the configuration
	// an error is returned. This is by intention, because mixing file and entity stores will have such a huge
	// penality when used the wrong way.
	Open(name string, opts OpenStoreOptions) (Store, error)

	// Get returns any known store and may open it, if the implementation knows the type. It will never create
	// a store if it does not yet exist.
	Get(name string) (option.Opt[Store], error)

	// SetContentTypes connects certain runtime types with the named store. Usually these types may be used to
	// unmarshal json encoded blobs. However, technically, each entry may still be encoded arbitrarily. Any
	// existing types are replaced. Usually, this kind of information can never be serialized and restored.
	// Thus, if any Store user was created using this kind of knowledge, it should update the content types.
	// Usually, a single repository per store is used and this fact is typically obvious.
	SetContentTypes(name string, types []ContentType) error

	// Delete removes the store if possible. If protected or not supported, an error must be returned. Deleting a non-
	// existing store is not an error.
	Delete(name string) error
}
