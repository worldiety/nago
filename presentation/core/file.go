package core

import (
	"io"
	"io/fs"
)

// FilesReceiver must be implemented by components which requested a file selection.
// The receiver is called from the event loop, thus if you need to block for a long time, you must run that
// within a different executor.
// Small files and fast processing times are usually never a problem, because we don't need to invalidate within
// millisecond range as mobile apps itself.
// Note, that you must close the files carefully and release the FS manually, when you are done,
// because the scope don't know if you have spawned a concurrent go routine or want to continue processing later.
// Use [Release] for that, as you can't assert which implementation you will actually get.
//
// Intentionally there is no much sense on error return, because this callback is issued over the event looper and thus
// the actual caller cannot be notified anymore. So, if errors occur, the callee must handle it itself.
type FilesReceiver interface {
	OnFilesReceived(fsys fs.FS) error
}

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
