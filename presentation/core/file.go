package core

import (
	"bytes"
	"io"
	"iter"
)

// File provides a simple File interface.
// It has a pull semantics, where the data is opened and
// kept alive until closed. See also [PullFile].
type File interface {
	Open() (io.ReadCloser, error)
	// Name of the file
	Name() string
	// MimeType returns the known mime type, if available.
	MimeType() (string, bool)
	// Size returns the known file size, if available.
	Size() (int64, bool)
	// Transfer copies the underlying bytes into dst.
	Transfer(dst io.Writer) (int64, error)
}

type MemFile struct {
	Filename     string
	MimeTypeHint string
	Bytes        []byte
}

func (b MemFile) Transfer(dst io.Writer) (int64, error) {
	n, err := dst.Write(b.Bytes)
	return int64(n), err
}

func (b MemFile) Open() (io.ReadCloser, error) {
	return readerReadCloser{bytes.NewReader(b.Bytes)}, nil
}

func (b MemFile) Name() string {
	return b.Filename
}

func (b MemFile) MimeType() (string, bool) {
	return b.MimeTypeHint, b.MimeTypeHint != ""
}

func (b MemFile) Size() (int64, bool) {
	return int64(len(b.Bytes)), true
}

type readerReadCloser struct {
	*bytes.Reader
}

func (r readerReadCloser) Close() error {
	return nil
}

// FilesReceiver must be implemented by components which requested a file selection.
// The receiver is called from the event loop, thus if you need to block for a long time, you must run that
// within a different executor.
// Small files and fast processing times are usually never a problem, because we don't need to invalidate within
// millisecond range as mobile apps itself.
// Note, that you must close the files carefully and release the FS manually, when you are eolDone,
// because the scope don't know if you have spawned a concurrent go routine or want to continue processing later.
// Use [Release] for that, as you can't assert which implementation you will actually get.
//
// Intentionally there is no much sense on error return, because this callback is issued over the event looper and thus
// the actual caller cannot be notified anymore. So, if errors occur, the callee must handle it itself.
type FilesReceiver func(it iter.Seq2[File, error]) error

// Release tries to clear and close the given thing. If no such interfaces are implemented, the call has no side effects
// and no error is returned.
func Release(a any) error {
	if clearable, ok := a.(interface{ Clear() error }); ok {
		if err := clearable.Clear(); err != nil {
			return err
		}
	}

	if closer, ok := a.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return err
		}
	}

	return nil
}

type ReaderWithMimeType interface {
	io.Reader
	MimeType() string
}

type basicMTReader struct {
	io.Reader
	mt string
}

func (b basicMTReader) MimeType() string {
	return b.mt
}

func (b basicMTReader) Close() error {
	if closer, ok := b.Reader.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}

func WithMimeType(mimeType string, r io.Reader) ReaderWithMimeType {
	return basicMTReader{r, mimeType}
}
