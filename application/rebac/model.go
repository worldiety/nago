// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

import (
	"context"
	"fmt"
	"iter"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
)

const Global Namespace = "global"

// AllInstances is different from the empty string literal which marks a don't-care case. Do not confuse it with
// this star expression which is literally stored and processed and carries the meaning of a wildcard.
const AllInstances Instance = "*"

type Entity struct {
	Namespace Namespace
	Instance  Instance
}

func (e Entity) IsZero() bool {
	return Entity{} == e
}

func (e Entity) String() string {
	return fmt.Sprintf("%s:%s", e.Namespace, e.Instance)
}

// Triple represents a triple of a subject, relation, and object.
// It is the core of the relation-based access control.
type Triple struct {
	Source   Entity
	Relation Relation
	Target   Entity
}

func (t Triple) Identity() string {
	return t.String()
}

func (t Triple) size() int {
	return len(t.Target.Namespace) + len(t.Target.Instance) +
		len(t.Relation) + len(t.Source.Namespace) + len(t.Source.Instance) + 4
}

func (t Triple) String() string {
	return fmt.Sprintf("%s:%s:%s", t.Source, t.Relation, t.Target)
}

// Namespace is the resource type identifier. Normally this
// may just be a repository name (identifier). However, this needs not to be a real repository.
//
// Examples
//   - nago.user
//   - nago.flow.workspace
//   - some/absolute?ar.b&trary=string}
type Namespace string

// A NamespaceInfo describes an available class or category of resources. It also contains a name and description
// so that users can easily identify and understand the purpose of the resource type. This is not created
// automatically for repositories and must be declared explicitly. Usually each module provides at least a
// configure-function which cares about the registration.
//
// The UI allows the translation of @123 string reference notation of the i18n package.
type NamespaceInfo struct {
	ID          Namespace
	Name        string
	Description string
}

// Instance is the actual resource instance identifier and may refer to a store or repository entry. However, it may be
// also some arbitrary aggregate root, which does not directly refer to a specific store entry as it happens
// within an event sourcing system where the aggregate root is represented by hundreds of events stored each
// as single entry within a store.
// This is usually a domain-specific identifier and created by [data.RandIdent].
// Examples
//   - 6c3143a9b704550bef2c00b7b7d9d8ab
//   - 550e8400-e29b-41d4-a716-446655440000
//   - some/absolute?ar.b&trary=string}
type Instance string

// InstanceInfo represents a human-readable resource instance that can be managed and accessed within the platform.
// This is a generalized summary of a domain-specific instance.
// The UI allows the translation of @123 string reference notation of the i18n package.
type InstanceInfo struct {
	ID          Instance
	Name        string
	Description string
}

func (i InstanceInfo) Identity() Instance {
	return i.ID
}

// Resources allow the inspection and traversal of a namespace and which relations make sense.
// This must always follow the semantics of the actual domain and according use-cases and should not
// be another abstraction on top of a repository (besides CRUD). For example, event-sourcing-based subdomains
// have a huge number of events in various repositories, but the actual aggregates and permissions
// cardinalities are totally different.
type Resources interface {
	// Identity returns the namespace which is represented by this provider.
	Identity() Namespace

	// Info returns information about the namespace.
	Info(bundler i18n.Bundler) NamespaceInfo

	// All returns all instances within the namespace.
	All(ctx context.Context) iter.Seq2[Instance, error]

	// FindByID returns information about the instance with the given id.
	FindByID(ctx context.Context, id Instance) (option.Opt[InstanceInfo], error)

	// Relations describe all relations that are applicable for the given instance of the namespace.
	// The returned triple sequence must be in stable order for later calls. See also [AllInstances] for
	// the meaning of wildcard and don't-care expressions.
	Relations(ctx context.Context, id Instance) iter.Seq[Triple]
}
