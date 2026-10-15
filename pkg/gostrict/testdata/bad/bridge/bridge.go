package bridge

import "go.wdy.de/nago/pkg/gostrict/testdata/bad/bridge/hdr"

func Run() int32 {
	return hdr.Add(1, 2)
}
