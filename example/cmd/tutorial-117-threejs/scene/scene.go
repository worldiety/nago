// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package scene is a program of the gostrict subset, which runs in the browser. It animates a torus knot with
// balls on orbits around it, reports clicks to its backend and takes the color and the speed from it.
package scene

import (
	"math"

	"go.wdy.de/nago/example/cmd/tutorial-117-threejs/scene/proto"
	"go.wdy.de/nago/pkg/gostrict/lib/input"
	"go.wdy.de/nago/pkg/gostrict/lib/three"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/msg"
)

// State is the state of the scene, which the component holds.
type State struct {
	Speed float64
	// Angle is the rotation, which advances by the speed in each frame.
	Angle float64
	Last  float64
	Balls list.List[Ball]
}

type Ball struct {
	Mesh  three.Mesh
	Phase float64
}

var scene = three.NewScene()
var camera = three.NewCamera(60, 1, 0.1, 100)
var material = three.NewMaterial(0x4f8cff, 0.6, 0.25)
var knot = three.NewMesh(three.NewTorusKnot(1, 0.3), material)

// Init builds the scene with n balls and shows it.
func Init(n int32) *State {
	three.SetBackground(scene, 0x101418)
	three.SetPosition(camera, 0, 2.5, 8)
	three.LookAt(camera, 0, 0, 0)
	three.Add(scene, knot)

	key := three.NewDirectionalLight(0xffffff, 2.5)
	three.SetPosition(key, 3, 5, 4)
	three.Add(scene, key)
	three.Add(scene, three.NewAmbientLight(0x8090a0, 0.8))

	palette := list.Of[uint32](0xff6b6b, 0xffd93d, 0x6bcb77, 0x4d96ff, 0xc77dff, 0xff9f45)
	geometry := three.NewSphere(0.25, 24)
	balls := list.New[Ball]()
	for i := range n {
		color, ok := palette.Get(i % 6)
		if !ok {
			color = 0xffffff
		}
		m := three.NewMesh(geometry, three.NewMaterial(color, 0.1, 0.4))
		three.Add(scene, m)
		balls.Append(Ball{Mesh: m, Phase: float64(i) / float64(n) * 2 * math.Pi})
	}

	three.Show(scene, camera)
	return &State{Speed: 1, Balls: balls}
}

// Resize adapts the camera to the aspect ratio of the component.
func Resize(s *State, aspect float64) {
	three.SetAspect(camera, aspect)
}

// Frame moves the scene to the time t in seconds.
func Frame(s *State, t float64) {
	dt := t - s.Last
	s.Last = t
	if dt > 0.1 {
		// e.g. after the tab was hidden
		dt = 0.1
	}
	s.Angle += dt * s.Speed

	a := s.Angle
	three.SetRotation(knot, a*0.4, a*0.7, 0)
	for _, b := range s.Balls.All() {
		x, y, z := orbit(b.Phase, a)
		three.SetPosition(b.Mesh, x, y, z)
		three.SetScale(b.Mesh, 1+0.3*math.Sin(a*3+b.Phase))
	}
}

func orbit(phase, t float64) (float64, float64, float64) {
	a := phase + t*0.8
	r := 3 + 0.4*math.Sin(t*2+phase*3)
	return r * math.Cos(a), math.Sin(a*2) * 0.8, r * math.Sin(a)
}

// OnInput reports a click to the backend.
func OnInput(s *State, e input.Event) msg.Out[proto.Clicked] {
	out := msg.New[proto.Clicked]()
	if e.Type == input.PointerDown {
		out.Send(proto.Clicked{X: e.X, Y: e.Y})
	}
	return out
}

// OnColor takes the color, which the backend sends.
func OnColor(s *State, c proto.Color) {
	three.SetColor(material, c.RGB)
}

// OnSpeed takes the speed, which the backend sends.
func OnSpeed(s *State, m proto.Speed) {
	s.Speed = m.Factor
}
