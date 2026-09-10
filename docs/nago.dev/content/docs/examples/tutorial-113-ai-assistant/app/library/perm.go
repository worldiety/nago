// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package library

import (
	"go.wdy.de/nago/application/permission"
)

// Permissions of the library, one per use case and bound to it through the type parameter.
//
// One per use case is what makes authorisation assignable and auditable: an operator hands out the ability to
// lend without also handing out the ability to take back. A permission covering several use cases can only be
// granted or withheld wholesale.
//
// The texts come from the framework's translation catalogue via the Declare<Verb> helpers rather than being
// written here, because these strings appear in the role editor, where a non-developer decides who may do
// what - in their own language.
//
// This is also the entire authorisation model of the assistant. The tools in ai/ call these use cases, the
// use cases audit these permissions against the acting subject, and so the assistant can never do more than
// the person operating it.
var (
	PermFindAllBooks = permission.DeclareFindAll[FindAllBooks]("tutorial.library.book.find_all", "Buch")
	PermLendBook     = permission.DeclareUpdate[LendBook]("tutorial.library.book.lend", "Buch")
	PermReturnBook   = permission.DeclareUpdate[ReturnBook]("tutorial.library.book.return", "Buch")
)
