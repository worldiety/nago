package tmpfs

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// the expectation is, that for large files or filesets a batch process may be started later and so
// multiple instances may share the same path, causing races on modifications.
var sharedLocks = map[string]*sync.Mutex{}
var sharedLocksLock sync.Mutex

// FS provides a flat view on files. There are no directories and all temporary files are flat and get
// a strict monotonic increasing number as a file name. Inspect the custom [FileInfo.ResourceName] value for
// the potential initial name.
type FS struct {
	mutex      *sync.Mutex
	scratchDir string
	nextFHnd   int
}

// NewFS initializes its state with the given scratchDir to keep sidecar metadata files and the actual file blobs.
// Use the Import* methods to copy additional files into the storage.
// Use Clear to remove the entire persisted state.
func NewFS(scratchDir string) (*FS, error) {
	sharedLocksLock.Lock()
	defer sharedLocksLock.Unlock()

	lock, ok := sharedLocks[scratchDir]
	if !ok {
		lock = &sync.Mutex{}
		sharedLocks[scratchDir] = lock
	}

	if err := os.MkdirAll(scratchDir, 0700); err != nil {
		return nil, err
	}

	files, err := os.ReadDir(scratchDir)
	if err != nil {
		return nil, err
	}

	var lastNum int
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			id := strings.TrimRight(file.Name(), ".json")
			if num, err := strconv.Atoi(id); err != nil && lastNum < num {
				lastNum = num
			}
		}
	}

	return &FS{scratchDir: scratchDir, nextFHnd: lastNum + 1, mutex: lock}, nil
}

// LocalPath returns the internal scratch dir. If you don't [FS.Clear] it, you can create another instance any
// time later, as long as the system has not reclaimed the temporary space.
func (f *FS) LocalPath() string {
	return f.scratchDir
}

func (f *FS) Open(name string) (fs.File, error) {
	stat, err := f.Stat(name)
	if err != nil {
		return nil, err
	}

	info, ok := stat.(FileInfo)
	if !ok {
		return nil, fs.ErrNotExist
	}

	file, err := os.Open(filepath.Join(f.scratchDir, info.FHash))
	if err != nil {
		return nil, err
	}

	return NewFile(info, file), nil
}

func (f *FS) Stat(name string) (fs.FileInfo, error) {
	if name == "." {
		return fakeDir("."), nil
	}

	name = filepath.Clean(name) + ".json"
	buf, err := os.ReadFile(filepath.Join(f.scratchDir, name))
	if err != nil {
		return nil, err
	}

	var info FileInfo
	if err := json.Unmarshal(buf, &info); err != nil {
		return nil, err
	}

	return info, nil
}

func (f *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if name != "." {
		return nil, fmt.Errorf("not a directory")
	}

	files, err := os.ReadDir(f.scratchDir)
	if err != nil {
		slog.Error("error reading directory", slog.String("path", f.scratchDir))
	}

	// we may get a partial result, so continue anyway
	var res []fs.DirEntry
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			res = append(res, NewDirEntry(strings.TrimRight(file.Name(), ".json"), filepath.Join(f.scratchDir, file.Name())))
		}
	}

	return res, err
}

func (f *FS) Import(resourceName string, r io.Reader) (e error) {
	f.mutex.Lock()
	fhnd := f.nextFHnd
	f.nextFHnd++
	f.mutex.Unlock()

	tmpFileAbsPath := filepath.Join(f.scratchDir, fmt.Sprintf("%d.tmp", fhnd))
	file, err := os.OpenFile(tmpFileAbsPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("cannot open tmp file for write: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil && e == nil {
			e = err
		}
	}()

	size, err := io.Copy(file, r)
	if err != nil {
		return fmt.Errorf("cannot copy data: %w", err)
	}

	// hash that file, often of interest at the domain level e.g. for deduplication or content-addressed-storage
	hashStr, err := calculateSha512_256(tmpFileAbsPath)
	if err != nil {
		return fmt.Errorf("cannot calculate hash: %w", err)
	}

	// because we can, we just dedup now
	absBlobPath := filepath.Join(f.scratchDir, hashStr)
	if err := os.Rename(tmpFileAbsPath, absBlobPath); err != nil {
		return fmt.Errorf("cannot rename tmp file to %s: %w", absBlobPath, err)
	}

	// create sidecar file
	meta := FileInfo{
		FResourceName: resourceName,
		FName:         fmt.Sprintf("%d", fhnd),
		FSize:         size,
		FHash:         hashStr,
		CreatedAt:     time.Now(),
		SeqNum:        int64(fhnd),
	}

	sidecarBuf, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("cannot marshal meta: %w", err)
	}

	sidecarFile := filepath.Join(f.scratchDir, fmt.Sprintf("%d.json", fhnd))
	if err := os.WriteFile(sidecarFile, sidecarBuf, 0600); err != nil {
		return fmt.Errorf("cannot write sidecar file: %w", err)
	}

	return nil
}

// Clear removes all files within the given scratchDir.
func (f *FS) Clear() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return os.RemoveAll(f.scratchDir)
}

func calculateSha512_256(path string) (string, error) {
	hasher := sha512.New512_256()
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("cannot open file: %w", err)
	}

	defer f.Close()

	if _, err := io.Copy(hasher, f); err != nil {
		return "", fmt.Errorf("cannot read file: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
