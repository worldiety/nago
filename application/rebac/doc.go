// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package rebac provides a relational-based access control implementation.
// Each authorization rule is stored as a triple which consists of a subject, relation, and object.
// Even though this generic design allows an arbitrary traversal of complex hierarchies and
// relationships, we currently do not encourage such design.
// However, all RBAC (role-based access control) or direct permission assignments are modeled on top of
// this design to unify the access control model and be open for future extensions.
package rebac
