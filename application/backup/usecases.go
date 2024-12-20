package backup

import (
	"context"
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/presentation/core"
	"io"
	"iter"
	"time"
)

type Persistence interface {
	// FileStores returns all names of those bucket stores which are tagged for the use of streamable very large
	// binary files.
	FileStores() iter.Seq2[string, error]

	// EntityStores returns all names of those buckets, which are tagged for the use of rather small entities,
	// usually json files.
	EntityStores() iter.Seq2[string, error]

	// FileStore creates or opens the named store as a file storage. This likely has a different implementation
	// than an EntityStore.
	FileStore(name string) (blob.Store, error)

	// EntityStore creates or opens the named store as an entity storage. This likely has a different implementation
	// than a FileStore.
	EntityStore(name string) (blob.Store, error)
}

type Backup func(ctx context.Context, subject auth.Subject, dst io.Writer) error
type Restore func(ctx context.Context, subject auth.Subject, src io.Reader) error

type UseCases struct {
	Backup  Backup
	Restore Restore
}

func NewUseCases(p Persistence) UseCases {
	return UseCases{
		Backup:  NewBackup(p),
		Restore: NewRestore(p),
	}
}

func AsBackupFile(ctx context.Context, subject auth.Subject, backup Backup) core.File {
	return backupFile{
		backup:  backup,
		subject: subject,
		ctx:     ctx,
	}
}

type backupFile struct {
	backup  Backup
	subject auth.Subject
	ctx     context.Context
}

func (b backupFile) Open() (io.ReadCloser, error) {
	return nil, fmt.Errorf("unsupported operation")
}

func (b backupFile) Name() string {
	return fmt.Sprintf("backup_%s.zip", time.Now().Format(time.RFC3339))
}

func (b backupFile) MimeType() (string, bool) {
	return "application/zip", true
}

func (b backupFile) Size() (int64, bool) {
	return 0, false
}

func (b backupFile) Transfer(dst io.Writer) (int64, error) {
	err := b.backup(b.ctx, b.subject, dst)
	return 0, err
}
