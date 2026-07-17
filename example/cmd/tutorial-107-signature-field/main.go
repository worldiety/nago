// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	_ "embed"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_107")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			stateDefault := core.StateOf[ui.Signature](wnd, "stateDefault").Init(func() ui.Signature {
				return ui.Signature{
					SVG: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 512 256\" width=\"512\" height=\"256\"><path d=\"M 176.801,63.465 C 173.513,63.668 173.689,64.129 170.578,64.793\" stroke-width=\"5.165\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 170.578,64.793 C 167.519,65.964 167.667,65.909 165.109,67.945\" stroke-width=\"4.847\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 165.109,67.945 C 162.777,70.772 162.321,70.238 160.184,73.340\" stroke-width=\"4.170\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 160.184,73.340 C 155.586,77.715 155.678,77.801 150.910,82.004\" stroke-width=\"3.520\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 150.910,82.004 C 140.056,91.939 140.067,91.951 129.145,101.813\" stroke-width=\"2.203\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 129.145,101.813 C 118.369,110.721 118.808,111.166 108.414,120.457\" stroke-width=\"2.094\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 108.414,120.457 C 101.477,128.181 101.001,127.707 94.410,135.785\" stroke-width=\"2.383\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 94.410,135.785 C 88.604,142.040 88.602,142.035 82.664,148.164\" stroke-width=\"2.665\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 82.664,148.164 C 72.641,154.864 79.229,151.705 75.660,155.113\" stroke-width=\"3.276\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 75.660,155.113 C 92.653,152.020 86.203,155.983 109.789,150.402\" stroke-width=\"3.559\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 109.789,150.402 C 129.310,148.989 129.251,148.565 148.855,148.203\" stroke-width=\"2.020\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 148.855,148.203 C 165.342,147.681 165.332,147.540 181.832,147.504\" stroke-width=\"1.944\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 181.832,147.504 C 197.964,146.866 197.938,147.331 214.047,147.504\" stroke-width=\"1.943\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 214.047,147.504 C 230.888,147.564 230.599,148.161 247.102,150.094\" stroke-width=\"1.863\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 247.102,150.094 C 254.066,151.239 251.839,150.833 255.949,154.043\" stroke-width=\"2.772\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 255.949,154.043 C 255.266,157.300 257.064,155.704 253.098,159.023\" stroke-width=\"3.921\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 253.098,159.023 C 248.335,162.661 248.725,162.499 242.867,164.441\" stroke-width=\"3.500\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 242.867,164.441 C 230.371,168.870 230.544,168.454 217.516,170.609\" stroke-width=\"2.421\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 217.516,170.609 C 204.828,171.616 205.022,172.300 192.168,171.301\" stroke-width=\"2.244\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 192.168,171.301 C 176.383,172.159 176.809,170.919 161.477,169.215\" stroke-width=\"2.069\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 161.477,169.215 C 149.880,166.703 151.516,167.247 142.434,161.477\" stroke-width=\"2.329\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 142.434,161.477 C 141.596,158.453 140.431,160.071 142.578,155.949\" stroke-width=\"3.405\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 142.578,155.949 C 145.174,150.441 144.719,150.494 148.680,145.559\" stroke-width=\"3.325\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 148.680,145.559 C 154.451,136.291 154.920,136.869 162.070,128.805\" stroke-width=\"2.691\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 162.070,128.805 C 176.104,115.127 175.887,115.230 191.551,103.438\" stroke-width=\"1.857\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 191.551,103.438 C 200.014,98.145 199.466,97.448 208.793,93.445\" stroke-width=\"2.316\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 208.793,93.445 C 216.048,89.462 216.050,89.561 223.621,86.270\" stroke-width=\"2.670\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 223.621,86.270 C 230.566,80.793 228.119,84.374 232.934,83.270\" stroke-width=\"3.242\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 232.934,83.270 C 234.447,91.597 237.019,87.543 236.527,99.770\" stroke-width=\"3.563\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 236.527,99.770 C 237.380,103.807 237.529,103.703 239.098,107.480\" stroke-width=\"3.729\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 239.098,107.480 C 241.786,121.647 244.454,118.934 250.676,130.023\" stroke-width=\"2.561\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 250.676,130.023 C 266.879,143.505 264.751,143.287 285.027,150.762\" stroke-width=\"1.868\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 285.027,150.762 C 295.418,152.231 294.436,154.524 305.789,152.063\" stroke-width=\"2.238\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 305.789,152.063 C 317.400,151.661 317.326,151.875 328.844,150.051\" stroke-width=\"2.337\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 328.844,150.051 C 341.744,148.550 341.599,148.256 354.188,145.254\" stroke-width=\"2.319\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 354.188,145.254 C 371.264,140.151 371.359,140.705 388.074,134.359\" stroke-width=\"1.885\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 388.074,134.359 C 391.810,133.120 391.741,133.096 395.141,131.145\" stroke-width=\"2.907\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path><path d=\"M 395.141,131.145 C 403.520,126.188 403.447,126.218 411.348,120.555\" stroke-width=\"2.707\" stroke=\"black\" fill=\"none\" stroke-linecap=\"round\"></path></svg>",
				}
			})

			return ui.VStack(
				ui.ThemeSwitcher(
					ui.PrimaryButton(nil).Title("Toggle theme"),
				),
				ui.SignatureField("Unterschrift", stateDefault),
				ui.Image().Embed([]byte(stateDefault.Get().SVG)).Frame(ui.Frame{Width: ui.L400, Height: ui.L200, MaxWidth: ui.Full}),
			).Gap(ui.L32).Padding(ui.Padding{}.All(ui.L16)).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}
