package blob

import (
	"fmt"
	"go.wdy.de/nago/pkg/blob"
	"io"
	"io/fs"
	"log/slog"
	"time"
)

type fileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
	sys     any
}

func (f fileInfo) Name() string {
	return f.name
}

func (f fileInfo) Size() int64 {
	return f.size
}

func (f fileInfo) Mode() fs.FileMode {
	return f.mode
}

func (f fileInfo) ModTime() time.Time {
	return f.modTime
}

func (f fileInfo) IsDir() bool {
	return f.isDir
}

func (f fileInfo) Sys() any {
	return f.sys
}

type fakeDir string

func (f fakeDir) Name() string       { return string(f) }
func (f fakeDir) Size() int64        { return 0 }
func (f fakeDir) Mode() fs.FileMode  { return fs.ModeDir | 0500 }
func (f fakeDir) ModTime() time.Time { return time.Unix(0, 0) }
func (f fakeDir) IsDir() bool        { return true }
func (f fakeDir) Sys() any           { return nil }

func (f fakeDir) String() string {
	return fs.FormatFileInfo(f)
}

type dirEnt struct {
	f            *blobFS
	absolutePath string
	name         string
}

func (d dirEnt) Name() string {
	return d.name
}

func (d dirEnt) IsDir() bool {
	return false
}

func (d dirEnt) Type() fs.FileMode {
	return 0
}

func (d dirEnt) Info() (fs.FileInfo, error) {
	return d.f.Stat(d.absolutePath)
}

type vFile struct {
	f            *blobFS
	absolutePath string
	name         string
	reader       io.ReadCloser
}

func (v vFile) Stat() (fs.FileInfo, error) {
	return v.f.Stat(v.absolutePath)
}

func (v vFile) Read(bytes []byte) (int, error) {
	return v.reader.Read(bytes)
}

func (v vFile) Close() error {
	return v.reader.Close()
}

type blobFS struct {
	store blob.Store
}

func (f *blobFS) Open(name string) (fs.File, error) {
	var res vFile
	// TODO this is actually totally incompatible from the interface side. its an implementation detail that this works with the fs implementation
	err := f.store.View(func(tx blob.Tx) error {
		optEnt, err := tx.Get(name)
		if err != nil {
			return err
		}
		if !optEnt.Valid {
			return fs.ErrNotExist
		}

		ent := optEnt.Unwrap()
		reader, err := ent.Open()
		if err != nil {
			return err
		}

		res.f = f
		res.name = name
		res.absolutePath = name
		res.reader = reader
		return nil
	})

	return res, err
}

func (f *blobFS) ReadDir(name string) ([]fs.DirEntry, error) {
	// TODO what should we with the key? split it into virtual directories? that would be a lot of work. Or just prefix search?
	if name != "." {
		return nil, fs.ErrNotExist
	}

	var res []fs.DirEntry
	err := f.store.View(func(tx blob.Tx) error {
		tx.Each(func(entry blob.Entry, err error) bool {
			res = append(res, dirEnt{
				f:            f,
				absolutePath: entry.Key,
				name:         entry.Key,
			})

			return true
		})

		return nil
	})

	return res, err
}

func (f *blobFS) Stat(name string) (fs.FileInfo, error) {
	if name == "." {
		return fakeDir("."), nil
	}

	var info fileInfo
	err := f.store.View(func(tx blob.Tx) error {
		optEnt, err := tx.Get(name)
		if err != nil {
			return err
		}

		if !optEnt.Valid {
			return fs.ErrNotExist
		}

		info.name = name

		ent := optEnt.Unwrap()
		r, err := ent.Open()
		if err != nil {
			return err
		}

		defer r.Close()

		switch file := r.(type) {
		case interface{ Len() int }:
			info.size = int64(file.Len())
		case fs.File:
			stat, err := file.Stat()
			if err != nil {
				return err
			}

			info.size = stat.Size()
			info.modTime = stat.ModTime()
			info.mode = stat.Mode()
			info.sys = stat
		default:
			slog.Error("blob store returned a type which cannot tell me the file size", "type", fmt.Sprintf("%T", file))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return info, nil
}
