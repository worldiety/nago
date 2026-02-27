// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

type Query struct {
	triple          Triple
	groupByRelation bool
}

func (q Query) hasGroupBy() bool {
	return q.groupByRelation
}

func Select() Query {
	return Query{}
}

func (q Query) Where() QWhere {
	return QWhere{
		q: q,
	}
}

func (q Query) GroupByRelation() Query {
	q.groupByRelation = true
	return q
}

type QRelation struct {
	q Query
}

func (b QRelation) IsAny() Query {
	return b.Has("")
}

func (b QRelation) Has(r Relation) Query {
	b.q.triple.Relation = r
	return b.q
}

type QWhere struct {
	q Query
}

func (b QWhere) Source() QSource {
	return QSource{q: b.q}
}

func (b QWhere) Target() QTarget {
	return QTarget{q: b.q}
}

func (b QWhere) Relation() QRelation {
	return QRelation{q: b.q}
}

type QSource struct {
	q Query
}

func (b QSource) Set(e Entity) Query {
	b.q.triple.Source = e
	return b.q
}

func (b QSource) IsNamespace(ns Namespace) Query {
	q := b.q
	q.triple.Source.Namespace = ns
	return q
}

func (b QSource) Is(namespace Namespace, instance Instance) Query {
	q := b.q
	q.triple.Source.Namespace = namespace
	q.triple.Source.Instance = instance
	return q
}

func (b QSource) IsInstance(in Instance) Query {
	q := b.q
	q.triple.Source.Instance = in
	return q
}

func (b QSource) IsGlobal() Query {
	q := b.q
	q.triple.Source.Namespace = Global
	q.triple.Source.Instance = AllInstances
	return q
}

func (b QSource) IsAny() Query {
	return b.IsInstance("")
}

type QTarget struct {
	q Query
}

func (b QTarget) IsNamespace(ns Namespace) Query {
	q := b.q
	q.triple.Target.Namespace = ns
	return q
}

func (b QTarget) IsInstance(in Instance) Query {
	q := b.q
	q.triple.Target.Instance = in
	return q
}

// IsGlobal is a shortcut for IsNamespace(Global).IsInstance(AllInstances)
func (b QTarget) IsGlobal() Query {
	q := b.q
	q.triple.Target.Namespace = Global
	q.triple.Target.Instance = AllInstances

	return q
}

func (b QTarget) IsAny() Query {
	return b.IsInstance("")
}

func (b QTarget) Is(namespace Namespace, instance Instance) Query {
	q := b.q
	q.triple.Target.Namespace = namespace
	q.triple.Target.Instance = instance
	return q
}

func (b QTarget) Set(e Entity) Query {
	b.q.triple.Target = e
	return b.q
}
