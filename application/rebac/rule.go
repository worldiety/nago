// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

import (
	"encoding/binary"
	"iter"
	"slices"
)

// MaxRuleDepth limits how many memberships [DB.Resolve] and [DB.Relations] follow transitively. It bounds the
// work of chained rules and terminates cyclic memberships.
const MaxRuleDepth = 8

// A Rule derives relations, which are not stored as triples, from the memberships of a source. The rule
//
//	Rule{Source: "nago.iam.user", Via: "nago.iam.role"}
//
// reads: a user has every relation on every target, which a role has, whose member the user is. With the stored
// triples
//
//	nago.iam.role:admin:member:nago.iam.user:torben
//	nago.iam.role:admin:read:nago.doc:*
//
// [DB.Resolve] allows nago.iam.user:torben:read:nago.doc:d1. Rules chain: if Via is the Source of another rule, the
// memberships are followed transitively, at most [MaxRuleDepth] levels deep.
//
// A rule is data, not code. The database knows what it means and derives both decisions from it: the point check
// of [DB.Resolve] and the enumeration of [DB.Relations], which can therefore never diverge.
type Rule struct {
	// Source is the namespace of the sources, for which the rule applies, e.g. nago.iam.user.
	Source Namespace
	// Via is the namespace of the entities, whose relations are inherited by their members, e.g. nago.iam.role.
	Via Namespace
	// Membership is the relation, which connects a Via entity to its members and is stored as
	// Via:x:Membership:Source:y. The zero value means [Member].
	Membership Relation
}

func (r Rule) normalize() Rule {
	if r.Membership == "" {
		r.Membership = Member
	}

	return r
}

// bRule is the interned representation of a Rule.
type bRule struct {
	src        bNamespace
	via        bNamespace
	membership bRelation
}

// AddRule registers the rule for [DB.Resolve] and [DB.Relations]. A rule with an empty Source or Via, or one which
// has already been added, is ignored.
func (db *DB) AddRule(rule Rule) {
	rule = rule.normalize()
	if rule.Source == "" || rule.Via == "" {
		return
	}

	b := bRule{
		src:        bNamespace(db.strTable.Intern(string(rule.Source))),
		via:        bNamespace(db.strTable.Intern(string(rule.Via))),
		membership: bRelation(db.strTable.Intern(string(rule.Membership))),
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	// copy on write, so that the hot path reads the rules without a lock
	rules := slices.Clone(*db.rules.Load())
	if slices.Contains(rules, b) {
		return
	}

	rules = append(rules, b)
	db.rules.Store(&rules)
}

// Rules returns all registered rules in registration order.
func (db *DB) Rules() iter.Seq[Rule] {
	return func(yield func(Rule) bool) {
		for _, rule := range *db.rules.Load() {
			if !yield(Rule{
				Source:     Namespace(db.strTable.String(uint32(rule.src))),
				Via:        Namespace(db.strTable.String(uint32(rule.via))),
				Membership: Relation(db.strTable.String(uint32(rule.membership))),
			}) {
				return
			}
		}
	}
}

// probe is a question to the binary layer: which relations has src on tgt. The target instance may be unknown to
// the string table (and thus 0), because a wildcard triple also covers instances, which the database has never
// seen. wild is the interned [AllInstances] or 0, if no wildcard triple has ever been stored.
type probe struct {
	src  bEntity
	rel  bRelation // 0 asks for all relations
	tgt  bEntity
	wild bInstance
}

// newProbe looks up all strings with a single lock. It returns false, if the question cannot have a positive
// answer, because a string has never been interned or an entity is incomplete.
func (db *DB) newProbe(src Entity, rel Relation, tgt Entity) (probe, bool) {
	if src.Namespace == "" || src.Instance == "" || tgt.Namespace == "" || tgt.Instance == "" {
		return probe{}, false
	}

	t := db.strTable
	t.mutex.RLock()
	srcNs, ok1 := t.lookupLocked(string(src.Namespace))
	srcInst, ok2 := t.lookupLocked(string(src.Instance))
	relID, ok3 := t.lookupLocked(string(rel))
	tgtNs, ok4 := t.lookupLocked(string(tgt.Namespace))
	tgtInst, hasTgtInst := t.lookupLocked(string(tgt.Instance))
	wild, hasWild := t.lookupLocked(string(AllInstances))
	t.mutex.RUnlock()

	if !(ok1 && ok2 && ok3 && ok4) || !(hasTgtInst || hasWild) {
		return probe{}, false
	}

	if !hasTgtInst {
		tgtInst = 0
	}

	if !hasWild {
		wild = 0
	}

	return probe{
		src:  bEntity{Namespace: bNamespace(srcNs), Instance: bInstance(srcInst)},
		rel:  bRelation(relID),
		tgt:  bEntity{Namespace: bNamespace(tgtNs), Instance: bInstance(tgtInst)},
		wild: bInstance(wild),
	}, true
}

// hasWildcard returns true, if the wildcard target of the probe must be considered as well.
func (p probe) hasWildcard() bool {
	return p.wild != 0 && p.tgt.Instance != p.wild
}

// wildcard returns the probe for the wildcard target of the same namespace.
func (p probe) wildcard() probe {
	p.tgt.Instance = p.wild
	return p
}

// via returns the probe for a membership, i.e. the same question asked for the entity the source is a member of.
func (p probe) via(ns bNamespace, inst bInstance) probe {
	p.src = bEntity{Namespace: ns, Instance: inst}
	return p
}

// contains is the exact lookup of the probe.
func (db *DB) contains(p probe) bool {
	_, ok := db.forward.Get(binaryTriple{A: p.src, Relation: p.rel, B: p.tgt}.Binary())
	return ok
}

// check decides the probe: the triple is stored as is, or for the wildcard target, or a rule derives it from a
// membership of the source. It never allocates for the usual case of a few memberships.
func (db *DB) check(p probe, depth int) bool {
	if db.contains(p) {
		return true
	}

	if p.hasWildcard() && db.contains(p.wildcard()) {
		return true
	}

	if depth >= MaxRuleDepth {
		return false
	}

	for _, rule := range *db.rules.Load() {
		if rule.src != p.src.Namespace {
			continue
		}

		// collect first: the btree must not be entered recursively while it is scanned
		var buf [16]bInstance
		for _, inst := range db.appendMemberships(buf[:0], p.src, rule) {
			if db.check(p.via(rule.via, inst), depth+1) {
				return true
			}
		}
	}

	return false
}

// appendMemberships appends the instances of the rule's Via namespace, which the source is a member of. The
// backward index stores target, relation and source, thus they are a single contiguous range.
func (db *DB) appendMemberships(dst []bInstance, src bEntity, rule bRule) []bInstance {
	pivot := binaryTriple{A: src, Relation: rule.membership, B: bEntity{Namespace: rule.via}}.Binary()
	db.backward.Ascend(pivot, func(item fixedBinaryTriple) bool {
		if [16]byte(item[:16]) != [16]byte(pivot[:16]) {
			return false
		}

		dst = append(dst, bInstance(binary.LittleEndian.Uint32(item[16:20])))
		return true
	})

	return dst
}
