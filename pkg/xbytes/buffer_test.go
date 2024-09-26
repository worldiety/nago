package xbytes

import (
	"math/rand"
	"testing"
)

func TestBuffer_WriteUvarint(t *testing.T) {
	const max = 64 * 1024 * 8
	var buf Buffer
	for i := uint64(0); i < max; i++ {
		if _, err := buf.WriteUvarint(i); err != nil {
			t.Fatal(err)
		}
	}

	buf.Reset()
	for i := uint64(0); i < max; i++ {
		v, err := buf.ReadUvarint()
		if err != nil {
			t.Fatal(err)
		}

		if v != i {
			t.Fatal(v)
		}
	}

}

func TestBuffer_WriteString(t *testing.T) {
	var buf Buffer
	for _, str := range makeTestSet() {
		_, err := buf.WriteString(string(str))
		if err != nil {
			t.Fatal(err)
		}
	}

	buf.Reset()

	for _, str := range makeTestSet() {
		s, err := buf.ReadString()
		if err != nil {
			t.Fatal(err)
		}

		if s != string(str) {
			t.Fatal(s)
		}
	}
}

func makeTestSet() [][]byte {
	var res [][]byte
	r := rand.New(rand.NewSource(1234))
	for range 10_000 {
		buf := make([]byte, r.Intn(1024*64))
		r.Read(buf)
		res = append(res, buf)
	}

	return res
}
