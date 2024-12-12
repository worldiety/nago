package permission

import "iter"

type FindAll func(subject Auditable) iter.Seq2[Permission, error]
