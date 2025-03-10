package form

import "go.wdy.de/nago/pkg/data"

type Aggregate[A any, ID comparable] interface {
	data.Aggregate[ID]
	WithIdentity(ID) A
}
