package tdb

import (
	"go.wdy.de/nago/pkg/xbytes"
	"math/rand"
	"reflect"
	"testing"
)

func Test_node_write(t *testing.T) {
	testSet := makeTestNodes()
	var buf xbytes.Buffer
	var tmp Node
	for _, n := range testSet {
		buf.Reset()
		n.write(&buf)
		buf.Reset()
		if err := tmp.read(&buf); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(tmp, n) {
			t.Fatal("Node re-read failure")
		}
	}
}

func makeTestNodes() []Node {
	var res []Node
	r := rand.New(rand.NewSource(1234))
	for range 10_000 {
		if r.Intn(1) == 1 {
			res = append(res, Node{
				kind:   setKeyValue,
				tx:     r.Uint64(),
				bucket: nextBuf(r),
				key:    nextBuf(r),
				val:    nextBuf(r),
			})
		} else {
			res = append(res, Node{
				kind:   removeKeyValue,
				tx:     r.Uint64(),
				bucket: nextBuf(r),
				key:    nextBuf(r),
			})
		}
	}

	return res
}

func nextBuf(r *rand.Rand) []byte {
	buf := make([]byte, r.Intn(1024*64))
	r.Read(buf)
	return buf
}
