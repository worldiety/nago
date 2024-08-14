package main

import (
	"bytes"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		// EntityStore is only for small files upto a few kilobytes.
		// This store saves everything within a single file using etcd bbolt fork.
		// Transactions are supported.
		//
		// In this example, the database file is ~/.de.worldiety.tutorial/bbolt/bolt.db
		dbstore := cfg.EntityStore("small-blobs")

		// quickly write some bytes using a transaction with a single write.
		if err := blob.Write(dbstore, "my key", bytes.NewBufferString("I'm a fine blob")); err != nil {
			panic(err)
		}

		// read them out
		var buf1 bytes.Buffer
		if err := blob.Read(dbstore, "my key", &buf1); err != nil {
			panic(err)
		}

		// FileStore is for large blobs, from hundreds of kilobytes to gigabytes.
		// It stores each blobs as a single file directly in the local filesystem.
		// There is no transaction support at all, however some tricks like atomic rename are used to
		// lower the risk of damaged files.
		//
		// In this example, the files land in ~/.de.worldiety.tutorial/files/my-large-blobs
		fstore := cfg.FileStore("my-large-blobs")

		// quickly write some bytes
		if err := blob.Write(fstore, "my key", bytes.NewBufferString("I'm a good blob")); err != nil {
			panic(err)
		}

		// read them out
		var buf2 bytes.Buffer
		if err := blob.Read(fstore, "my key", &buf2); err != nil {
			panic(err)
		}

		cfg.Component(".", func(wnd core.Window) core.View {
			return ui.Text(buf1.String() + " & " + buf2.String())
		})
	}).Run()
}
