package sdb

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func Test_wal_set(t *testing.T) {
	f, err := os.OpenFile(filepath.Join(t.TempDir(), "test.wal"), os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	type entryt struct {
		k, v []byte
	}

	wal := newWal(f, nil)
	rnd := rand.New(rand.NewSource(1234))
	var testset []entryt
	for i := range 1000 {
		k := make([]byte, rnd.Intn(128))
		v := make([]byte, rnd.Intn(16*1024))
		rnd.Read(k)
		rnd.Read(v)
		testset = append(testset, entryt{k, v})

		entry := logEntry{walEntrySet, uint64(i), k[:], v[:]}
		if _, err := wal.write(entry); err != nil {
			t.Error(err)
		}
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Error(err)
	}

	var reEntryT []entryt
	var entry logEntry
	for {

		err := wal.read(&entry)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			t.Error(err)
			return
		}

		reEntryT = append(reEntryT, entryt{
			k: bytes.Clone(entry.key),
			v: bytes.Clone(entry.val),
		})
	}

	if !reflect.DeepEqual(reEntryT, testset) {
		t.Errorf("got %v, want %v", reEntryT, testset)
	}
}

func Benchmark_wal_set(t *testing.B) {
	f, err := os.OpenFile(filepath.Join(t.TempDir(), "test.wal"), os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	wal := newWal(f, nil)
	var k [32]byte
	var v [512]byte
	rnd := rand.New(rand.NewSource(1234))
	rnd.Read(k[:])
	rnd.Read(v[:])

	start := time.Now()
	writes := 0
	for i := 0; i < t.N; i++ {
		writes++
		entry := logEntry{walEntrySet, uint64(i), k[:], v[:]}
		if _, err := wal.write(entry); err != nil {
			t.Error(err)
		}
	}

	t.Logf("wrote %d entries in %v\n", writes, time.Now().Sub(start))
}
