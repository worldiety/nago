// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package libsync

import (
	"context"
	"crypto/sha3"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/document"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/drive"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/events"
)

func NewSynchronize(bus events.Bus, findProvider ai.FindProviderByID, repo Repository, syncRepo SyncRepository, stores blob.Stores, walkDir drive.WalkDir, openFile drive.Get, statFile drive.Stat) Synchronize {
	return func(subject auth.Subject, lib library.ID) error {
		if err := subject.AuditResource(repo.Name(), string(lib), PermSynchronize); err != nil {
			return err
		}

		optJob, err := repo.FindByID(lib)
		if err != nil {
			return err
		}

		if optJob.IsNone() {
			return fmt.Errorf("sync job for library not found: %s: %w", lib, os.ErrNotExist)
		}

		job := optJob.Unwrap()
		optProv, err := findProvider(user.SU(), job.Provider)
		if err != nil {
			return fmt.Errorf("sync job requires provider which has not been found: %s: %w", job.Provider, err)
		}

		if len(job.Sources) == 0 {
			slog.Warn("sync job does not have any sources, ignoring", "provider", job.Provider, "job", job.ID)
			return nil
		}

		if optProv.IsNone() {
			return fmt.Errorf("sync job referred to a provider which is not found: %s: %w", job.Provider, os.ErrNotExist)
		}

		prov := optProv.Unwrap()

		if prov.Libraries().IsNone() {
			return fmt.Errorf("provider does not support libraries: %s: %w", job.Provider, os.ErrNotExist)
		}

		remoteDocs, err := collectRemoteDocs(prov, job)
		if err != nil {
			return err
		}

		localStoreDocs, err := collectLocalStoreDocs(stores, job)
		if err != nil {
			return err
		}

		localDriveDocs, err := collectLocalDriveDocs(walkDir, openFile, statFile, job)
		if err != nil {
			return err
		}

		syncedDocs, err := collectSyncedDocs(syncRepo, job)
		if err != nil {
			return err
		}

		tasks := buildTaskList(syncedDocs, remoteDocs, localStoreDocs, localDriveDocs)
		if err := applyTasks(stores, openFile, statFile, prov, syncRepo, job, tasks); err != nil {
			return err
		}

		slog.Info("sync job finished", "provider", job.Provider, "job", job.ID)
		return nil
	}
}

func applyTasks(stores blob.Stores, openFile drive.Get, statFile drive.Stat, prov provider.Provider, syncRepo SyncRepository, job Job, tasks []tTask) error {
	lib := prov.Libraries().Unwrap().Library(job.ID)

	slog.Info("sync job task plan calculated", "tasks", len(tasks), "job", job.ID, "provider", prov.Identity())

	// first apply only deletes, otherwise we may get some unwanted stability toggle effects
	for _, task := range tasks {
		if task.DeleteRemote.Valid {
			if err := lib.Delete(user.SU(), task.DeleteRemote.ID); err != nil {
				return fmt.Errorf("cannot delete remote document: %s: %w", task.DeleteRemote.ID, err)
			}

			slog.Info("sync job deleted remote", "document", task.DeleteRemote.ID, "job", job.ID)
		}

		if task.DeleteSynced.Valid {
			if err := syncRepo.DeleteByID(task.DeleteSynced.ID); err != nil {
				return fmt.Errorf("cannot delete SyncInfo: %s: %w", task.DeleteSynced.ID, err)
			}
		}
	}

	// than apply all inserts
	for _, task := range tasks {
		if task.InsertRemote.Valid {
			remote := task.InsertRemote
			optReader, err := open(remote.Src, stores, openFile)
			if err != nil {
				return fmt.Errorf("cannot open load reader: %w", err)
			}

			if optReader.IsNone() {
				slog.Info("sync job stale local source ignoring")
				continue // may got stale during execution, which may be normal
			}

			reader := optReader.Unwrap()

			var fname string
			if remote.Src.Drive.Valid {
				optStat, _ := statFile(user.SU(), remote.Src.Drive.File)
				fname = optStat.UnwrapOr(drive.File{Filename: string(remote.Src.Drive.File)}).Name()
			}

			if remote.Src.Store.Valid {
				fname = remote.Src.Store.Name + "/" + remote.Src.Store.Key
			}

			doc, err := lib.Create(user.SU(), document.CreateOptions{
				Filename: fname,
				Reader:   reader,
			})

			if err := reader.Close(); err != nil {
				return fmt.Errorf("cannot close reader: %w", err)
			}

			if err != nil {
				return fmt.Errorf("cannot create provider document: %w", err)
			}

			if err := syncRepo.Save(SyncInfo{
				Remote: doc.ID,
				Src:    remote.Src,
				Size:   remote.stat.Size,
				Hash:   remote.stat.Hash,
			}); err != nil {
				return fmt.Errorf("cannot save provider document: %w", err)
			}

			slog.Info("sync job inserted remote document", "document", doc.ID, "job", job.ID)
		}

	}

	return nil
}

func open(src SourceDocument, stores blob.Stores, openFile drive.Get) (option.Opt[io.ReadCloser], error) {
	var zero option.Opt[io.ReadCloser]
	if src.Store.Valid {
		optStore, err := stores.Get(src.Store.Name)
		if err != nil {
			return zero, err
		}

		if optStore.IsNone() {
			return zero, nil
		}

		return optStore.Unwrap().NewReader(context.Background(), src.Store.Key)
	}

	if src.Drive.Valid {
		optFile, err := openFile(user.SU(), src.Drive.File, "")
		if err != nil {
			return zero, err
		}

		if optFile.IsNone() {
			return zero, nil
		}

		reader, err := optFile.Unwrap().Open()
		if err != nil {
			return zero, err
		}

		return option.Some(reader), nil
	}

	return zero, nil
}

type tTask struct {
	DeleteRemote struct {
		Valid bool
		ID    document.ID
	}

	InsertRemote struct {
		Valid bool
		Src   SourceDocument
		stat  srcStat
	}

	DeleteSynced struct {
		Valid bool
		ID    document.ID
	}
}

func buildTaskList(syncInfos map[document.ID]SyncInfo, remoteDocs map[document.ID]document.Document, localStoreDocs map[srcStoreKey]srcStat, localDrive map[drive.FID]srcStat) []tTask {
	var tasks []tTask

	// find existing files which are just not available at remote: these are files which are new and don't have any sync info
	for key, stat := range localStoreDocs {
		found := false
		for _, info := range syncInfos {
			if store := info.Src.Store; store.Valid && store.Name == key.Store && store.Key == key.Key {
				found = true
				break
			}
		}

		if !found {
			var task tTask
			task.InsertRemote.Valid = true
			task.InsertRemote.Src.Store.Valid = true
			task.InsertRemote.Src.Store.Name = key.Store
			task.InsertRemote.Src.Store.Key = key.Key
			task.InsertRemote.stat = stat
			tasks = append(tasks, task)
		}
	}

	for fid, stat := range localDrive {
		found := false
		for _, info := range syncInfos {
			if drv := info.Src.Drive; drv.Valid && drv.File == fid {
				found = true
				break
			}
		}

		if !found {
			var task tTask
			task.InsertRemote.Valid = true
			task.InsertRemote.Src.Drive.Valid = true
			task.InsertRemote.Src.Drive.File = fid
			task.InsertRemote.stat = stat
			tasks = append(tasks, task)
		}
	}

	// find existing files which have been changed locally
	for key, stat := range localStoreDocs {
		found := false
		var ifo SyncInfo
		for _, info := range syncInfos {
			if store := info.Src.Store; store.Valid && store.Name == key.Store && store.Key == key.Key {
				found = true
				ifo = info
				break
			}
		}

		if !found {
			continue
		}

		if ifo.Hash == stat.Hash {
			continue
		}

		var task tTask
		task.InsertRemote.Valid = true
		task.InsertRemote.Src.Store.Valid = true
		task.InsertRemote.Src.Store.Name = key.Store
		task.InsertRemote.Src.Store.Key = key.Key
		task.InsertRemote.stat = stat
		tasks = append(tasks, task)

		task = tTask{}
		task.DeleteRemote.Valid = true
		task.DeleteRemote.ID = ifo.Remote
		tasks = append(tasks, task)

		task = tTask{}
		task.DeleteSynced.Valid = true
		task.DeleteSynced.ID = ifo.Remote
		tasks = append(tasks, task)
	}

	for fid, stat := range localDrive {
		found := false
		var ifo SyncInfo
		for _, info := range syncInfos {
			if drv := info.Src.Drive; drv.Valid && drv.File == fid {
				found = true
				ifo = info
				break
			}
		}

		if !found {
			continue
		}

		if ifo.Hash == stat.Hash {
			continue
		}

		var task tTask
		task.InsertRemote.Valid = true
		task.InsertRemote.Src.Drive.Valid = true
		task.InsertRemote.Src.Drive.File = fid
		task.InsertRemote.stat = stat
		tasks = append(tasks, task)

		task = tTask{}
		task.DeleteRemote.Valid = true
		task.DeleteRemote.ID = ifo.Remote
		tasks = append(tasks, task)

		task = tTask{}
		task.DeleteSynced.Valid = true
		task.DeleteSynced.ID = ifo.Remote
		tasks = append(tasks, task)
	}

	// find extra files: these are files which are remote but not in syncInfo. These are manual uploads from outside
	for id := range remoteDocs {
		if _, ok := syncInfos[id]; !ok {
			var task tTask
			task.DeleteRemote.Valid = true
			task.DeleteRemote.ID = id
			tasks = append(tasks, task)
		}
	}

	// find synced files which are not available anymore, these are files which have been synced once but are removed locally
	for id, info := range syncInfos {
		found := false
		if info.Src.Store.Valid {
			for srcKey := range localStoreDocs {
				if srcKey.Store == info.Src.Store.Name && srcKey.Key == info.Src.Store.Key {
					found = true
					break
				}
			}
		}

		if found {
			continue
		}

		if info.Src.Drive.Valid {
			for fid := range localDrive {
				if info.Src.Drive.File == fid {
					found = true
					break
				}
			}
		}

		if !found {
			// remove the remote, if we ever synced it
			var task tTask
			if _, ok := remoteDocs[id]; ok {
				task.DeleteRemote.Valid = true
				task.DeleteRemote.ID = id
			}

			tasks = append(tasks, task)

			// always remove the synced infos
			task = tTask{}
			task.DeleteSynced.Valid = true
			task.DeleteSynced.ID = id
			tasks = append(tasks, task)
		}
	}

	// find files which were synced but now missing remote
	for id, info := range syncInfos {
		if _, ok := remoteDocs[id]; !ok {
			var task tTask
			task.DeleteSynced.Valid = true
			task.DeleteSynced.ID = id // cleanup, perhaps removed AND changed which will result in new doc-id
			tasks = append(tasks, task)

			task = tTask{}
			task.InsertRemote.Valid = true
			task.InsertRemote.Src = info.Src
			tasks = append(tasks, task)
		}
	}

	return tasks
}

func collectSyncedDocs(repo SyncRepository, job Job) (map[document.ID]SyncInfo, error) {
	localDocs := make(map[document.ID]SyncInfo)
	for info, err := range repo.All() {
		if err != nil {
			return localDocs, err
		}

		localDocs[info.Remote] = info
	}

	slog.Info("sync job collected synced info list", "docs", len(localDocs), "provider", job.Provider, "job", job.ID)

	return localDocs, nil
}

func collectRemoteDocs(prov provider.Provider, job Job) (map[document.ID]document.Document, error) {
	remoteDocs := map[document.ID]document.Document{}

	for doc, err := range prov.Libraries().Unwrap().Library(job.ID).All(user.SU()) {
		if err != nil {
			return nil, fmt.Errorf("cannot build remote document list: %w", err)
		}

		remoteDocs[doc.ID] = doc
	}

	slog.Info("sync job collected remote document list", "docs", len(remoteDocs), "provider", job.Provider, "job", job.ID)
	return remoteDocs, nil
}

func collectLocalStoreDocs(stores blob.Stores, job Job) (map[srcStoreKey]srcStat, error) {
	ctx := context.Background()
	localStoreDocs := map[srcStoreKey]srcStat{}
	for _, src := range job.Sources {
		if src.Store.Valid {
			optStore, err := stores.Get(src.Store.Name)
			if err != nil {
				return nil, fmt.Errorf("cannot open store '%s': %w", src.Store.Name, err)
			}

			if optStore.IsNone() {
				return nil, fmt.Errorf("sync job referred to a non-existing store '%s': %w", src.Store.Name, os.ErrNotExist)
			}

			store := optStore.Unwrap()
			for id, err := range store.List(ctx, blob.ListOptions{}) {
				if err != nil {
					return nil, fmt.Errorf("cannot list store '%s': %w", src.Store.Name, err)
				}

				optReader, err := store.NewReader(ctx, id)
				if err != nil {
					return nil, fmt.Errorf("cannot open store reader '%s.%s': %w", src.Store.Name, id, err)
				}

				if optReader.IsNone() {
					continue // item just gone under currency, that is usually fine
				}

				reader := optReader.Unwrap()
				hash, size, err := statFromStream(reader)
				localStoreDocs[srcStoreKey{Store: store.Name(), Key: id}] = srcStat{
					Hash: hash,
					Size: size,
				}
			}
		}
	}

	slog.Info("sync job collected local store entry list", "entries", len(localStoreDocs), "provider", job.Provider, "job", job.ID)
	return localStoreDocs, nil
}

func collectLocalDriveDocs(walkDir drive.WalkDir, openFile drive.Get, statFile drive.Stat, job Job) (map[drive.FID]srcStat, error) {
	localDriveDocs := map[drive.FID]srcStat{}
	for _, src := range job.Sources {
		if src.Drive.Valid {
			if optStat, err := statFile(user.SU(), src.Drive.Root); err != nil || optStat.IsNone() {
				if err != nil {
					return nil, fmt.Errorf("cannot stat file '%s': %w", src.Drive.Root, err)
				}

				slog.Warn("library sync job contains a drive root FID which is no longer available", "fid", src.Drive.Root, "job", job.ID)
				continue
			}

			err := walkDir(user.SU(), src.Drive.Root, func(fid drive.FID, file drive.File, err error) error {
				if err != nil {
					return err
				}

				if file.Mode().IsRegular() {
					optFile, err := openFile(user.SU(), fid, "")
					if err != nil {
						return fmt.Errorf("cannot get drive file '%s': %w", fid, err)
					}

					if optFile.IsNone() {
						return nil // may be fine, was perhaps deleted concurrently
					}

					f := optFile.Unwrap()
					reader, err := f.Open()
					if err != nil {
						return fmt.Errorf("cannot open drive file '%s': %w", fid, err)
					}

					hash, size, err := statFromStream(reader)
					localDriveDocs[fid] = srcStat{
						Hash: hash,
						Size: size,
					}
				}

				return nil
			})

			if err != nil {
				return nil, fmt.Errorf("cannot walk src root: %w", err)
			}
		}
	}

	return localDriveDocs, nil
}

type srcStoreKey struct {
	Store string
	Key   string
}

type srcStat struct {
	Size int64
	Hash string
}

func statFromStream(reader io.ReadCloser) (hash string, size int64, err error) {
	defer reader.Close()
	hasher := sha3.New256()

	n, err := io.Copy(hasher, reader)
	if err != nil {
		return "", n, err
	}
	return hex.EncodeToString(hasher.Sum(nil)), n, nil
}
