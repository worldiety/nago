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
	"strings"

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

type InfoID string

func NewInfoID(ns Namespace, inst Instance) InfoID {
	srcNs := string(ns)
	srcInst := string(inst)

	// fast path: no escaping required
	if strings.IndexByte(srcNs, ':') == -1 &&
		strings.IndexByte(srcInst, ':') == -1 {
		var sb strings.Builder
		sb.Grow(len(srcNs) + len(srcInst) + 4)
		sb.WriteString(srcNs)
		sb.WriteByte(':')
		sb.WriteString(srcInst)
		return InfoID(sb.String())
	}

	// slow path: escape : as ::
	escape := func(s string) string {
		return strings.ReplaceAll(s, ":", "::")
	}

	var sb strings.Builder
	sb.Grow(len(srcNs) + len(srcInst) + 4) // worst case: all colons doubled
	sb.WriteString(escape(srcNs))
	sb.WriteByte(':')
	sb.WriteString(escape(srcInst))
	return InfoID(sb.String())
}

func (id InfoID) Parse() (Namespace, Instance, error) {
	s := string(id)
	var ns strings.Builder
	var inst strings.Builder
	found := false

	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			if i+1 < len(s) && s[i+1] == ':' {
				// escaped colon
				if found {
					inst.WriteByte(':')
				} else {
					ns.WriteByte(':')
				}
				i++ // skip next ':'
			} else {
				// separator
				if found {
					return "", "", fmt.Errorf("unexpected second separator in InfoID %q", s)
				}
				found = true
			}
		} else {
			if found {
				inst.WriteByte(s[i])
			} else {
				ns.WriteByte(s[i])
			}
		}
	}

	if !found {
		return "", "", fmt.Errorf("missing separator in InfoID %q", s)
	}

	return Namespace(ns.String()), Instance(inst.String()), nil
}

// InstanceInfo represents a human-readable resource instance that can be managed and accessed within the platform.
// This is a generalized summary of a domain-specific instance.
// The UI allows the translation of @123 string reference notation of the i18n package.
type InstanceInfo struct {
	Namespace   Namespace
	ID          Instance
	Name        string
	Description string
}

func (i InstanceInfo) Identity() InfoID {
	return NewInfoID(i.Namespace, i.ID)
}

// Resources allow the inspection and traversal of a namespace.
// This must always follow the semantics of the actual domain and according use-cases and should not
// be another abstraction on top of a repository (besides CRUD). For example, event-sourcing-based subdomains
// have a huge number of events in various repositories, but the actual aggregates and permissions
// cardinalities are totally different. See also [StaticRelationRule].
type Resources interface {
	// Identity returns the namespace which is represented by this provider.
	Identity() Namespace

	// Info returns information about the namespace.
	Info(bundler i18n.Bundler) NamespaceInfo

	// All returns all instances within the namespace.
	All(ctx context.Context) iter.Seq2[InfoID, error]

	// FindByID returns information about the instance with the given id.
	FindByID(ctx context.Context, id InfoID) (option.Opt[InstanceInfo], error)
}

// A StaticRelationRule defines an allowed relation pattern between two namespaces. For example, it is allowed that
// groups can have members but no other relations are allowed. Then a new use case arrives and allows that
// a group has a new relation which allows to delete entities from the pet resources. The new use case
// can register a static relation rule to allow this (e.g. nago.iam.group:delete-pet:my.pets). This allows to
// insert the according Triple (e.g. nago.iam.group:123:delete-pet:my.pets:*).
type StaticRelationRule struct {
	Source   Namespace
	Relation Relation
	Target   Namespace
}

// bStaticRelationRule is the binary representation of a StaticRelationRule which is optimized for cache
// locality and comparision performance. Micro benchmarks have shown that this trick has a huge performance gain.
type bStaticRelationRule struct {
	Source   bNamespace
	Relation bRelation
	Target   bNamespace
}
