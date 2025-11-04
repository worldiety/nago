// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"go.wdy.de/nago/pkg/std"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/drive"
	cfgdrive "go.wdy.de/nago/application/drive/cfg"
	uidrive "go.wdy.de/nago/application/drive/ui"
	cfginspector "go.wdy.de/nago/application/inspector/cfg"
	cfglocalization "go.wdy.de/nago/application/localization/cfg"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_76")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		adminID := option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(cfg.NewScaffold().Decorator())
		option.Must(cfginspector.Enable(cfg))
		option.Must(cfglocalization.Enable(cfg))
		drives := option.Must(cfgdrive.Enable(cfg))
		option.Must(cfginspector.Enable(cfg))
		um := std.Must(cfg.UserManagement())
		optAdmin, err := um.UseCases.SubjectFromUser(user.SU(), adminID)
		if err != nil {
			panic(err)
		}
		var admin user.Subject
		if optAdmin.IsSome() {
			admin = optAdmin.Unwrap()
		}
		root := option.Must(drives.UseCases.OpenRoot(admin, drive.OpenRootOptions{
			//User:   "5ead60950c88e64bf04b633f2fb438d4",
			User:   "ce38e2959843819daeaced9dad7151a3",
			Create: true,
			//Name:   "mein Drive",
			Group: "nago.devs",
			Mode:  0740,
		}))

		fmt.Printf("root mode: %#o\n", root.Mode())
		fmt.Printf("owner: %v\n", root.Owner)
		fmt.Printf("group: %v\n", root.Group)

		//root, err := drives.UseCases.OpenRoot(user.SU(), drive.OpenRootOptions{
		//	Name:   "Nago Devs",
		//	Create: true,
		//	Group:  "group.nago.devs", // group ID
		//	Mode:   0000,              // owner: rwx, group: r-x? (adjust as needed); ensure group write bit is set when needed
		//})
		//
		//err = drives.UseCases.
		//
		//if err != nil {
		//	panic(err)
		//}
		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Text("hello world!"),
				uidrive.PageDrive(wnd, drives.UseCases),
			).FullWidth()

		})
	}).Run()
}

//allDrive, err := drives.UseCases.OpenRoot(wnd.Subject(), drive.OpenRootOptions{})
//if err != nil {
//panic(err)
//}
//
//var fileExist bool
//err = drives.UseCases.WalkDir(wnd.Subject(), allDrive.ID, func(fid drive.FID, file drive.File, err error) error {
//	if file.Name() == "MyData" {
//		fileExist = true
//	}
//	return nil
//
//})
//
//if !fileExist {
//
//}
//file := core.MemFile{
//Filename:     "MyData",
//MimeTypeHint: "png",
//Bytes:        Screenshot,
//}
//
//reader, err := file.Open()
//if err != nil {
//panic(err)
//}
//
//err = drives.UseCases.Put(wnd.Subject(), allDrive.ID, "keine Ahnung", reader, drive.PutOptions{
//OriginalFilename: "test",
//SourceHint:       0,
//KeepVersion:      true,
//Mode:             0,
//Owner:            "Peter Pan",
//Group:            "nago.devs",
//})
//if err != nil {
//panic(err)
//}
