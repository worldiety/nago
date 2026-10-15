// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package spin animates a torus knot with balls on orbits around it in three.js.
package spin

import (
	"math"

	"go.wdy.de/nago/pkg/gostrict/lib/three"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

type Ball struct {
	Mesh  three.Mesh
	Phase float64
}

var scene = three.NewScene()
var camera = three.NewCamera(60, 1, 0.1, 100)
var knot = three.NewMesh(three.NewTorusKnot(1, 0.3), three.NewMaterial(0x4f8cff, 0.6, 0.25))
var balls = list.Freeze(makeBalls(12))

func makeBalls(n int32) list.List[Ball] {
	palette := list.Of[uint32](0xff6b6b, 0xffd93d, 0x6bcb77, 0x4d96ff, 0xc77dff, 0xff9f45)
	geometry := three.NewSphere(0.25, 24)
	l := list.New[Ball]()
	for i := range n {
		color, ok := palette.Get(i % 6)
		if !ok {
			color = 0xffffff
		}
		m := three.NewMesh(geometry, three.NewMaterial(color, 0.1, 0.4))
		three.Add(scene, m)
		l.Append(Ball{Mesh: m, Phase: float64(i) / float64(n) * 2 * math.Pi})
	}
	return l
}

// Setup shows the scene for the aspect ratio of the canvas and returns the number of meshes.
func Setup(aspect float64) int32 {
	three.SetBackground(scene, 0x101418)
	three.SetAspect(camera, aspect)
	three.SetPosition(camera, 0, 2.5, 8)
	three.LookAt(camera, 0, 0, 0)
	three.Add(scene, knot)

	key := three.NewDirectionalLight(0xffffff, 2.5)
	three.SetPosition(key, 3, 5, 4)
	three.Add(scene, key)
	three.Add(scene, three.NewAmbientLight(0x8090a0, 0.8))

	three.Show(scene, camera)
	return balls.Len() + 1
}

// Frame moves the scene to the time t in seconds.
func Frame(t float64) {
	three.SetRotation(knot, t*0.4, t*0.7, 0)
	for _, b := range balls.All() {
		x, y, z := orbit(b.Phase, t)
		three.SetPosition(b.Mesh, x, y, z)
		three.SetScale(b.Mesh, 1+0.3*math.Sin(t*3+b.Phase))
	}
}

func orbit(phase, t float64) (float64, float64, float64) {
	a := phase + t*0.8
	r := 3 + 0.4*math.Sin(t*2+phase*3)
	return r * math.Cos(a), math.Sin(a*2) * 0.8, r * math.Sin(a)
}

// Orbit returns the position of a ball at the time t.
func Orbit(phase, t float64) (float64, float64, float64) {
	return orbit(phase, t)
}
